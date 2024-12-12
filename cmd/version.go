package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version    = "Unknown"
	buildDate  = "Unknown"
	commitHash = "Unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Outputs the current version of the CLI with commit and date.",
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Print(version)
			return
		}

		fmt.Printf("Version: %s\nBuild Date: %s\nCommit Hash: %s\n", version, buildDate, commitHash)
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "Print only the version number.")
	rootCmd.AddCommand(versionCmd)
}
