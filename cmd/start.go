package cmd

import (
	"fmt"
	"os"

	"github.com/ryanolee/ryan-pot/http"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Resync secrets from gitleaks",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		panic(http.Serve(http.ServerConfig{
			Port:  port,
			Debug: true,
		}))
	},
}

func init() {
	startCmd.Flags().Int("port", 8080, "Port to listen on")
	rootCmd.AddCommand(startCmd)
}
