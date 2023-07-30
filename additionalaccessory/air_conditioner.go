package additionalaccessory

import (
	"net/http"

	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
	"github.com/legnoh/hap-nature-remo/additionalcharacteristic"
	"github.com/legnoh/hap-nature-remo/util"
	"github.com/sirupsen/logrus"
	"github.com/tenntenn/natureremo"
)

type AirConditioner struct {
	*accessory.A
	HeaterCooler *service.HeaterCooler
}

func NewAirConditioner(nr *natureremo.Client, ac *natureremo.Appliance, devices []*natureremo.Device) *AirConditioner {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	acceInfo := accessory.Info{
		Name:         ac.Nickname,
		Manufacturer: ac.Model.Manufacturer,
		Model:        ac.Model.RemoteName,
		SerialNumber: ac.ID,
	}

	a := AirConditioner{
		A:            accessory.New(acceInfo, accessory.TypeAirConditioner),
		HeaterCooler: service.NewHeaterCooler(),
	}

	a.HeaterCooler.TargetHeaterCoolerState.ValidVals = []int{}
	a.HeaterCooler.CurrentHeaterCoolerState.ValidVals = []int{
		characteristic.CurrentHeaterCoolerStateInactive,
		characteristic.CurrentHeaterCoolerStateIdle,
	}

	// 現在の動作モードを呼び出された時の処理
	a.HeaterCooler.TargetHeaterCoolerState.ValueRequestFunc = func(*http.Request) (interface{}, int) {
		log.Debug("Get now AirConditioner Mode Request")
		aps := util.GetAppliances(nr)
		for _, ap := range aps.Appliances {
			if ap.ID == ac.ID {
				switch ac.AirConSettings.OperationMode {
				case natureremo.OperationModeCool:
					return characteristic.TargetHeaterCoolerStateCool, 0
				case natureremo.OperationModeWarm:
					return characteristic.TargetHeaterCoolerStateHeat, 0
				}
			}
		}
		return nil, -1
	}
	a.HeaterCooler.CurrentHeaterCoolerState.ValueRequestFunc = func(*http.Request) (interface{}, int) {
		log.Debug("Get now AirConditioner Mode Request")
		aps := util.GetAppliances(nr)
		for _, ap := range aps.Appliances {
			if ap.ID == ac.ID {
				switch ac.AirConSettings.OperationMode {
				case natureremo.OperationModeCool:
					return characteristic.CurrentHeaterCoolerStateCooling, 0
				case natureremo.OperationModeWarm:
					return characteristic.CurrentHeaterCoolerStateHeating, 0
				case natureremo.OperationModeAuto:
					return characteristic.CurrentHeaterCoolerStateIdle, 0
				case natureremo.OperationModeBlow:
					return characteristic.CurrentHeaterCoolerStateIdle, 0
				case natureremo.OperationModeDry:
					return characteristic.CurrentHeaterCoolerStateCooling, 0
				}
			}
		}
		return nil, -1
	}

	// 動作モードが変わった時の処理
	a.HeaterCooler.TargetHeaterCoolerState.OnValueRemoteUpdate(func(target int) {
		log.Infof("AirConditioner TargetHeaterCoolerState Changed: %d", target)
		req := natureremo.AirConSettings{}
		switch target {
		case characteristic.TargetHeaterCoolerStateCool:
			req.OperationMode = natureremo.OperationModeCool
		case characteristic.TargetHeaterCoolerStateHeat:
			req.OperationMode = natureremo.OperationModeWarm
		case characteristic.TargetHeatingCoolingStateOff:
			req.Button = natureremo.ButtonPowerOff
		}
		if err := util.SendAirconRequest(nr, ac, &req); err != nil {
			log.Error(err)
		}
	})

	a.HeaterCooler.Active.OnValueRemoteUpdate(func(target int) {
		log.Infof("AirConditioner Active Changed: %d", target)
		req := natureremo.AirConSettings{}
		if target == characteristic.ActiveInactive {
			req.Button = natureremo.ButtonPowerOff
		}
		if err := util.SendAirconRequest(nr, ac, &req); err != nil {
			log.Error(err)
		}
	})

	// 動作モードの初期化処理
	targetState := a.HeaterCooler.TargetHeaterCoolerState.ValidVals
	currentState := a.HeaterCooler.CurrentHeaterCoolerState.ValidVals

	// 冷房/暖房があればそれぞれ動作選択肢に登録
	if cooler, coolerFound := ac.AirCon.Range.Modes[natureremo.OperationModeCool]; coolerFound {
		log.Infof("Cooler detected: %s", ac.Nickname)
		targetState = append(targetState, characteristic.TargetHeaterCoolerStateCool)
		currentState = append(currentState, characteristic.CurrentHeaterCoolerStateCooling)
		threshold := *additionalcharacteristic.NewCoolingThresholdTemperature(cooler, nr, ac)
		a.HeaterCooler.AddC(threshold.C)
	}
	if heater, heaterFound := ac.AirCon.Range.Modes[natureremo.OperationModeWarm]; heaterFound {
		log.Infof("Heater detected: %s", ac.Nickname)
		targetState = append(targetState, characteristic.TargetHeaterCoolerStateHeat)
		currentState = append(currentState, characteristic.CurrentHeaterCoolerStateHeating)
		threshold := *additionalcharacteristic.NewHeatingThresholdTemperature(heater, nr, ac)
		a.HeaterCooler.AddC(threshold.C)
	}

	a.HeaterCooler.TargetHeaterCoolerState.ValidVals = targetState
	a.HeaterCooler.CurrentHeaterCoolerState.ValidVals = currentState

	// 現在の動作状況確認を初期状態で入れる処理(室温)
	var temp float64
	for _, device := range devices {
		if val, found := device.NewestEvents[natureremo.SensorTypeTemperature]; found {
			if device.ID == ac.Device.ID {
				a.HeaterCooler.CurrentTemperature.SetValue(val.Value)
			}
		}
	}

	// Natureデバイスが温度計を持っていないものだった場合、別の端末で計測したものがあったら代わりに使う
	if temp == 0 {
		for _, device := range devices {
			if val, found := device.NewestEvents[natureremo.SensorTypeTemperature]; found {
				log.Warnf("%s don't have temperature sensor. Using %s sensor instead for %s", ac.Device.Name, device.Name, ac.Nickname)
				a.HeaterCooler.CurrentTemperature.SetValue(val.Value)
			}
		}
	}

	// 現在の動作状況確認を初期状態で入れる処理(モード)
	switch ac.AirConSettings.OperationMode {
	case natureremo.OperationModeCool:
		a.HeaterCooler.Active.SetValue(characteristic.ActiveActive)
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateCooling)
		a.HeaterCooler.TargetHeaterCoolerState.SetValue(characteristic.TargetHeaterCoolerStateCool)
	case natureremo.OperationModeWarm:
		a.HeaterCooler.Active.SetValue(characteristic.ActiveActive)
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateHeating)
		a.HeaterCooler.TargetHeaterCoolerState.SetValue(characteristic.TargetHeaterCoolerStateHeat)
	default:
		a.HeaterCooler.Active.SetValue(characteristic.ActiveActive)
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateIdle)
		a.HeaterCooler.TargetHeaterCoolerState.SetValue(targetState[0])
	}
	if ac.AirConSettings.Button == natureremo.ButtonPowerOff {
		a.HeaterCooler.Active.SetValue(characteristic.ActiveInactive)
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateInactive)
	}

	// 現在気温の確認処理
	a.HeaterCooler.CurrentTemperature.ValueRequestFunc = func(*http.Request) (interface{}, int) {
		var temp float64
		devices := util.GetDevices(nr)
		for _, device := range devices.Devices {
			if val, found := device.NewestEvents[natureremo.SensorTypeTemperature]; found {
				if device.ID == ac.Device.ID {
					temp = val.Value
					log.Infof("%s: Get now AirCon Temperature Request Successful: %.1f", ac.Nickname, temp)
					return temp, 0
				}
			}
		}

		if temp == 0 {
			for _, device := range devices.Devices {
				if val, found := device.NewestEvents[natureremo.SensorTypeTemperature]; found {
					temp = val.Value
					log.Infof("%s: Get now AirCon Temperature Request Successful(%s): %.1f", ac.Nickname, device.Name, temp)
					return temp, 0
				}
			}
		}
		log.Warnf("%s: Get now AirCon Temperature Request devices was not found(%s)", ac.Nickname, ac.Device.Name)
		return nil, -1
	}
	a.AddS(a.HeaterCooler.S)
	return &a
}
