package util

import (
	"context"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tenntenn/natureremo"
)

type NrDevices struct {
	Devices   []*natureremo.Device
	UpdatedAt time.Time
}

type NrAppliances struct {
	Appliances []*natureremo.Appliance
	UpdatedAt  time.Time
}

var (
	nrDevices    NrDevices
	nrAppliances NrAppliances
)

// NatureRemoの Appliance取得リクエストを行う関数
// (大量のリクエストが走ることを防ぐため、10秒未満のリクエストの場合は前回のリクエスト結果を使う)
func GetAppliances(nr *natureremo.Client) NrAppliances {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	nrctx := context.Background()

	now := time.Now()
	delta := now.Sub(nrAppliances.UpdatedAt).Seconds()
	log.Debugf("delta: %2f(%t)", delta, delta > 10)

	if delta > 10 {
		if aps, err := nr.ApplianceService.GetAll(nrctx); err != nil {
			log.Error(err)
			return nrAppliances
		} else {
			log.Info("Get Latest Appliances Successful.")
			nrAppliances = NrAppliances{
				Appliances: aps,
				UpdatedAt:  time.Now(),
			}
			return nrAppliances
		}
	}
	log.Debugf("Using cached Appliances responses: %s", nrAppliances.UpdatedAt)
	return nrAppliances
}

// NatureRemoの Device 取得リクエストを行う関数
// (大量のリクエストが走ることを防ぐため、10秒未満のリクエストの場合は前回のリクエスト結果を使う)
func GetDevices(nr *natureremo.Client) NrDevices {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	nrctx := context.Background()
	now := time.Now()

	if now.Sub(nrDevices.UpdatedAt).Seconds() > 10 {
		if dvs, err := nr.DeviceService.GetAll(nrctx); err != nil {
			log.Error(err)
			return nrDevices
		} else {
			log.Info("Get Latest Devices Successful.")
			nrDevices = NrDevices{
				Devices:   dvs,
				UpdatedAt: time.Now(),
			}
			return nrDevices
		}
	}
	log.Debugf("Using cached Devices responses: %s", nrDevices.UpdatedAt)
	return nrDevices
}

// エアコンのモード変更リクエストを行う関数
func SendAirconRequest(nr *natureremo.Client, ac *natureremo.Appliance, mode *natureremo.AirConSettings) error {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	nrctx := context.Background()

	// (リクエストを散らすため、5秒以内でランダム秒待つ処理を加える)
	wait := rand.Intn(5)
	log.Debugf("SendAirconRequest: Sleeping %d seconds...\n", wait)
	time.Sleep(time.Duration(wait) * time.Second)

	return nr.ApplianceService.UpdateAirConSettings(nrctx, ac, mode)
}

func SendSignalRequest(nr *natureremo.Client, signal *natureremo.Signal) error {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	nrctx := context.Background()

	// (リクエストを散らすため、5秒以内でランダム秒待つ処理を加える)
	wait := rand.Intn(5)
	log.Debugf("SendSignalRequest: Sleeping %d seconds...\n", wait)
	time.Sleep(time.Duration(wait) * time.Second)

	return nr.SignalService.Send(nrctx, signal)
}
