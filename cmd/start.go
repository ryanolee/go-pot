package cmd

import (
	"fmt"
	"os"

	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/di"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Allows for all pots to be started from a single command",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.NewConfig(cmd, config.GetStartFlags())
		if err != nil {
			fmt.Println("Failed to start go pot due to a bad configuration. Please check your GO__POT__ environment variables, cli flags and config file (if set).\nThe errors are as follows:")
			fmt.Println(err)
			os.Exit(1)
		}

		di := di.CreateContainer(conf)
		di.Run()

	},
}

func init() {
	config.BindConfigFlags(startCmd, config.GetStartFlags())
	config.BindConfigFileFlags(startCmd)
	rootCmd.AddCommand(startCmd)
}
