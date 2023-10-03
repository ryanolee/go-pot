package cmd

import (
	"github.com/ryanolee/ryan-pot/http"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Resync secrets from gitleaks",
	Run: func(cmd *cobra.Command, args []string) {
		panic(http.Serve(http.ServerConfig{
			Port: 8080,
			Debug: true,
		}))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
