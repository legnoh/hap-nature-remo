package additionalaccessory

import (
	"context"
	"regexp"
	"strconv"

	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
	"github.com/legnoh/hap-nature-remo/util"
	"github.com/sirupsen/logrus"
	"github.com/tenntenn/natureremo"
)

type Fan struct {
	*accessory.A
	Fan *service.Fan
}

func NewFan(nr *natureremo.Client, appliance *natureremo.Appliance) Fan {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	nrctx := context.Background()
	speedRe := regexp.MustCompile(`^ico_number_(\d)$`)
	directionRe := regexp.MustCompile(`^ico_(.*)ward$`)

	acceInfo := accessory.Info{
		Name: appliance.Nickname,
	}

	a := Fan{
		A:   accessory.New(acceInfo, accessory.TypeFan),
		Fan: service.NewFan(),
	}

	signals, err := nr.SignalService.GetAll(nrctx, appliance)
	if err != nil {
		log.Fatalf("can't get enough singnals: %s", err)
	}

	rotationSpeedSignals := make(map[int]*natureremo.Signal)
	rotationDirectionSignals := make(map[string]*natureremo.Signal)
	maxLevel := 0

	// 全てのシグナル情報からHomeKitで操作可能なものを抽出
	for _, signal := range signals {

		// 数字アイコン(風量)
		numberPattern := speedRe.FindSubmatch([]byte(signal.Image))
		if len(numberPattern) == 2 {
			level, _ := strconv.Atoi(string(numberPattern[1]))
			if maxLevel < level {
				maxLevel = level
			}
			log.Debugf("Signal Level%d: %s", level, signal.ID)
			rotationSpeedSignals[level] = signal
		}

		// 方向アイコン(風向き)
		directionPattern := directionRe.FindSubmatch([]byte(signal.Image))
		if len(directionPattern) == 2 {
			direction := string(directionPattern[1])
			log.Debugf("Signal Direction(%s): %s", direction, signal.ID)
			rotationDirectionSignals[direction] = signal
		}
	}
	if maxLevel == 0 {
		log.Fatal("RotationSpeed Signal not found")
	}

	// 風量のcharacteristicとリモート動作を設定
	minStep := 100 / maxLevel
	speed := characteristic.NewRotationSpeed()
	speed.SetStepValue(float64(minStep))
	speed.OnValueRemoteUpdate(func(v float64) {
		log.Infof("speed changed: %d", int(v))
		targetLevel := int(v) / minStep
		if rotationSpeedSignals[targetLevel] == nil {
			log.Errorf("target level(%d) signal is not defined", targetLevel)
		} else {
			targetSignal := rotationSpeedSignals[targetLevel]
			if err := util.SendSignalRequest(nr, targetSignal); err != nil {
				log.Error(err)
			} else {
				log.Debugf("Send signal Successful: %d", targetLevel)
			}
		}
	})
	a.Fan.AddC(speed.C)

	// 風向きアイコンがあった場合のcharacteristicとリモート動作を設定
	if rotationSigCount := len(rotationDirectionSignals); rotationSigCount == 0 {
		log.Debug("RotationDirection Signal not found")
	} else {
		log.Debugf("RotationDirection Signal Count: %d", rotationSigCount)
		f, fFound := rotationDirectionSignals["for"]
		b, bFound := rotationDirectionSignals["back"]
		direction := characteristic.NewRotationDirection()

		// 両方向あった時はそれぞれに適した信号に、片方しかない時は回転のたびに同じ信号にする
		if fFound && bFound {
			direction.OnValueRemoteUpdate(func(v int) {
				log.Infof("rotation changed: %d", v)
				if v == characteristic.RotationDirectionClockwise {
					if err := util.SendSignalRequest(nr, f); err != nil {
						log.Error(err)
					}
				} else if v == characteristic.RotationDirectionCounterclockwise {
					if err := util.SendSignalRequest(nr, b); err != nil {
						log.Error(err)
					}
				}
			})
		} else if fFound {
			direction.OnValueRemoteUpdate(func(v int) {
				if err := util.SendSignalRequest(nr, f); err != nil {
					log.Error(err)
				}
			})
		} else if bFound {
			direction.OnValueRemoteUpdate(func(v int) {
				if err := util.SendSignalRequest(nr, b); err != nil {
					log.Error(err)
				}
			})
		} else {
			log.Warn("target direction signal not found")
		}
		a.Fan.AddC(direction.C)
	}

	a.AddS(a.Fan.S)
	return a
}
