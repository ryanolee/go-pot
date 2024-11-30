package config

import (
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

type (
	// This struct covers the entire application configuration
	Config struct {
		Server         serverConfig         `koanf:"server"`
		FtpServer      ftpServerConfig      `koanf:"ftp_server"`
		Logging        loggingConfig        `koanf:"logging"`
		Cluster        clusterConfig        `koanf:"cluster"`
		TimeoutWatcher timeoutWatcherConfig `koanf:"timeout_watcher"`
		Recast         recastConfig         `koanf:"recast"`
		Telemetry      telemetryConfig      `koanf:"telemetry"`
		Staller        stallerConfig        `koanf:"staller"`
	}

	// Server specific configuration
	serverConfig struct {
		// If the http server should be disabled
		Disable bool `koanf:"disable"`

		// Server port to listen on
		Port int `koanf:"port" validate:"required,min=1,max=65535"`

		// Server host to listen on
		Host string `koanf:"host" validate:"required"`

		// Network stack to use (tcp, tcp4, tcp6)
		Network string `koanf:"network" validate:"required,oneof=tcp tcp4 tcp6"`

		// The proxy header to use if the application is behind a proxy
		ProxyHeader string `koanf:"proxy_header" validate:"omitempty"`

		// The list of trusted proxies to use if the application is behind a proxy
		// Must be a list of IP addresses or CIDR ranges
		TrustedProxies []string `koanf:"trusted_proxies" validate:"omitempty,dive,ipv4|ipv6|cidr|cidrv6"`

		// Enable access logs
		AccessLog httpAccessLogConfig `koanf:"access_log"`
	}

	// Config relating to FTP Server File Transfer
	ftpTransferConfig struct {
		// The size of each chunk that makes up each file to be transferred (in bytes)
		ChunkSize int `koanf:"chunk_size" validate:"omitempty,min=1"`

		// How often a chunk of data (In milliseconds)
		ChunkSendRate int `koanf:"chunk_rate" validate:"omitempty,min=1"`

		// The size of each file to be advertised to connected clients (in bytes)
		FileSize int `koanf:"file_size" validate:"omitempty,min=1"`
	}

	ftpThrottleConfig struct {
		// The maximum number of pending operations to throttle
		MaxPendingOperations int `koanf:"max_pending_operations" validate:"required,min=1"`

		// The time to wait before releasing a pending operation
		WaitTime int `koanf:"wait_time" validate:"required,min=1"`
	}

	// Settings relating to the FTP server
	ftpServerConfig struct {
		// If the FTP server should be enabled
		Enabled bool `koanf:"enabled"`

		// The port to listen on N.b this is the control port
		// port 20 is used for data transfer by default in active mode.
		Port int `koanf:"port" validate:"required,min=1,max=65535"`

		// Host to listen on
		Host string `koanf:"host" validate:"required"`

		// Lower bound of ports exposed for passive mode default 50000-50100
		PassivePortRange string `koanf:"passive_port_range" validate:"omitempty,port_range"`

		// The common name for the self signed certificate
		CertCommonName string `koanf:"cert_common_name" validate:"omitempty"`

		// Commands relating to throttling ongoing connections to the ftp server
		Throttle ftpThrottleConfig `koanf:"throttle"`

		// Ftp Transfer Configuration
		Transfer ftpTransferConfig `koanf:"transfer"`

		// Command logging configuration
		CommandLog ftpCommandLogConfig `koanf:"command_log"`
	}

	ftpCommandLogConfig struct {
		// The path to write the access logs to (Otherwise stdout)
		Path string `koanf:"path" validate:"omitempty"`

		// A list of commands to log against each command from the FTP server Context (All commands are logged by default)
		// Note that thease commands are based on calls to an internal emulated filesystem and not the actual FTP commands
		// meaning that though the commands are similar they are not a 1 to 1 mapping ith the FTP protocol
		CommandsToLog []string `koanf:"commands_to_log" validate:"omitempty,dive,oneof=all all_detailed create_file create_directory create_directory_recursive open open_file remove remove_all rename stat chown chtimes close_file read_file read_file_at seek_file write_file write_file_at read_dir read_dir_names stat sync truncate write_string client_connected client_disconnected auth_user none"`

		// Additional fields to log against each command from the FTP server Context
		AdditionalFields []string `koanf:"additional_fields" validate:"omitempty,dive,oneof=id dest_addr src_addr client_version type dest_port src_port src_host dest_host none"`
	}

	// Cluster specific configuration
	clusterConfig struct {
		// If cluster mode is enabled (Nodes will become aware of each other)
		Enabled bool `koanf:"enabled"`

		// The mode for the cluster to use. The modes are as follows:
		// fargate_ecs - This mode is for when the application is running in AWS Fargate
		//               Ip addresses are supplied by the AWS v4 Metadata endpoint so no other IP addresses are needed
		// lan         - This mode is for when the application is running on a local network (See https://github.com/hashicorp/memberlist/blob/8ddac568337bd6b77df1aed75317a52f8b930e83/config.go#L296 for more info)
		// wan         - This mode is for when the application is running on a public network (See https://github.com/hashicorp/memberlist/blob/8ddac568337bd6b77df1aed75317a52f8b930e83/config.go#L340C6-L340C22 for more info)
		Mode string `koanf:"mode" validate:"required_if=Enabled true,omitempty,oneof=fargate_ecs lan wan"`

		// The bind address for the cluster to listen on
		BindPort int `koanf:"bind_port" validate:"required_if=Enabled true,omitempty,min=1,max=65535"`

		// Known Peers
		KnownPeerIps []string `koanf:"known_peer_ips" validate:"required_if=Mode lan Mode wan,omitempty"`

		// Host Ip Address
		// Advertisement IP Address for this node to use against other nodes in the cluster
		AdvertiseIp string `koanf:"advertise_ip" validate:"required_if=Mode lan Mode wan,omitempty,ipv4"`

		// Enable member list logging output
		EnableLogging bool `koanf:"enable_logging"`

		// Number of connection attempts to make to other nodes in the cluster before giving up
		ConnectionAttempts int `koanf:"connection_attempts" validate:"min=1"`

		// The timeout for each connection attempt in seconds
		ConnectionTimeout int `koanf:"connection_timeout_secs" validate:"min=1"`
	}

	// Logging specific configuration
	loggingConfig struct {
		// Logging level
		Level string `koanf:"level"`

		// If the startup log is enabled
		StartUpLogEnabled bool `koanf:"startup_log_enabled"`

		// Sets global log path for protocol specific logs
		Path string `koanf:"path"`
	}

	httpAccessLogConfig struct {
		// The path to write the access logs to (Otherwise stdout)
		Path string `koanf:"path" validate:"omitempty"`

		// The format to write the access logs in given sometimes requests can take a very long time
		// to resolve by design. The following modes are available:
		//    - Start: Only log the start of the request
		//    - End: Only log the end of the request
		//    - Both: Log both the start and end of the request (Each request will be logged twice)
		//    - None: Do not log any requests
		Mode string `koanf:"mode" validate:"omitempty,oneof=start end both none"`

		// The fields to log in the access logs (Note that not all fields are aviailable for all protocols and will be omitted if not present)
		FieldsToLog []string `koanf:"fields_to_log" validate:"omitempty,dive,oneof=timestamp status src_ip method path qs dest_port type host user_agent browser browser_version os os_version device device_brand phase duration id"`
	}

	// Timeout watcher specific configuration
	timeoutWatcherConfig struct {
		// If the timeout watcher is enabled. In the event that this is disabled
		// an infinite timeout will be given to all requests
		Enabled bool `koanf:"enabled"`

		// The number of requests that are allowed before things begin slowing down
		GraceRequests int

		// The TTL (in seconds) for the hot cache pool
		CacheHotPoolTTL int `koanf:"hot_pool_ttl_sec" validate:"omitempty,min=1"`

		// The TTL (in seconds) for the cold cache pool
		CacheColdPoolTTL int `koanf:"cold_pool_ttl_sec" validate:"omitempty,min=1"`

		// The maximum amount of time a given IP can be hanging before we consider the IP
		// to be vulnerable to hanging forever on a request. Any ips that get past this threshold
		// will always be given the longest timeout
		InstantCommitThreshold int `koanf:"instant_commit_threshold_ms" validate:"omitempty,min=1"`

		// The upper bound for increasing timeouts in milliseconds. Once the timeout increases to reach this bound we will hang forever.
		UpperTimeoutBound int `koanf:"upper_timeout_bound_ms" validate:"min=1"`

		// The smallest timeout we will ever give im milliseconds
		LowerTimeoutBound int `koanf:"lower_timeout_bound_ms" validate:"min=1"`

		// The timeout given by requests that are in the grace set of requests in milliseconds
		GraceTimeout int `koanf:"grace_timeout_ms" validate:"omitempty,min=1"`

		// The amount of time to wait when hanging an IP "forever"
		LongestTimeout int `koanf:"longest_timeout_ms" validate:"omitempty,min=1"`

		// The increment we will increase timeouts by for requests with timeouts larger than 30 seconds
		TimeoutOverThirtyIncrement int `koanf:"timeout_over_thirty_increment_ms" validate:"omitempty,min=1"`

		// The increment we will increase timeouts by for requests with timeouts smaller than 30 seconds
		TimeoutSubThirtyIncrement int `koanf:"timeout_sub_thirty_increment_ms" validate:"omitempty,min=1"`

		// The increment we will increase timeouts by for requests with timeouts smaller than 10 seconds
		TimeoutSubTenIncrement int `koanf:"timeout_sub_ten_increment_ms" validate:"omitempty,min=1"`

		// The number of samples to take to detect a timeout
		DetectionSampleSize int `koanf:"sample_size" validate:"omitempty,min=2"`

		// How standard deviation of the last "sample_size" requests to take before committing to a timeout
		DetectionSampleDeviation int `koanf:"sample_deviation_ms" validate:"omitempty,min=1"`
	}

	// Telemetry specific configuration
	telemetryConfig struct {
		// If telemetry is enabled or not
		Enabled bool `koanf:"enabled"`

		// The node name for identifying the said node
		NodeName string `koanf:"node_name" validate:"omitempty"`

		// Using with push gateway
		PushGateway telemetryPushGatewayConfig `koanf:"push_gateway"`

		// Prometheus metrics
		Metrics telemetryMetricsConfig `koanf:"metrics"`

		// Prometheus configuration
		Prometheus telemetryPrometheusConfig `koanf:"prometheus"`
	}

	// Configuration related to the telemetry prometheus
	telemetryPrometheusConfig struct {
		// If the prometheus server is enabled
		Enabled bool `koanf:"enabled"`

		// The port for the prometheus collection endpoint
		Port int `koanf:"prometheus_port" validate:"required,min=1,max=65535"`

		// The path for the prometheus endpoint
		Path string `koanf:"prometheus_path" validate:"required"`
	}

	// Configuration related to the telemetry push gateway
	telemetryPushGatewayConfig struct {
		Enabled bool `koanf:"enabled"`

		// The address of the push gateway
		Endpoint string `koanf:"endpoint" validate:"required_if=Enabled true,omitempty"`

		// The username for the push gateway (For basic auth)
		Username string `koanf:"username" validate:"required_with=Password"`

		// The password for the push gateway (For basic auth)
		Password string `koanf:"password" validate:"required_with=Username"`

		// The interval in seconds to push metrics to the push gateway
		// Default: 60
		PushIntervalSecs int `koanf:"push_interval_secs" validate:"omitempty,min=1"`
	}

	// Configuration related to the prometheus metrics
	telemetryMetricsConfig struct {
		// If prometheus should expose the secrets generated metric
		TrackSecretsGenerated bool `koanf:"track_secrets_generated"`

		// If prometheus should expose the time wasted metric
		TrackTimeWasted bool `koanf:"track_time_wasted"`
	}

	// Configuration related to "recasting" a process in which the node will shutdown in the event that
	// there has been no significant amount of time wasted by the node
	recastConfig struct {
		// If the recast system is enabled or not
		Enabled bool `koanf:"enabled"`

		// The minimum interval in minutes to wait before recasting
		// Default: 30
		MinimumRecastIntervalMin int `koanf:"minimum_recast_interval_min" validate:"omitempty,min=1"`

		// The maximum interval in minutes to wait before recasting
		// Default: 120
		MaximumRecastIntervalMin int `koanf:"maximum_recast_interval_min" validate:"omitempty,min=1"`

		// The ratio of time wasted to time spent. If the ratio is less than this value then the node should recast
		// Default: 0.05
		TimeWastedRatio float64 `koanf:"time_wasted_ratio" validate:"omitempty,min=0,max=1"`
	}

	stallerConfig struct {
		// The maximum number of connections that can be made to the pot at any given time
		MaximumConnections int `koanf:"maximum_connections" validate:"required,min=1"`

		// The maximum number of stallers allowed per group (Normally representing a single connected client)
		// Any connections that exceed this limit will be rejected
		GroupLimit int `koanf:"group_limit" validate:"required,min=1"`

		// The transfer rate for the staller (bytes per second)
		BytesPerSecond int `koanf:"bytes_per_second" validate:"omitempty,min=1"`
	}
)

func NewConfig(cmd *cobra.Command, flagsUsed flagMap) (*Config, error) {

	k := koanf.New(".")

	// Load the default configuration
	if err := k.Load(structs.Provider(defaultConfig, "koanf"), nil); err != nil {
		return nil, err
	}

	if err := loadConfigFile(k, cmd); err != nil {
		return nil, err
	}

	// Override the default configuration with values given by the flags
	k = writeFlagValues(k, cmd, flagsUsed)

	// Write environment variables to the configuration
	err := k.Load(env.ProviderWithValue("GOPOT__", ".", func(s string, v string) (string, interface{}) {
		key := strings.Replace(strings.ToLower(strings.TrimPrefix(s, "GOPOT__")), "__", ".", -1)
		if v == "true" || v == "false" {
			return key, v == "true"
		}

		return key, v
	}), nil)

	if err != nil {
		return nil, err
	}

	// Handle special cases
	setStringSlice(k, "cluster.known_peer_ips")
	setStringSlice(k, "server.trusted_proxies")
	setStringSlice(k, "server.access_log.fields_to_log")
	setStringSlice(k, "ftp_server.command_log.commands_to_log")
	setStringSlice(k, "ftp_server.command_log.additional_fields")

	var cfg *Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		return nil, err
	}

	validator, err := newConfigValidator()

	if err != nil {
		return nil, err
	}

	if err := validator.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Sets the value of a string slice if the value is not empty
func setStringSlice(k *koanf.Koanf, key string) {
	// If we already have a slice value then we don't need to do anything
	sliceVal := k.Strings(key)
	if len(sliceVal) > 0 {
		return
	}

	stringVal := k.String(key)

	if stringVal != "" && stringVal != "[]" {
		if err := k.Set(key, strings.Split(k.String(key), ",")); err != nil {
			return
		}
	} else {
		k.Delete(key)
	}
}
