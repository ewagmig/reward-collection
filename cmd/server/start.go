package server

import (
	decron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/starslabhq/rewards-collection/controllers"
	"github.com/starslabhq/rewards-collection/models"
	"github.com/starslabhq/rewards-collection/server"
	"github.com/starslabhq/rewards-collection/utils"
	"github.com/starslabhq/rewards-collection/version"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	prod = "prod"
)

var listenAddr string

func startCmd() *cobra.Command {
	flags := serverStartCmd.Flags()
	flags.StringVarP(&listenAddr, "addr", "a", ":8005", "Listen address")
	return serverStartCmd
}

func start(mode string) error {
	logrus.Infof("Start heco common component server in %s mode, with %s", mode, version.Version())

	s := server.New(getServerOptions(mode)...)
	// make the cron job
	c := decron.New()
	//refresh every 10 minute
	c.AddFunc("@every 3m", models.SyncEpochBackground)

	c.AddFunc("@every 3m", models.ProcessSendBackground)
	c.Start()

	// Startup server to accept requests...
	fastexit := make(chan struct{})
	go func() {
		err := s.Startup(listenAddr)
		if err != nil {
			logrus.Errorf("Fail to startup server with error: %v", err)
			fastexit <- struct{}{}
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case <-quit:
		logrus.Info("Shutdown baas server...")
		if err := s.Shutdown(3 * time.Second); err != nil {
			logrus.Errorf("Fail to shutdown baas server with error: %v", err)
			return err
		}
	case <-fastexit:
		// DOTHING
	}

	logrus.Info("Baas server exists...")
	return nil
}

func getServerOptions(mode string) []server.Option {
	opts := []server.Option{}

	mod := server.DEV
	if strings.ToLower(mode) == prod {
		mod = server.PROD
	}
	opts = append(opts, server.WithMode(mod))

	if basePath := viper.GetString("request.basepath"); basePath != "" {
		opts = append(opts, server.RequestBasePath(basePath))
	}

	if disableCtrls := viper.GetStringSlice("controller.disable"); len(disableCtrls) > 0 {
		opts = append(opts, server.ControllerFilter(func(c server.Controller) bool {
			return utils.StrInSlice(disableCtrls, c.Name())
		}))
	}

	//if enableMws := viper.GetStringSlice("middleware.enable"); len(enableMws) > 0 {
	//	opts = append(opts, server.MiddlewareFilter(func(m server.Middleware) bool {
	//		return utils.StrInSlice(enableMws, m.Name())
	//	}))
	//}

	if d := viper.GetDuration("request.timeout"); d > 0 {
		opts = append(opts, server.RequestTimeout(d))
	}

	if v := viper.GetString("api.version"); v != "" {
		opts = append(opts, server.WithVersion(v))
	}
	return opts
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the baas server",
	Long:  "Start the baas server to listen remote HTTP API requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := cmd.Flags().GetString("mode")
		if err != nil {
			return err
		}
		return start(mode)
	},
}
