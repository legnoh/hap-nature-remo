package additionalaccessory

import (
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/service"
	"github.com/tenntenn/natureremo"
)

type Sensor struct {
	*accessory.A
	TemperatureSensor *service.TemperatureSensor
	HumiditySensor    *service.HumiditySensor
	LightSensor       *service.LightSensor
	MotionSensor      *service.MotionSensor
}

func NewSensor(info accessory.Info, sensors map[natureremo.SensorType]float64) *Sensor {
	a := Sensor{}
	a.A = accessory.New(info, accessory.TypeSensor)

	if te, found := sensors[natureremo.SensorTypeTemperature]; found {
		a.TemperatureSensor = service.NewTemperatureSensor()
		a.TemperatureSensor.CurrentTemperature.SetValue(te)
		a.AddS(a.TemperatureSensor.S)
	}

	if hu, found := sensors[natureremo.SensorTypeHumidity]; found {
		a.HumiditySensor = service.NewHumiditySensor()
		a.HumiditySensor.CurrentRelativeHumidity.SetValue(hu)
		a.AddS(a.HumiditySensor.S)
	}

	if il, found := sensors[natureremo.SensorTypeIllumination]; found {
		a.LightSensor = service.NewLightSensor()
		a.LightSensor.CurrentAmbientLightLevel.SetMinValue(0)
		a.LightSensor.CurrentAmbientLightLevel.SetMaxValue(200)
		a.LightSensor.CurrentAmbientLightLevel.SetStepValue(1)
		a.LightSensor.CurrentAmbientLightLevel.SetValue(il)
		a.AddS(a.LightSensor.S)
	}

	if mo, found := sensors[natureremo.SensorTypeMovement]; found {
		a.MotionSensor = service.NewMotionSensor()
		switch mo {
		case 0:
			a.MotionSensor.MotionDetected.SetValue(false)
		case 1:
			a.MotionSensor.MotionDetected.SetValue(true)
		}
		a.AddS(a.MotionSensor.S)
	}

	return &a
}
