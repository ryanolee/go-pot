package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

// Binds configuration flags to the provided command
func BindConfigFlags(cmd *cobra.Command) *cobra.Command {
	// Server flags
	cmd.Flags().Int("port", defaultConfig.Server.Port, "The port for the honeypot to listen on.")
	cmd.Flags().String("host", defaultConfig.Server.Host, "The host for the honeypot to listen on.")

	// Cluster mode flags
	cmd.Flags().Bool("cluster-mode-enabled", defaultConfig.Cluster.Enabled, "Enable cluster mode for connectivity with other honeypots.")
	cmd.Flags().String("cluster-advertise-ip", defaultConfig.Cluster.AdvertiseIp, "The IP address to advertise to other honeypots in the cluster.")
	cmd.Flags().String("cluster-known-peers", "", "A comma separated list of known peers to connect to.")
	cmd.Flags().Int("cluster-port", defaultConfig.Cluster.BindPort, "The port for the honeypot to listen on for cluster communication. [This port should not be exposed to the internet.]")
	cmd.Flags().Bool("cluster-logging-enabled", defaultConfig.Cluster.EnableLogging, "Enable cluster communication logging. (Useful for debugging cluster connectivity issues)")

	// Telemetry flags
	cmd.Flags().String("telemetry-name", defaultConfig.Telemetry.NodeName, "The telemetry node name.")

	// Push gateway flags
	cmd.Flags().Bool("push-gateway-enabled", defaultConfig.Telemetry.PushGateway.Enabled, "Enable prometheus push gateway integration.")
	cmd.Flags().String("push-gateway-url", defaultConfig.Telemetry.PushGateway.Endpoint, "The URL for the prometheus push gateway.")
	
	// Prometheus flags
	cmd.Flags().Bool("prometheus-enabled", defaultConfig.Telemetry.Prometheus.Enabled, "Enable prometheus metrics collection endpoint.")
	cmd.Flags().String("prometheus-path", defaultConfig.Telemetry.Prometheus.Path, "The path for the prometheus metrics collection endpoint.")
	cmd.Flags().Int("prometheus-port", defaultConfig.Telemetry.Prometheus.Port, "The port for the prometheus metrics collection endpoint.")
	
	// Recast flags
	cmd.Flags().Bool("recast-enabled", defaultConfig.Recast.Enabled, "Enable recast metrics collection.")
	
	// Other flags
	cmd.Flags().Int("maximum-connections", defaultConfig.Staller.MaximumConnections, "The maximum number of open connections to the honeypot to allow.")
	cmd.Flags().Int("bytes-per-second", defaultConfig.Staller.BytesPerSecond, "The number of bytes to transfer per second.")

	return cmd
}

// Reads a flag or panics if the flag is not set
func readFlagOrPanic[F any](flag F, err error) F {
	if err != nil {
		panic(err)
	}
	return flag
}

func writeFlagValues(k *koanf.Koanf, cmd *cobra.Command) *koanf.Koanf {
	k.Set("server.port", readFlagOrPanic(cmd.Flags().GetInt("port")))
	k.Set("server.host", readFlagOrPanic(cmd.Flags().GetString("host")))
	k.Set("cluster.enabled", readFlagOrPanic(cmd.Flags().GetBool("cluster-mode-enabled")))
	k.Set("cluster.bind_port", readFlagOrPanic(cmd.Flags().GetInt("cluster-port")))
	k.Set("cluster.advertise_ip", readFlagOrPanic(cmd.Flags().GetString("cluster-advertise-ip")))
	k.Set("cluster.known_peers", readFlagOrPanic(cmd.Flags().GetString("cluster-known-peers")))
	k.Set("cluster.enable_logging", readFlagOrPanic(cmd.Flags().GetBool("cluster-logging-enabled")))
	k.Set("telemetry.node_name", readFlagOrPanic(cmd.Flags().GetString("telemetry-name")))
	k.Set("telemetry.push_gateway.enabled", readFlagOrPanic(cmd.Flags().GetBool("push-gateway-enabled")))
	k.Set("telemetry.push_gateway.endpoint", readFlagOrPanic(cmd.Flags().GetString("push-gateway-url")))
	k.Set("telemetry.prometheus.enabled", readFlagOrPanic(cmd.Flags().GetBool("prometheus-enabled")))
	k.Set("telemetry.prometheus.path", readFlagOrPanic(cmd.Flags().GetString("prometheus-path")))
	k.Set("telemetry.prometheus.port", readFlagOrPanic(cmd.Flags().GetInt("prometheus-port")))
	k.Set("recast.enabled", readFlagOrPanic(cmd.Flags().GetBool("recast-enabled")))
	k.Set("staller.maximum_connections", readFlagOrPanic(cmd.Flags().GetInt("maximum-connections")))
	k.Set("staller.bytes_per_second", readFlagOrPanic(cmd.Flags().GetInt("bytes-per-second")))

	return k
}