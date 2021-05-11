package server

import (
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

const (
	serverFuncName = "server"
	serverDes      = "Operate a common component server: start"
)

var (
	logger = logging.MustGetLogger("common.cmd.server")
)

func Cmd() *cobra.Command {
	serverCmd.AddCommand(startCmd())
	return serverCmd
}

var serverCmd = &cobra.Command{
	Use:   serverFuncName,
	Short: serverDes,
	Long:  serverDes,
	//PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
	//	return acmd.InitDBConnectionString()
	//},
}


