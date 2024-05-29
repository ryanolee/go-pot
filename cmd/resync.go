package cmd

import (
	"fmt"
	"os"

	"github.com/ryanolee/ryan-pot/secrets"
	"github.com/spf13/cobra"
)

var resyncCmd = &cobra.Command{
	Use:   "resync",
	Short: "Resync secrets from gitleaks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Resyncing secret definitions from gitleaks")
		data, err := secrets.UpdateSecretsFromGitLeaks()
		if err != nil {
			panic(err)
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		err = os.MkdirAll(homeDir + "/.config/go-pot", 0755)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(homeDir + "/.config/go-pot/go-pot-rules.yml", data, 0755)
		if err != nil {
			panic(err)
		}

		fmt.Println("Successfully resynced secret definitions from gitleaks")
	},
}

func init() {
	// @todo: This needs reworking
	//rootCmd.AddCommand(resyncCmd)
}
