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
	Short: "Resync secrets from gitleaks",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		conf, err := config.NewConfig(cmd)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		
		di := di.CreateContainer(conf)
		di.Run()
	},
}

func init() {
	config.BindConfigFlags(startCmd)
	rootCmd.AddCommand(startCmd)
}
