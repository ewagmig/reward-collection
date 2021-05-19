package server

import (
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	acmd "github.com/starslabhq/rewards-collection/cmd"
)

const (
	serverFuncName = "server"
	serverDes      = "Operate a common component server: start | migrate"
)

var (
	logger = logging.MustGetLogger("common.cmd.server")
)

func Cmd() *cobra.Command {
	serverCmd.AddCommand(startCmd(),migrateCmd())
	return serverCmd
}

var serverCmd = &cobra.Command{
	Use:   serverFuncName,
	Short: serverDes,
	Long:  serverDes,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return acmd.InitDBConnectionString()
	},
}


