package cmd

import (
	"fmt"
	"os"

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
		di := di.CreateContainer()
		di.Run()
	},
}

func init() {
	startCmd.Flags().Int("port", 8080, "Port to listen on")
	rootCmd.AddCommand(startCmd)
}
