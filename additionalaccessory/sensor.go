package additionalaccessory

import (
	"net/http"
	"time"

	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/service"
	"github.com/legnoh/hap-nature-remo/util"
	"github.com/sirupsen/logrus"
	"github.com/tenntenn/natureremo"
)

type Sensor struct {
	*accessory.A
	TemperatureSensor *service.TemperatureSensor
	HumiditySensor    *service.HumiditySensor
	LightSensor       *service.LightSensor
	MotionSensor      *service.MotionSensor
}

func NewSensor(nr *natureremo.Client, device natureremo.Device) *Sensor {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	acceInfo := accessory.Info{
		Name:         device.DeviceCore.Name,
		Manufacturer: "Nature Inc.",
		Model:        device.FirmwareVersion,
		SerialNumber: device.SerialNumber,
	}

	a := Sensor{}
	a.A = accessory.New(acceInfo, accessory.TypeSensor)

	if te, found := device.NewestEvents[natureremo.SensorTypeTemperature]; found {
		log.Infof("Temperature Sensor Detected(%s): %.1f", device.Name, te.Value)
		a.TemperatureSensor = service.NewTemperatureSensor()
		a.TemperatureSensor.CurrentTemperature.SetValue(te.Value)
		a.AddS(a.TemperatureSensor.S)

		a.TemperatureSensor.CurrentTemperature.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debug("Get now Temperature Request")
			aps := util.GetAppliances(nr)
			devices := util.GetDevices(nr)
			for _, appliance := range aps.Appliances {
				if appliance.Device.Name == a.A.Info.Name.Val {
					for _, device := range devices.Devices {
						if appliance.Device.Name == device.DeviceCore.Name {
							temp := device.NewestEvents[natureremo.SensorTypeTemperature].Value
							log.Infof("%s: Get now Temperature Request Successful: %.1f", a.A.Info.Name.Val, temp)
							return temp, 0
						}
					}
				}
			}
			log.Warnf("%s: Get now Temperature Request devices was not found", a.A.Info.Name.Val)
			return nil, -1
		}
	}

	if hu, found := device.NewestEvents[natureremo.SensorTypeHumidity]; found {
		log.Infof("Humidity Sensor Detected(%s): %.0f", device.Name, hu.Value)
		a.HumiditySensor = service.NewHumiditySensor()
		a.HumiditySensor.CurrentRelativeHumidity.SetValue(hu.Value)
		a.AddS(a.HumiditySensor.S)

		a.HumiditySensor.CurrentRelativeHumidity.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debug("Get now Temperature Request")
			aps := util.GetAppliances(nr)
			devices := util.GetDevices(nr)
			for _, appliance := range aps.Appliances {
				if appliance.Device.Name == a.A.Info.Name.Val {
					for _, device := range devices.Devices {
						if appliance.Device.Name == device.DeviceCore.Name {
							humi := device.NewestEvents[natureremo.SensorTypeHumidity].Value
							log.Infof("%s: Get now Humidity Request Successful: %.0f", a.A.Info.Name.Val, humi)
							return humi, 0
						}
					}
				}
			}
			log.Warnf("%s: Get now Humidity Request devices was not found", a.A.Info.Name.Val)
			return nil, -1
		}
	}

	if il, found := device.NewestEvents[natureremo.SensorTypeIllumination]; found {
		log.Infof("Illumination Sensor Detected(%s): %.0f", device.Name, il.Value)
		a.LightSensor = service.NewLightSensor()
		a.LightSensor.CurrentAmbientLightLevel.SetMinValue(0)
		a.LightSensor.CurrentAmbientLightLevel.SetMaxValue(200)
		a.LightSensor.CurrentAmbientLightLevel.SetStepValue(1)
		a.LightSensor.CurrentAmbientLightLevel.SetValue(il.Value)
		a.AddS(a.LightSensor.S)

		a.LightSensor.CurrentAmbientLightLevel.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debug("Get now Temperature Request")
			aps := util.GetAppliances(nr)
			devices := util.GetDevices(nr)
			for _, appliance := range aps.Appliances {
				if appliance.Device.Name == a.A.Info.Name.Val {
					for _, device := range devices.Devices {
						if appliance.Device.Name == device.DeviceCore.Name {
							illu := device.NewestEvents[natureremo.SensorTypeIllumination].Value
							log.Infof("%s: Get now Lightlevel Request Successful: %.0f", a.A.Info.Name.Val, illu)
							return illu, 0
						}
					}
				}
			}
			log.Warnf("%s: Get now Illuminate Request devices was not found", a.A.Info.Name.Val)
			return nil, -1
		}
	}

	if mo, found := device.NewestEvents[natureremo.SensorTypeMovement]; found {
		log.Infof("Movement Sensor Detected(%s): %.1f", device.Name, mo.Value)
		a.MotionSensor = service.NewMotionSensor()
		if mo.Value == 0 {
			a.MotionSensor.MotionDetected.SetValue(false)
		} else {
			// モーションセンサーは基本的に常時ONで返される仕様らしいので、
			// 更新から5分以内の場合のみ検知したものとして扱う
			now := time.Now()
			if now.Sub(mo.CreatedAt).Minutes() > 5 {
				a.MotionSensor.MotionDetected.SetValue(false)
			} else {
				a.MotionSensor.MotionDetected.SetValue(true)
			}
		}
		a.AddS(a.MotionSensor.S)

		a.MotionSensor.MotionDetected.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debug("Get now Temperature Request")
			aps := util.GetAppliances(nr)
			devices := util.GetDevices(nr)
			for _, appliance := range aps.Appliances {
				if appliance.Device.Name == a.A.Info.Name.Val {
					for _, device := range devices.Devices {
						if appliance.Device.Name == device.DeviceCore.Name {
							state := device.NewestEvents[natureremo.SensorTypeTemperature]
							if state.Value == 0 {
								log.Infof("%s: Get now Motion Request Successful: %t", a.A.Info.Name.Val, false)
								return false, 0
							} else {
								// モーションセンサーは基本的に常時ONで返される仕様らしいので、
								// 更新から5分以内の場合のみ検知したものとして扱う
								now := time.Now()
								if now.Sub(state.CreatedAt).Minutes() > 5 {
									log.Infof("%s: Get now Motion Request Successful: %t", a.A.Info.Name.Val, false)
									return false, 0
								} else {
									log.Infof("%s: Get now Motion Request Successful: %t", a.A.Info.Name.Val, true)
									return true, 0
								}
							}
						}
					}
				}
			}
			log.Warnf("%s: Get now Movement Request devices was not found", a.A.Info.Name.Val)
			return nil, -1
		}
	}

	return &a
}
