package config

import "go.uber.org/zap/zapcore"

// Default configuration values for the application
var defaultConfig = Config{
	MultiProtocol: multiProtocolConfig{
		Enabled:   false,
		Host:      "0.0.0.0",
		Protocols: []string{"all"},
		Port:      8081,
	},
	Server: serverConfig{
		Disable:        false,
		Port:           8080,
		Host:           "127.0.0.1",
		Network:        "tcp4",
		ProxyHeader:    "X-Forwarded-For",
		TrustedProxies: []string{},
		AccessLog: httpAccessLogConfig{
			Mode: "end",
			FieldsToLog: []string{
				"src_ip",
				"method",
				"path",
				"qs",
				"duration",
			},
		},
	},
	FtpServer: ftpServerConfig{
		Enabled:          false,
		Port:             2121,
		Host:             "0.0.0.0",
		PassivePortRange: "50000-50100",
		CertCommonName:   "unknown",
		Throttle: ftpThrottleConfig{
			WaitTime:             1000,
			MaxPendingOperations: 10,
		},
		Transfer: ftpTransferConfig{
			ChunkSize:     1,
			ChunkSendRate: 1000,
			FileSize:      1024 * 1024 * 20, // 20Mb
		},
		CommandLog: ftpCommandLogConfig{
			CommandsToLog: []string{
				"all",
			},
			AdditionalFields: []string{
				"id",
			},
		},
	},
	Logging: loggingConfig{
		Level:             zapcore.InfoLevel.String(),
		StartUpLogEnabled: true,
	},
	Cluster: clusterConfig{
		Enabled:            false,
		Mode:               "lan",
		ConnectionTimeout:  5,
		ConnectionAttempts: 5,
		EnableLogging:      false,
		BindPort:           7946,
	},
	TimeoutWatcher: timeoutWatcherConfig{
		Enabled: true,

		// Threaholds
		InstantCommitThreshold: 180 * 1000, // 3 minutes
		UpperTimeoutBound:      60 * 1000,  // 1 minute
		LowerTimeoutBound:      1000,       // 1 second

		// Grace Periods
		GraceRequests: 3,
		GraceTimeout:  100, // 100ms

		// Timeouts
		LongestTimeout: 7 * 24 * 60 * 60 * 1000, // 7 days

		// Increments
		TimeoutOverThirtyIncrement: 10 * 1000, // 10 seconds
		TimeoutSubThirtyIncrement:  5 * 1000,  // 5 seconds
		TimeoutSubTenIncrement:     2 * 1000,  // 1 second

		// Detection
		DetectionSampleSize:      3,
		DetectionSampleDeviation: 1000, // 1 second

		// Cache TTLs
		CacheHotPoolTTL:  60 * 60,      // 1 hour
		CacheColdPoolTTL: 60 * 60 * 48, // 2 days
	},
	Telemetry: telemetryConfig{
		Enabled:  true,
		NodeName: "go-pot",
		PushGateway: telemetryPushGatewayConfig{
			Enabled:          false,
			PushIntervalSecs: 60,
			Endpoint:         "",
			Username:         "",
			Password:         "",
		},
		Prometheus: telemetryPrometheusConfig{
			Enabled: false,
			Port:    9001,
			Path:    "/metrics",
		},
		Metrics: telemetryMetricsConfig{
			TrackSecretsGenerated: true,
			TrackTimeWasted:       true,
		},
	},
	Recast: recastConfig{
		Enabled:                  false,
		MinimumRecastIntervalMin: 30,
		MaximumRecastIntervalMin: 120,
		TimeWastedRatio:          0.05,
	},
	Staller: stallerConfig{
		MaximumConnections: 200,
		GroupLimit:         50,
		BytesPerSecond:     8,
	},
}
