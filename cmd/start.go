package cmd

import (
	"fmt"
	"os"

	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/di"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Allows for all pots to be started from a single command",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.NewConfig(cmd, config.GetStartFlags())
		if err != nil {
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
