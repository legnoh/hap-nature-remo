package util

import (
	"sort"
	"strconv"

	"github.com/brutella/hap/accessory"
	"github.com/tenntenn/natureremo"
)

func GetStepInfo(values []string) (float64, float64, float64) {

	var steps []float64

	for _, v := range values {
		val, _ := strconv.ParseFloat(v, 64)
		steps = append(steps, val)
	}
	sort.Float64s(sort.Float64Slice(steps))

	min := steps[0]
	max := steps[len(steps)-1]
	step := steps[1] - steps[0]

	return min, max, step
}

func SetBridgeFirmwareInfo(bridgeMeta accessory.Info, nd natureremo.DeviceCore) accessory.Info {

	if bridgeMeta.Firmware == "" {
		bridgeMeta.Model = nd.Name
		bridgeMeta.Manufacturer = "Nature Inc."
		bridgeMeta.Firmware = nd.FirmwareVersion
		bridgeMeta.SerialNumber = nd.SerialNumber
	}
	return bridgeMeta
}
