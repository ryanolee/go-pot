package config

import (
	validate "github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap/zapcore"
)



type (
    // This struct covers the entire application configuration
	Config struct {
		HttpServer serverConfig `koanf:"http_server"`
        Logging loggingConfig `koanf:"logging"`
        Cluster clusterConfig `koanf:"cluster"`
        TimeoutWatcher timeoutWatcherConfig `koanf:"timeout_watcher"`
        Recast recastConfig `koanf:"recast"`
        Telemetry telemetryConfig `koanf:"telemetry"`
        Staller httpStallerConfig `koanf:"staller"`
    }

    // Server specific configuration
   	serverConfig struct {
		// Server port to listen on
		Port int `koanf:"port" validate:"required,min=1,max=65535"`

        // Server host to listen on
		Host string `koanf:"host" validate:"required"`

        // Debug mode
		Debug bool `koanf:"debug"`
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
        Mode string `koanf:"mode" validate:"required_if=Enabled true,omitempty,oneof='fargate_ecs lan wan'"`

        // The bind address for the cluster to listen on
        BindPort int `koanf:"bind_port" validate:"required_if=Enabled true,omitempty,min=1,max=65535"`

        // Known Peers
        KnownPeerIps []string `koanf:"known_peers" validate:"required_if=Mode lan Mode wan,omitempty,dive,ipv4"`

        // Host Ip Address
        // Advertisement IP Address for this node to use against other nodes in the cluster
        AdvertiseIp string `koanf:"advertise_ip" validate:"required_if=Mode lan Mode wan,omitempty,ipv4"`

        // Enable member list logging output
        EnableLogging bool `koanf:"enable_logging"`

        // Number of connection attempts to make to other nodes in the cluster before giving up
        ConnectionAttempts int `koanf:"connection_attempts" validate:"min=1"`

        // The timeout for each connection attempt in seconds
        ConnectionTimeout int `koanf:"connection_timeout" validate:"min=1"`

    }

    // Logging specific configuration
    loggingConfig struct {
        // Logging level
        Level string `koanf:"level"`
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
        DetectionSampleSize 	int `koanf:"sample_size" validate:"omitempty,min=2"`

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
        Prometheus telemetryPrometheus `koanf:"prometheus"`
    }

    // Configuration related to the telemetry prometheus
    telemetryPrometheus struct {
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

    httpStallerConfig struct {
        // The maximum number of connections that can be made to the pot at any given time
        MaximumConnections int `koanf:"maximum_connections" validate:"required,min=1"`

        // The transfer rate for the staller (bytes per second)
        BytesPerSecond int `koanf:"bytes_per_second" validate:"omitempty,min=1"`
    }
)

var defaultConfig = Config{
	HttpServer: serverConfig{
		Port: 8080,
		Host: "127.0.0.1",
		Debug: false,
	},
    Logging: loggingConfig {
        Level: zapcore.InfoLevel.String(),
    },
    Cluster: clusterConfig {
        Enabled: false,
        ConnectionTimeout: 5,
        ConnectionAttempts: 5,
        EnableLogging: false,
        BindPort: 7946,
    },
    TimeoutWatcher: timeoutWatcherConfig {
        Enabled: true,
        
        // Threaholds
        InstantCommitThreshold: 180 * 1000, // 3 minutes
        UpperTimeoutBound: 60 * 1000, // 1 minute
        LowerTimeoutBound:  1000, // 1 second

        // Grace Periods
        GraceRequests: 3,
        GraceTimeout:  100, // 100ms

        // Timeouts
        LongestTimeout: 7 * 24 * 60 * 60 * 1000, // 7 days

        // Increments
        TimeoutOverThirtyIncrement: 10 * 1000, // 10 seconds
        TimeoutSubThirtyIncrement: 5 * 1000, // 5 seconds
        TimeoutSubTenIncrement: 2 * 1000, // 1 second

        // Detection
        DetectionSampleSize: 3,
        DetectionSampleDeviation: 1000, // 1 second

        // Cache TTLs
        CacheHotPoolTTL: 60 * 60, // 1 hour
        CacheColdPoolTTL: 60 * 60 * 48, // 2 days
    },
    Telemetry: telemetryConfig {
        Enabled: true,
        NodeName: "go-pot",
        PushGateway: telemetryPushGatewayConfig{
            Enabled: false,
            PushIntervalSecs: 60,
            Endpoint: "",
            Username: "",
            Password: "",
        },
        Prometheus: telemetryPrometheus{
            Enabled: true,
            Port: 9001,
            Path: "/metrics",
        },
        Metrics: telemetryMetricsConfig{
            TrackSecretsGenerated: true,
            TrackTimeWasted: true,
        },
    },
    Recast: recastConfig{
        Enabled: false,
        MinimumRecastIntervalMin: 30,
        MaximumRecastIntervalMin: 120,
        TimeWastedRatio: 0.05,
    },
    Staller: httpStallerConfig{
        MaximumConnections: 200,
        BytesPerSecond: 8,
    },
}


func NewConfig () (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(structs.Provider(defaultConfig, "koanf"), nil); err != nil {
		return nil, err
	}

	var cfg *Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		return nil, err
	}
    validator := validate.New()
    if err := validator.Struct(cfg); err != nil {
        return nil, err
    }

	return cfg, nil
}