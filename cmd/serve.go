package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
	"github.com/creasty/defaults"
	"github.com/legnoh/hap-nature-remo/additionalaccessory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tenntenn/natureremo"
)

type A []*accessory.A

type HeaterCooler struct {
	*A
	HeaterCooler *service.HeaterCooler
}

type NrDevices struct {
	Devices   []*natureremo.Device
	UpdatedAt time.Time
}

type NrAppliances struct {
	Appliances []*natureremo.Appliance
	UpdatedAt  time.Time
}

var (
	bridgeMeta   accessory.Info
	accessories  A
	nrDevices    NrDevices
	nrAppliances NrAppliances
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start homekit bridge daemon",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: preStartServer,
	Run:    startServer,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	confDir = home + "/.hap-nature-remo"

	serveCmd.Flags().StringVarP(&cfgFile, "config", "c", confDir+"/config.yml", "config file path")
	serveCmd.Flags().StringVarP(&fsStoreDirectory, "fs-store", "f", confDir+"/db", "fsStore directory path")
	serveCmd.Flags().BoolVar(&resetFs, "reset", false, "reset fsStore before start")
}

func preStartServer(cmd *cobra.Command, args []string) {
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatal(err)
	}
	if err := defaults.Set(&conf); err != nil {
		log.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9]{8}$`).MatchString(conf.Pin) {
		log.Fatalf("Your PinCode(%s) is invalid format. Please fix to 8-digit code(e.g. 12344321)", conf.Pin)
	}

	if resetFs {
		err := os.RemoveAll(fsStoreDirectory)
		if err != nil {
			log.Fatalf("Reset FileStore Failed: %s", err)
		} else {
			log.Infof("Reset FileStore Successfully: %s", fsStoreDirectory)
		}
	}
}

func startServer(cmd *cobra.Command, args []string) {

	bridgeMeta.Name = conf.Name

	nr := natureremo.NewClient(conf.Token)

	// Natureデバイス一覧を取得
	nrDevices := getDevices(nr)

	// 全てのデバイスから、センサー数値が取れるものを記録する
	sensors := make(map[natureremo.SensorType]float64)
	for _, d := range nrDevices.Devices {
		if val, ok := d.NewestEvents[natureremo.SensorTypeTemperature]; ok {
			log.Infof("Temperature Sensor Detected(%s): %.1f", d.Name, val.Value)
			sensors[natureremo.SensorTypeTemperature] = val.Value
			setBridgeFirmwareInfo(d.DeviceCore)
		}
		if val, ok := d.NewestEvents[natureremo.SensorTypeHumidity]; ok {
			log.Infof("Humidity Sensor Detected(%s): %.1f", d.Name, val.Value)
			sensors[natureremo.SensorTypeHumidity] = val.Value
		}
		if val, ok := d.NewestEvents[natureremo.SensorTypeIllumination]; ok {
			log.Infof("Illumination Sensor Detected(%s): %.1f", d.Name, val.Value)
			sensors[natureremo.SensorTypeIllumination] = val.Value
		}
		if val, ok := d.NewestEvents[natureremo.SensorTypeMovement]; ok {
			log.Infof("Movement Sensor Detected(%s): %.1f", d.Name, val.Value)
			sensors[natureremo.SensorTypeMovement] = val.Value
		}
	}

	// センサーが1つでもある場合はSensorアプライアンスを作る
	if len(sensors) != 0 {
		a := createSensorAppliance(sensors)
		accessories = append(accessories, a.A)
	}

	// NatureRemoに登録済の家電一覧を取得
	nrAppliances = getAppliances(nr)

	// 全ての家電から操作可能なものを登録していく
	for _, v := range nrAppliances.Appliances {

		// リモコン式ファンがある場合はFunアプライアンスを作る
		if v.Type == natureremo.ApplianceTypeIR && v.Image == "ico_fan" {
			log.Infof("Compatible Appliance Found: %s(%s)", v.Nickname, v.ID)
			a := createFunAppliance(v)
			accessories = append(accessories, a.A)
		}

		// エアコン(NatureRemo対応のもの)がある場合はAirConditionerアプライアンスを作る
		if v.Type == natureremo.ApplianceTypeAirCon {
			log.Infof("Compatible Appliance Found: %s(%s)", v.Nickname, v.ID)
			a := createAirConditionerAppliance(v, sensors)
			accessories = append(accessories, a.A)
			setBridgeFirmwareInfo(*v.Device)
		}
	}

	bridge := accessory.NewBridge(bridgeMeta)
	fs := hap.NewFsStore(fsStoreDirectory)

	server, err := hap.NewServer(fs, bridge.A, accessories...)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	server.Pin = conf.Pin

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		log.Info("Stopping HAP Server...")
		signal.Stop(c)
		cancel()
	}()

	log.Info("Starting HAP Server...")
	log.Infof("Device Name: %s", bridge.Name())
	log.Infof("   Pin Code: %s", conf.Pin)
	log.Debugf("Config File: %s", cfgFile)
	log.Debugf(" Store Path: %s", fsStoreDirectory)
	server.ListenAndServe(ctx)
}

func createSensorAppliance(sensors map[natureremo.SensorType]float64) *additionalaccessory.Sensor {
	a := additionalaccessory.NewSensor(accessory.Info{
		Name:         bridgeMeta.Name,
		Manufacturer: bridgeMeta.Manufacturer,
		Model:        bridgeMeta.Model,
		SerialNumber: bridgeMeta.SerialNumber,
	}, sensors)
	return a
}

func createFunAppliance(v *natureremo.Appliance) *accessory.Fan {

	nr := natureremo.NewClient(conf.Token)
	nrctx := context.Background()
	speedRe := regexp.MustCompile(`^ico_number_(\d)$`)
	directionRe := regexp.MustCompile(`^ico_(.*)ward$`)

	a := accessory.NewFan(accessory.Info{
		Name: v.Nickname,
	})

	signals, err := nr.SignalService.GetAll(nrctx, v)
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
			if nr.SignalService.Send(nrctx, targetSignal) != nil {
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

		if fFound && bFound {
			direction.OnValueRemoteUpdate(func(v int) {
				log.Infof("rotation changed: %d", v)
				if v == characteristic.RotationDirectionClockwise {
					if nr.SignalService.Send(nrctx, f) != nil {
						log.Error(err)
					}
				} else if v == characteristic.RotationDirectionCounterclockwise {
					if nr.SignalService.Send(nrctx, b) != nil {
						log.Error(err)
					}
				}
			})
		} else if fFound {
			direction.OnValueRemoteUpdate(func(v int) {
				if nr.SignalService.Send(nrctx, f) != nil {
					log.Error(err)
				}
			})
		} else if bFound {
			direction.OnValueRemoteUpdate(func(v int) {
				if nr.SignalService.Send(nrctx, b) != nil {
					log.Error(err)
				}
			})
		} else {
			log.Warn("target direction signal not found")
		}
		a.Fan.AddC(direction.C)
	}
	return a
}

func createAirConditionerAppliance(ac *natureremo.Appliance, sensors map[natureremo.SensorType]float64) *additionalaccessory.AirConditioner {

	nr := natureremo.NewClient(conf.Token)
	nrctx := context.Background()

	acceInfo := accessory.Info{
		Name:         ac.Nickname,
		Manufacturer: ac.Model.Manufacturer,
		Model:        ac.Model.RemoteName,
		SerialNumber: ac.ID,
	}
	a := additionalaccessory.NewAirConditioner(acceInfo)

	targetState := a.HeaterCooler.TargetHeaterCoolerState.ValidVals
	currentState := a.HeaterCooler.CurrentHeaterCoolerState.ValidVals

	// 冷房がある場合の処理
	if f, found := ac.AirCon.Range.Modes[natureremo.OperationModeCool]; found {

		log.Infof("Cooler detected: %s", ac.Nickname)
		targetState = append(targetState, characteristic.TargetHeaterCoolerStateCool)
		currentState = append(currentState, characteristic.CurrentHeaterCoolerStateCooling)

		min, max, step := getStepInfo(f.Temperature)
		log.Debugf("Cooling range: %2f ~ %2f", min, max)
		threshold := *characteristic.NewCoolingThresholdTemperature()
		threshold.SetMinValue(min)
		threshold.SetMaxValue(max)
		threshold.SetStepValue(step)
		nowSetting, _ := strconv.ParseFloat(ac.AirConSettings.Temperature, 64)
		threshold.SetValue(nowSetting)

		// 設定温度が変わった時(冷房)の処理
		threshold.OnValueRemoteUpdate(func(v float64) {
			target := strconv.FormatFloat(v, 'f', -1, 64)
			log.Infof("AirConditioner(Cooler) Temperature Updated: %s", target)
			err := nr.ApplianceService.UpdateAirConSettings(nrctx, ac, &natureremo.AirConSettings{
				Temperature: target,
			})
			if err != nil {
				log.Error(err)
			}
		})

		// 現在の設定値を呼び出された時の処理
		threshold.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debug("Get now AirConditioner threshold Request")
			aps := getAppliances(nr)
			for _, ap := range aps.Appliances {
				if ap.ID == ac.ID {
					temp, _ := strconv.ParseFloat(ac.AirConSettings.Temperature, 64)
					log.Debugf("Get now AirConditioner threshold Successful: %1f", temp)
					return temp, 0
				}
			}
			return nil, -1
		}
		a.HeaterCooler.AddC(threshold.C)
	}

	// 暖房がある場合の処理
	if f, found := ac.AirCon.Range.Modes[natureremo.OperationModeWarm]; found {

		log.Infof("Heater detected: %s", ac.Nickname)
		targetState = append(targetState, characteristic.TargetHeaterCoolerStateHeat)
		currentState = append(currentState, characteristic.CurrentHeaterCoolerStateHeating)

		min, max, step := getStepInfo(f.Temperature)
		log.Debugf("Heating range: %2f ~ %2f", min, max)
		threshold := *characteristic.NewHeatingThresholdTemperature()
		threshold.SetMinValue(min)
		threshold.SetMaxValue(max)
		threshold.SetStepValue(step)
		nowSetting, _ := strconv.ParseFloat(ac.AirConSettings.Temperature, 64)
		threshold.SetValue(nowSetting)

		// 設定温度が変わった時(暖房)の処理
		threshold.OnValueRemoteUpdate(func(v float64) {
			target := strconv.FormatFloat(v, 'f', -1, 64)
			log.Infof("AirConditioner(Heater) Temperature Updated: %s", target)
			err := nr.ApplianceService.UpdateAirConSettings(nrctx, ac, &natureremo.AirConSettings{
				Temperature: target,
			})
			if err != nil {
				log.Error(err)
			}
		})

		// 現在の設定値を呼び出された時の処理
		threshold.ValueRequestFunc = func(*http.Request) (interface{}, int) {
			log.Debug("Get now AirConditioner threshold Request")
			aps := getAppliances(nr)
			for _, ap := range aps.Appliances {
				if ap.ID == ac.ID {
					temp, _ := strconv.ParseFloat(ac.AirConSettings.Temperature, 64)
					return temp, 0
				}
			}
			return nil, -1
		}
		a.HeaterCooler.AddC(threshold.C)
	}

	a.HeaterCooler.TargetHeaterCoolerState.ValidVals = targetState
	a.HeaterCooler.CurrentHeaterCoolerState.ValidVals = currentState

	// 動作モードが変わった時の処理
	a.HeaterCooler.TargetHeaterCoolerState.OnValueRemoteUpdate(func(v int) {
		log.Infof("AirConditioner Mode Changed: %d", v)
	})

	// 動作モード確認の処理
	a.HeaterCooler.TargetHeaterCoolerState.ValueRequestFunc = func(*http.Request) (interface{}, int) {
		log.Debug("Get now AirConditioner Mode Request")
		aps := getAppliances(nr)
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

	// 現在の動作状況確認を初期状態で入れる処理(室温・電源・モード)
	if s, sFound := sensors[natureremo.SensorTypeTemperature]; sFound {
		a.HeaterCooler.CurrentTemperature.SetValue(s)
	}
	switch ac.AirConSettings.OperationMode {
	case natureremo.OperationModeCool:
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateCooling)
		a.HeaterCooler.TargetHeaterCoolerState.SetValue(characteristic.TargetHeaterCoolerStateCool)
	case natureremo.OperationModeWarm:
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateHeating)
		a.HeaterCooler.TargetHeaterCoolerState.SetValue(characteristic.TargetHeaterCoolerStateHeat)
	default:
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateIdle)
	}
	if ac.AirConSettings.Button == natureremo.ButtonPowerOff {
		a.HeaterCooler.CurrentHeaterCoolerState.SetValue(characteristic.CurrentHeaterCoolerStateInactive)
	}
	return a
}

func setBridgeFirmwareInfo(d natureremo.DeviceCore) {

	if bridgeMeta.Firmware == "" {
		bridgeMeta.Model = d.Name
		bridgeMeta.Manufacturer = "Nature Inc."
		bridgeMeta.Firmware = d.FirmwareVersion
		bridgeMeta.SerialNumber = d.SerialNumber
	}
}

func getStepInfo(values []string) (float64, float64, float64) {

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

// NatureRemoの Appliance取得リクエストを行う関数
// (大量のリクエストが走ることを防ぐため、10秒未満のリクエストの場合は前回のリクエスト結果を使う)
func getAppliances(nr *natureremo.Client) NrAppliances {

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
func getDevices(nr *natureremo.Client) NrDevices {

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
