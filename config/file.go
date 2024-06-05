package config

import (
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

func loadConfigFile(k *koanf.Koanf, command *cobra.Command) error {
	configFile := command.Flag("config-file").Value.String()
	if configFile == "" {
		configFile = os.Getenv("GOPOT__CONFIG_FILE")
	}

	if configFile == "" {
		return nil
	}

	if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
		return err
	}

	return nil
}

func BindConfigFileFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().String("config-file", "", "The path to the configuration file to use. (In yml format)")
	return cmd
}
