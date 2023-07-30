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
}

func NewSensor(nr *natureremo.Client, device *natureremo.Device) Sensor {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	acceInfo := accessory.Info{
		Name:         device.DeviceCore.Name,
		Manufacturer: "Nature Inc.",
		Model:        device.FirmwareVersion,
		SerialNumber: device.SerialNumber,
	}

	a := Sensor{
		A: accessory.New(acceInfo, accessory.TypeSensor),
	}

	if te, found := device.NewestEvents[natureremo.SensorTypeTemperature]; found {
		log.Infof("Temperature Sensor Detected(%s): %.1f", device.Name, te.Value)
		temperatureSensor := service.NewTemperatureSensor()
		temperatureSensor.CurrentTemperature.SetValue(te.Value)

		temperatureSensor.CurrentTemperature.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debugf("%s: Get now Temperature Request", device.Name)
			devices := util.GetDevices(nr)
			for _, remoteDevice := range devices.Devices {
				if remoteDevice.Name == device.Name {
					temp := remoteDevice.NewestEvents[natureremo.SensorTypeTemperature].Value
					log.Infof("%s: Get now Temperature Request Successful: %.1f", device.Name, temp)
					return temp, 0
				}
			}
			log.Warnf("%s: Get now Temperature Request devices was not found", device.Name)
			return nil, -1
		}
		a.AddS(temperatureSensor.S)
	}

	if hu, found := device.NewestEvents[natureremo.SensorTypeHumidity]; found {
		log.Infof("Humidity Sensor Detected(%s): %.0f", device.Name, hu.Value)
		humiditySensor := service.NewHumiditySensor()
		humiditySensor.CurrentRelativeHumidity.SetValue(hu.Value)

		humiditySensor.CurrentRelativeHumidity.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debugf("%s: Get now Humidity Request", device.Name)
			devices := util.GetDevices(nr)
			for _, remoteDevice := range devices.Devices {
				if remoteDevice.Name == device.Name {
					humi := remoteDevice.NewestEvents[natureremo.SensorTypeHumidity].Value
					log.Infof("%s: Get now Humidity Request Successful: %.0f", device.Name, humi)
					return humi, 0
				}
			}
			log.Warnf("%s: Get now Humidity Request devices was not found", device.Name)
			return nil, -1
		}
		a.AddS(humiditySensor.S)
	}

	if il, found := device.NewestEvents[natureremo.SensorTypeIllumination]; found {
		log.Infof("Illumination Sensor Detected(%s): %.0f", device.Name, il.Value)
		lightSensor := service.NewLightSensor()
		lightSensor.CurrentAmbientLightLevel.SetMinValue(0)
		lightSensor.CurrentAmbientLightLevel.SetMaxValue(200)
		lightSensor.CurrentAmbientLightLevel.SetStepValue(1)
		lightSensor.CurrentAmbientLightLevel.SetValue(il.Value)

		lightSensor.CurrentAmbientLightLevel.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debugf("%s: Get now LightLevel Request", device.Name)
			devices := util.GetDevices(nr)
			for _, remoteDevice := range devices.Devices {
				if remoteDevice.DeviceCore.Name == device.DeviceCore.Name {
					illu := remoteDevice.NewestEvents[natureremo.SensorTypeIllumination].Value
					log.Infof("%s: Get now Lightlevel Request Successful: %.0f", device.Name, illu)
					return illu, 0
				}
			}
			log.Warnf("%s: Get now Illuminate Request devices was not found", device.Name)
			return nil, -1
		}
		a.AddS(lightSensor.S)
	}

	if mo, found := device.NewestEvents[natureremo.SensorTypeMovement]; found {
		log.Infof("Movement Sensor Detected(%s): %.1f", device.Name, mo.Value)
		motionSensor := service.NewMotionSensor()
		if mo.Value == 0 {
			motionSensor.MotionDetected.SetValue(false)
		} else {
			// モーションセンサーは基本的に常時ONで返される仕様らしいので、
			// 更新から5分以内の場合のみ検知したものとして扱う
			now := time.Now()
			if now.Sub(mo.CreatedAt).Minutes() > 5 {
				motionSensor.MotionDetected.SetValue(false)
			} else {
				motionSensor.MotionDetected.SetValue(true)
			}
		}

		motionSensor.MotionDetected.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debugf("%s: Get now MotionSensor Request", device.Name)
			devices := util.GetDevices(nr)
			for _, remoteDevice := range devices.Devices {
				if remoteDevice.Name == device.Name {
					state := remoteDevice.NewestEvents[natureremo.SensorTypeTemperature]
					if state.Value == 0 {
						log.Infof("%s: Get now Motion Request Successful: %t", device.Name, false)
						return false, 0
					} else {
						// モーションセンサーは基本的に常時ONで返される仕様らしいので、
						// 更新から5分以内の場合のみ検知したものとして扱う
						now := time.Now()
						if now.Sub(state.CreatedAt).Minutes() > 5 {
							log.Infof("%s: Get now Motion Request Successful: %t", device.Name, false)
							return false, 0
						} else {
							log.Infof("%s: Get now Motion Request Successful: %t", device.Name, true)
							return true, 0
						}
					}
				}
			}
			log.Warnf("%s: Get now Movement Request devices was not found", device.Name)
			return nil, -1
		}
		a.AddS(motionSensor.S)
	}

	return a
}
