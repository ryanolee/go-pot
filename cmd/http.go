package cmd

import (
	"fmt"
	"os"

	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/di"
	"github.com/spf13/cobra"
)

var httpCommand = &cobra.Command{
	Use:   "http",
	Short: "Starts the HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.NewConfig(cmd, config.GetHttpFlags())

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Make sure only the HTTP server is enabled
		conf.FtpServer.Enabled = false
		conf.Server.Disable = false

		di := di.CreateContainer(conf)
		di.Run()
	},
}

func init() {
	config.BindConfigFlags(httpCommand, config.GetHttpFlags())
	config.BindConfigFileFlags(httpCommand)
	rootCmd.AddCommand(httpCommand)
}
