package config

import (
	"os"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type flagConfig struct {
	flagName     string
	configKey    string
	description  string
	configType   string
	defaultValue interface{}
}

var flagsToConfigMap = map[string]flagConfig{
	"http-disabled": {
		flagName:     "http-disabled",
		configKey:    "server.disable",
		description:  "Disables the http server for the honeypot.",
		configType:   "bool",
		defaultValue: defaultConfig.Server.Disable,
	},
	"port": {
		flagName:     "port",
		configKey:    "server.port",
		description:  "The port for the honeypot to listen on.",
		configType:   "int",
		defaultValue: defaultConfig.Server.Port,
	},
	"host": {
		flagName:     "host",
		configKey:    "server.host",
		description:  "The host for the honeypot to listen on.",
		configType:   "string",
		defaultValue: defaultConfig.Server.Host,
	},
	"network": {
		flagName:     "network",
		configKey:    "server.network",
		description:  "The network stack to use (tcp, tcp4, tcp6).",
		configType:   "string",
		defaultValue: defaultConfig.Server.Network,
	},
	"cluster-mode-enabled": {
		flagName:     "cluster-mode-enabled",
		configKey:    "cluster.enabled",
		description:  "Enable cluster mode for connectivity with other honeypots.",
		configType:   "bool",
		defaultValue: defaultConfig.Cluster.Enabled,
	},
	"cluster-advertise-ip": {
		flagName:     "cluster-advertise-ip",
		configKey:    "cluster.advertise_ip",
		description:  "The IP address to advertise to other honeypots in the cluster.",
		configType:   "string",
		defaultValue: defaultConfig.Cluster.AdvertiseIp,
	},
	"cluster-known-peers": {
		flagName:     "cluster-known-peers",
		configKey:    "cluster.known_peers",
		description:  "A comma separated list of known peers to connect to.",
		configType:   "string",
		defaultValue: "",
	},
	"cluster-port": {
		flagName:     "cluster-port",
		configKey:    "cluster.bind_port",
		description:  "The port for the honeypot to listen on for cluster communication. [This port should not be exposed to the internet.]",
		configType:   "int",
		defaultValue: defaultConfig.Cluster.BindPort,
	},
	"cluster-logging-enabled": {
		flagName:     "cluster-logging-enabled",
		configKey:    "cluster.enable_logging",
		description:  "Enable cluster communication logging. (Useful for debugging cluster connectivity issues)",
		configType:   "bool",
		defaultValue: defaultConfig.Cluster.EnableLogging,
	},
	"telemetry-name": {
		flagName:     "telemetry-name",
		configKey:    "telemetry.node_name",
		description:  "The telemetry node name.",
		configType:   "string",
		defaultValue: defaultConfig.Telemetry.NodeName,
	},
	"push-gateway-enabled": {
		flagName:     "push-gateway-enabled",
		configKey:    "telemetry.push_gateway.enabled",
		description:  "Enable prometheus push gateway integration.",
		configType:   "bool",
		defaultValue: defaultConfig.Telemetry.PushGateway.Enabled,
	},
	"push-gateway-url": {
		flagName:     "push-gateway-url",
		configKey:    "telemetry.push_gateway.endpoint",
		description:  "The URL for the prometheus push gateway.",
		configType:   "string",
		defaultValue: defaultConfig.Telemetry.PushGateway.Endpoint,
	},
	"prometheus-enabled": {
		flagName:     "prometheus-enabled",
		configKey:    "telemetry.prometheus.enabled",
		description:  "Enable prometheus metrics collection endpoint.",
		configType:   "bool",
		defaultValue: defaultConfig.Telemetry.Prometheus.Enabled,
	},
	"prometheus-path": {
		flagName:     "prometheus-path",
		configKey:    "telemetry.prometheus.path",
		description:  "The path for the prometheus metrics collection endpoint.",
		configType:   "string",
		defaultValue: defaultConfig.Telemetry.Prometheus.Path,
	},
	"prometheus-port": {
		flagName:     "prometheus-port",
		configKey:    "telemetry.prometheus.port",
		description:  "The port for the prometheus metrics collection endpoint.",
		configType:   "int",
		defaultValue: defaultConfig.Telemetry.Prometheus.Port,
	},
	"recast-enabled": {
		flagName:     "recast-enabled",
		configKey:    "recast.enabled",
		description:  "Enable recast metrics collection.",
		configType:   "bool",
		defaultValue: defaultConfig.Recast.Enabled,
	},
	"maximum-connections": {
		flagName:     "maximum-connections",
		configKey:    "staller.maximum_connections",
		description:  "The maximum number of open connections to the honeypot to allow.",
		configType:   "int",
		defaultValue: defaultConfig.Staller.MaximumConnections,
	},
	"bytes-per-second": {
		flagName:     "bytes-per-second",
		configKey:    "staller.bytes_per_second",
		description:  "The number of bytes to transfer per second.",
		configType:   "int",
		defaultValue: defaultConfig.Staller.BytesPerSecond,
	},
	"ftp-enabled": {
		flagName:     "ftp-enabled",
		configKey:    "ftp_server.enabled",
		description:  "Enable the FTP service.",
		configType:   "bool",
		defaultValue: defaultConfig.FtpServer.Enabled,
	},
	"ftp-port": {
		flagName:     "ftp-port",
		configKey:    "ftp_server.port",
		description:  "The port for the FTP service to listen on.",
		configType:   "int",
		defaultValue: defaultConfig.FtpServer.Port,
	},
	"ftp-host": {
		flagName:     "ftp-host",
		configKey:    "ftp_server.host",
		description:  "The host for the FTP service to listen on.",
		configType:   "string",
		defaultValue: defaultConfig.FtpServer.Host,
	},
	"ftp-passive-ports": {
		flagName:     "ftp-passive-ports",
		configKey:    "ftp_server.passive_port_range",
		description:  "The range of passive ports to use for FTP data connections. (in the format of 'start-end')",
		configType:   "string",
		defaultValue: defaultConfig.FtpServer.PassivePortRange,
	},
}

// Binds configuration flags to the provided command
func BindConfigFlags(cmd *cobra.Command) *cobra.Command {
	for _, flag := range flagsToConfigMap {
		switch flag.configType {
		case "int":
			cmd.Flags().Int(flag.flagName, flag.defaultValue.(int), flag.description)
		case "string":
			cmd.Flags().String(flag.flagName, flag.defaultValue.(string), flag.description)
		case "bool":
			cmd.Flags().Bool(flag.flagName, flag.defaultValue.(bool), flag.description)
		}
	}

	return cmd
}

// Reads a flag or panics if the flag is not set
func readFlagOrPanic[F any](flag F, err error) F {
	if err != nil {
		zap.L().Sugar().Error(err)
		os.Exit(1)
	}

	return flag
}

func writeFlagValues(k *koanf.Koanf, cmd *cobra.Command) *koanf.Koanf {
	for _, flag := range flagsToConfigMap {
		if !cmd.Flags().Changed(flag.flagName) {
			continue
		}

		switch flag.configType {
		case "int":
			k.Set(flag.configKey, readFlagOrPanic(cmd.Flags().GetInt(flag.flagName)))
		case "string":
			k.Set(flag.configKey, readFlagOrPanic(cmd.Flags().GetString(flag.flagName)))
		case "bool":
			k.Set(flag.configKey, readFlagOrPanic(cmd.Flags().GetBool(flag.flagName)))
		}
	}

	return k
}
