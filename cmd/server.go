package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yohamta/dagu/internal/admin"
	"github.com/yohamta/dagu/internal/config"
)

func serverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		Long:  `dagu server [--dags=<DAGs dir>] [--host=<host>] [--port=<port>]`,
		Run: func(cmd *cobra.Command, args []string) {
			server := admin.NewServer(config.Get())
			listenSignals(func(sig os.Signal) { server.Shutdown() })
			cobra.CheckErr(server.Serve())
		},
	}
	cmd.Flags().StringP("dags", "d", "", "location of DAG files (default is $HOME/.dagu/dags)")
	cmd.Flags().StringP("host", "s", "", "server port (default is 8080)")
	cmd.Flags().StringP("port", "p", "", "server host (default is localhost)")

	viper.BindPFlag("port", cmd.Flags().Lookup("port"))
	viper.BindPFlag("host", cmd.Flags().Lookup("host"))
	viper.BindPFlag("dags", cmd.Flags().Lookup("dags"))
	return cmd
}
