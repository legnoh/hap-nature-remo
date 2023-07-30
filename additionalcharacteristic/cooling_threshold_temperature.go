package additionalcharacteristic

import (
	"net/http"
	"strconv"

	"github.com/brutella/hap/characteristic"
	"github.com/legnoh/hap-nature-remo/util"
	"github.com/sirupsen/logrus"
	"github.com/tenntenn/natureremo"
)

func NewCoolingThresholdTemperature(f *natureremo.AirConRangeMode, nr *natureremo.Client, ac *natureremo.Appliance) *characteristic.CoolingThresholdTemperature {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	min, max, step := util.GetStepInfo(f.Temperature)
	log.Debugf("Cooling range: %2f ~ %2f", min, max)

	threshold := *characteristic.NewCoolingThresholdTemperature()
	threshold.SetMinValue(min)
	threshold.SetMaxValue(max)
	threshold.SetStepValue(step)
	nowSetting, _ := strconv.ParseFloat(ac.AirConSettings.Temperature, 64)
	threshold.SetValue(nowSetting)

	// 設定温度が変わった時の処理
	threshold.OnValueRemoteUpdate(func(v float64) {
		target := strconv.FormatFloat(v, 'f', -1, 64)
		setting := natureremo.AirConSettings{
			Temperature: target,
		}
		log.Infof("AirConditioner(Cooler) Temperature Updating: %s", target)
		err := util.SendAirconRequest(nr, ac, &setting)
		if err != nil {
			log.Error(err)
		}
	})

	// 現在の設定値を呼び出された時の処理
	threshold.ValueRequestFunc = func(*http.Request) (interface{}, int) {
		log.Debug("Get now AirConditioner threshold Request")
		aps := util.GetAppliances(nr)
		for _, ap := range aps.Appliances {
			if ap.ID == ac.ID {
				temp, _ := strconv.ParseFloat(ac.AirConSettings.Temperature, 64)
				return temp, 0
			}
		}
		return nil, -1
	}
	return &threshold
}
