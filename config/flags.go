package config

import (
	"maps"
	"os"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type (
	flagConfig struct {
		flagName     string
		configKey    string
		description  string
		configType   string
		defaultValue interface{}
	}

	flagMap map[string]flagConfig
)

var commonFlags = flagMap{
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
	"log-path": {
		flagName:     "log-path",
		configKey:    "logging.path",
		description:  "The path to write the log to. (If not set, logs will be written to stdout.)",
		configType:   "string",
		defaultValue: defaultConfig.Logging.Path,
	},
}

var httpFlags = flagMap{

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
	"http-access-log-mode": {
		flagName:     "http-access-log-mode",
		configKey:    "server.access_log.mode",
		description:  "The mode to log requests in. Options: start (start of the request), end (end of the request), both (both start and end of the request), none (no logging).",
		configType:   "string",
		defaultValue: defaultConfig.Server.AccessLog.Path,
	},
	"http-access-log-path": {
		flagName:     "http-access-log-path",
		configKey:    "server.access_log.path",
		description:  "The path to write the http access log to. (If not set, logs will be written to stdout.)",
		configType:   "string",
		defaultValue: defaultConfig.Server.AccessLog.Path,
	},
	"http-access-log-fields": {
		flagName:     "http-access-log-fields",
		configKey:    "server.access_log.fields_to_log",
		description:  "The fields to log in the http access log as comma separated values. (Lookup documentation for available fields.)",
		configType:   "string",
		defaultValue: strings.Join(defaultConfig.Server.AccessLog.FieldsToLog, ","),
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
	"bytes-per-second": {
		flagName:     "bytes-per-second",
		configKey:    "staller.bytes_per_second",
		description:  "The number of bytes to transfer per second.",
		configType:   "int",
		defaultValue: defaultConfig.Staller.BytesPerSecond,
	},
}

var ftpFlags = flagMap{
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
	"ftp-log-path": {
		flagName:     "ftp-log-path",
		configKey:    "ftp_server.command_log.path",
		description:  "The path to write the ftp command log to. (If not set, logs will be written to stdout.)",
		configType:   "string",
		defaultValue: defaultConfig.FtpServer.CommandLog.Path,
	},
	"ftp-log-commands": {
		flagName:     "ftp-log-commands",
		configKey:    "ftp_server.command_log.commands_to_log",
		description:  "The commands to log in the ftp command log as comma separated values. (Lookup documentation for available commands.)",
		configType:   "string",
		defaultValue: strings.Join(defaultConfig.FtpServer.CommandLog.CommandsToLog, ","),
	},
	"ftp-log-fields": {
		flagName:     "ftp-log-fields",
		configKey:    "ftp_server.command_log.additional_fields",
		description:  "The additional fields to log in each line of the FTP log. (Lookup documentation for available fields.)",
		configType:   "string",
		defaultValue: strings.Join(defaultConfig.FtpServer.CommandLog.AdditionalFields, ","),
	},
}

var startFlags = flagMap{
	"http-disabled": {
		flagName:     "http-disabled",
		configKey:    "server.disable",
		description:  "Disables the http server for the honeypot.",
		configType:   "bool",
		defaultValue: defaultConfig.Server.Disable,
	},

	"ftp-enabled": {
		flagName:     "ftp-enabled",
		configKey:    "ftp_server.enabled",
		description:  "Enable the FTP service.",
		configType:   "bool",
		defaultValue: defaultConfig.FtpServer.Enabled,
	},
}

func GetStartFlags() flagMap {
	allFlags := make(flagMap)

	maps.Copy(allFlags, commonFlags)
	maps.Copy(allFlags, httpFlags)
	maps.Copy(allFlags, ftpFlags)
	maps.Copy(allFlags, startFlags)

	return allFlags
}

func GetFtpFlags() flagMap {
	internalFtpFlags := make(flagMap)
	maps.Copy(internalFtpFlags, ftpFlags)
	maps.Copy(internalFtpFlags, commonFlags)

	return internalFtpFlags
}

func GetHttpFlags() flagMap {
	internalHttpFlags := make(flagMap)
	maps.Copy(internalHttpFlags, httpFlags)
	maps.Copy(internalHttpFlags, commonFlags)

	return internalHttpFlags
}

// Binds configuration flags to the provided command
func BindConfigFlags(cmd *cobra.Command, flagsToMap flagMap) *cobra.Command {
	for _, flag := range flagsToMap {
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

func writeFlagValues(k *koanf.Koanf, cmd *cobra.Command, flagsUsed flagMap) *koanf.Koanf {
	for _, flag := range flagsUsed {
		if !cmd.Flags().Changed(flag.flagName) {
			continue
		}

		switch flag.configType {
		case "int":
			if err := k.Set(flag.configKey, readFlagOrPanic(cmd.Flags().GetInt(flag.flagName))); err != nil {
				continue
			}
		case "string":
			if err := k.Set(flag.configKey, readFlagOrPanic(cmd.Flags().GetString(flag.flagName))); err != nil {
				continue
			}
		case "bool":
			if err := k.Set(flag.configKey, readFlagOrPanic(cmd.Flags().GetBool(flag.flagName))); err != nil {
				continue
			}
		}
	}

	return k
}
