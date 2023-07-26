package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Config struct {
	Token string
	Name  string `default:"hap-nature-remo"`
	Pin   string `default:"12344321"`
	Fans  []struct {
		Nickname string
	}
}

var (
	cfgFile          string
	conf             Config
	confDir          string
	fsStoreDirectory string
	resetFs          bool
	version          string
	debug            bool
	log              *logrus.Logger
)

var rootCmd = &cobra.Command{
	Use:     "hap-nature-remo",
	Version: version,
	Short:   "Nature Remo HomeKit Bridge",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	if rootCmd.Execute() != nil {
		log.Fatal("Root execute is failed... exit")
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "print debug log")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	debug, err := rootCmd.PersistentFlags().GetBool("debug")
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		log.SetLevel(logrus.DebugLevel)
	}
}
