package cmd

import (
	"fmt"
	"os"

	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/di"
	"github.com/spf13/cobra"
)

var ftpCommand = &cobra.Command{
	Use:   "ftp",
	Short: "Starts the FTP server",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.NewConfig(cmd, config.GetFtpFlags())

		if err != nil {
			fmt.Println("Failed to start go pot in FTP mode due to a bad configuration.  Please check your GO__POT__ environment variables, cli flags and config file (if set).\nThe errors are as follows:")
			fmt.Println(err)
			os.Exit(1)
		}

		// Make sure only the FTP server is enabled
		conf.FtpServer.Enabled = true

		// Disable the server and multi protocol
		conf.Server.Disable = true
		conf.MultiProtocol.Enabled = false

		di := di.CreateContainer(conf)
		di.Run()

	},
}

func init() {
	config.BindConfigFlags(ftpCommand, config.GetFtpFlags())
	config.BindConfigFileFlags(ftpCommand)
	rootCmd.AddCommand(ftpCommand)
}
