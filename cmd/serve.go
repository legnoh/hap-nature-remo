package cmd

import (
	"context"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/creasty/defaults"
	"github.com/legnoh/hap-nature-remo/additionalaccessory"
	"github.com/legnoh/hap-nature-remo/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tenntenn/natureremo"
)

type A []*accessory.A

var (
	accessories A
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

	// ブリッジ作成
	bridge := accessory.NewBridge(accessory.Info{
		Name:         conf.Name,
		Firmware:     version,
		Manufacturer: "@legnoh",
	})

	// Natureデバイス一覧を取得
	nr := natureremo.NewClient(conf.Token)
	nrDevices := util.GetDevices(nr)

	// センサーが1つでもあった場合はSensorアプライアンスを作る
	for _, device := range nrDevices.Devices {
		if len(device.NewestEvents) != 0 {
			a := additionalaccessory.NewSensor(nr, *device)
			accessories = append(accessories, a.A)
		}
	}

	// NatureRemoに登録済の家電一覧を取得し、全ての家電から操作可能なものを登録していく
	for _, appliance := range util.GetAppliances(nr).Appliances {

		// リモコン式ファンがある場合はFanアプライアンスを作る
		if appliance.Type == natureremo.ApplianceTypeIR && appliance.Image == "ico_fan" {
			log.Infof("Compatible Appliance Found: %s(%s)", appliance.Nickname, appliance.ID)
			a := additionalaccessory.NewFan(nr, appliance)
			accessories = append(accessories, a.A)
		}

		// エアコン(NatureRemo対応のもの)がある場合はAirConditionerアプライアンスを作る
		if appliance.Type == natureremo.ApplianceTypeAirCon {
			log.Infof("Compatible Appliance Found: %s(%s)", appliance.Nickname, appliance.ID)
			a := additionalaccessory.NewAirConditioner(nr, appliance, nrDevices.Devices)
			accessories = append(accessories, a.A)
		}
	}

	server, err := hap.NewServer(hap.NewFsStore(fsStoreDirectory), bridge.A, accessories...)
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
