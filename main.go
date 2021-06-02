package main

import (
	"fmt"
	"github.com/awnumar/memguard"
	"github.com/op/go-logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdsvr "github.com/starslabhq/rewards-collection/cmd/server"
	kafkalog "github.com/starslabhq/rewards-collection/log"
	"github.com/starslabhq/rewards-collection/version"
	"log"
	"os"
	"strings"
)

const (
	envRoot              = "common"
	dev                  = "dev"
	defaultLoggingFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
)

var (
	mode        string
	versionFlag bool
	logOutput   = os.Stderr
	//logger      = logging.MustGetLogger("common.main")
	//logger = logrus.New()
	mainCmd = &cobra.Command{
		Use:   "commonComponent",
		Short: "commonComponent is a utility for managing fabric blockchain configurations",
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				fmt.Printf("Application %s\n%s\n", viper.GetString("application.name"), version.Version())
			} else {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}
)

func init() {
	initConf()
	mainFlags := mainCmd.PersistentFlags()
	mainFlags.BoolVarP(&versionFlag, "version", "v", false, "Display current version of Common Component")
	mainFlags.StringVarP(&mode, "mode", "m", dev, "Mode, dev or prod")
	viper.SetDefault("organization", "orgorderer")
	viper.SetDefault("username", "Admin")
	viper.SetDefault("password", "Admin")
	viper.SetDefault("gm", false)
	//log entrypoint here
	topic := viper.GetString("log.topic")
	brokers := viper.GetStringSlice("log.kafka.servers")
	level := viper.GetInt("log.level")

	err := kafkalog.AddKafkaHook(topic, brokers, level)
	if err != nil {
		fmt.Println(err)
	}
	kafkalog.AddConsoleOut(5)
	kafkalog.AddField("app", "reward-collection")
}

func main() {
	// MemGuard
	memguard.DisableUnixCoreDumps()
	// Tell memguard to listen out for interrupts, and cleanup in case of one.
	memguard.CatchInterrupt(func() {
		log.Println("MemGuard Interrupt")
		log.Println("Interrupt signal received. Exiting...")
	})

	// Make sure to destroy all LockedBuffers when returning.
	defer memguard.DestroyAll()

	mainCmd.AddCommand(cmdsvr.Cmd())

	if mainCmd.Execute() != nil {
		os.Exit(1)
	}
}

func initLogger() {
	formatSpec := viper.GetString("logging.format")
	if formatSpec == "" {
		formatSpec = defaultLoggingFormat
	}

	formatter := logging.MustStringFormatter(formatSpec)
	backend := logging.NewLogBackend(logOutput, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, formatter)

	lls := viper.GetString("logging.level")
	level, err := logging.LogLevel(lls)
	if err != nil {
		panic(fmt.Errorf("Fatal error invalid logging level: %s ", err))
	}
	logging.SetBackend(backendFormatter).SetLevel(level, "")
}

func initConf() {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetConfigName(mode)
	//envVal := os.Getenv("FABRIC_BAAS_CFG_PATH")
	//fmt.Println("enval", envVal)
	//if envVal != "" {
	//	viper.AddConfigPath(envVal)
	//} else {
	//	viper.AddConfigPath("conf/")
	//}


	viper.AddConfigPath("conf/")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Errorf("Fatal error config file: %s", err)
		os.Exit(1)
	}
}

