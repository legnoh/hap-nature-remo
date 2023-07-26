package additionalaccessory

import (
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
)

type AirConditioner struct {
	*accessory.A
	HeaterCooler *service.HeaterCooler
}

// NewCooler returns a cooler accessory.
func NewAirConditioner(info accessory.Info) *AirConditioner {
	a := AirConditioner{}
	a.A = accessory.New(info, accessory.TypeAirConditioner)

	a.HeaterCooler = service.NewHeaterCooler()

	a.HeaterCooler.TargetHeaterCoolerState.ValidVals = []int{}
	a.HeaterCooler.CurrentHeaterCoolerState.ValidVals = []int{
		characteristic.CurrentHeaterCoolerStateInactive,
		characteristic.CurrentHeaterCoolerStateIdle,
	}

	a.AddS(a.HeaterCooler.S)
	return &a
}
