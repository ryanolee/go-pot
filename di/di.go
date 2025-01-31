package di

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/core/gossip"
	"github.com/ryanolee/go-pot/core/gossip/action"
	"github.com/ryanolee/go-pot/core/gossip/handler"
	"github.com/ryanolee/go-pot/core/logging"
	"github.com/ryanolee/go-pot/core/metrics"
	"github.com/ryanolee/go-pot/core/stall"
	"github.com/ryanolee/go-pot/generator"
	"github.com/ryanolee/go-pot/protocol/detect"
	"github.com/ryanolee/go-pot/protocol/detect/detector"
	"github.com/ryanolee/go-pot/protocol/fallback"
	"github.com/ryanolee/go-pot/protocol/ftp"
	ftpDi "github.com/ryanolee/go-pot/protocol/ftp/di"
	"github.com/ryanolee/go-pot/protocol/ftp/driver"
	ftpLogging "github.com/ryanolee/go-pot/protocol/ftp/logging"
	ftpStall "github.com/ryanolee/go-pot/protocol/ftp/stall"
	"github.com/ryanolee/go-pot/protocol/ftp/throttle"
	"github.com/ryanolee/go-pot/protocol/http"
	httpLogger "github.com/ryanolee/go-pot/protocol/http/logging"
	httpStall "github.com/ryanolee/go-pot/protocol/http/stall"
	"github.com/ryanolee/go-pot/secrets"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"go.uber.org/zap"
)

// Creates the dependency injection container for the application
func CreateContainer(conf *config.Config) *fx.App {

	if !conf.FtpServer.Enabled && !conf.MultiProtocol.Enabled && conf.Server.Disable {
		fmt.Println("All honeypots disabled. There is nothing to do. Exiting.")
		os.Exit(0)
	}

	return fx.New(
		fx.Supply(conf),
		fx.Provide(
			// Logging
			logging.NewLogger,
			httpLogger.NewHttpAccessLogger,
			ftpLogging.NewFtpCommandLogger,

			// Metrics
			metrics.NewTimeoutWatcher,
			metrics.NewTelemetry,
			metrics.NewRecast,

			// Generators
			generator.NewConfigGeneratorCollection,
			secrets.NewSecretGeneratorCollection,

			// Stallers
			stall.NewStallerPool,
			httpStall.NewHttpStallerFactory,
			ftpStall.NewFtpFileStallerFactory,

			// Cluster Memberlist
			fx.Annotate(handler.NewBroadcastActionHandler,
				fx.As(new(handler.IBroadcastActionHandler)),
			),

			fx.Annotate(
				gossip.NewMemberList,
				fx.As(new(action.IBroadcastActionDispatcher)),
				fx.As(new(gossip.IMemberlist)),
			),

			// Detectors
			fx.Annotate(detector.NewHttpDetector, fx.As(new(detector.ProtocolDetector)), fx.ResultTags(`group:"detectors"`)),
			fx.Annotate(detector.NewFtpDetector, fx.As(new(detector.ProtocolDetector)), fx.ResultTags(`group:"detectors"`)),

			// Multi Protocol Listener
			fx.Annotate(detect.NewMulitProtocolListener, fx.ParamTags(``, `group:"detectors"`)),

			// Fallback Server
			fallback.NewFallbackProtocolServer,

			// Http Server
			http.NewServer,
			fx.Annotate(
				httpLogger.NewServerLogger,
				fx.As(new(httpLogger.IServerLogger)),
			),

			// Ftp Support
			throttle.NewFtpThrottle,

			// Ftp Server
			ftp.NewServer,
			driver.NewFtpServerDriver,
			driver.NewFtpClientDriverFactory,

			// Di Repositories
			ftpDi.NewFtpRepository,
		),

		// Set Global Logger with the one created in the container
		fx.Invoke(func(logger *zap.Logger) {
			zap.ReplaceGlobals(logger)
		}),

		// Resolve circular dependencies
		fx.Invoke(func(config *config.Config, watcher *metrics.TimeoutWatcher, dispatcher action.IBroadcastActionDispatcher) {
			if config.Cluster.Enabled && config.TimeoutWatcher.Enabled {
				watcher.SetActionDispatcher(dispatcher)
			}
		}),

		// Shutdown hook
		fx.Invoke(func(shutdown fx.Shutdowner, pool *stall.StallerPool, logger *zap.Logger) {
			go func() {
				shutdownChannel := make(chan os.Signal, 1)

				signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

				<-shutdownChannel

				logger.Warn("Shutting down...")

				// Stop pool before normal FX lifecycle hook so that all active connections are closed
				logger.Info("Flushing stall pool")
				pool.Stop()

				err := shutdown.Shutdown()
				if err != nil {
					logger.Sugar().Fatalf("Error shutting down, Forcing shutdown", zap.Error(err))
				}

				time.Sleep(time.Second * 30)
				logger.Error("Deadline has passed after 30 seconds, Forcing shutdown.")
				os.Exit(0)
			}()
		}),

		// Start HTTP server
		fx.Invoke(func(c *config.Config, s *http.Server, logger *zap.Logger, mlp *detect.MultiProtocolListener) {
			logger.Info("HTTP Server Enabled: ", zap.Bool("enabled", !conf.Server.Disable))
			if conf.Server.Disable && !mlp.ProtocolEnabled("http") {
				logger.Info("Http is disabled")
				return
			}
			logger.Info("Starting Http server",
				zap.Int("port", s.ListenPort),
				zap.String("host", s.ListenHost),
				zap.Bool("managed_by_multi_protocol_listener", mlp.ProtocolEnabled("http")),
			)
			go func() {
				if err := s.Start(); err != nil {
					logger.Fatal("Failed to start Http server", zap.Error(err))
					os.Exit(1)
				}
			}()
		}),

		// Start Ftp server
		fx.Invoke(func(c *config.Config, s *ftpserver.FtpServer, logger *zap.Logger, mlp *detect.MultiProtocolListener) {
			if !conf.FtpServer.Enabled && !mlp.ProtocolEnabled("ftp") {
				logger.Info("Ftp is disabled")
				return
			}

			logger.Info("Starting Ftp server",
				zap.Int("port", c.FtpServer.Port),
				zap.String("host", c.FtpServer.Host),
				zap.String("passive_port_range", c.FtpServer.PassivePortRange),
				zap.Bool("managed_by_multi_protocol_listener", mlp.ProtocolEnabled("ftp")),
			)

			go func() {
				if err := s.ListenAndServe(); err != nil {
					logger.Sugar().Warn("Failed to start Ftp server", "error", err)
				}
			}()
		}),

		// Start multi protocol listener
		fx.Invoke(func(ls *detect.MultiProtocolListener, conf *config.Config, logger *zap.Logger) {
			if ls == nil {
				return
			}

			logger.Info("Starting Multi Protocol Listener", zap.Int("port", conf.MultiProtocol.Port), zap.String("host", conf.MultiProtocol.Host))
			go func() {
				if err := ls.Listen(); err != nil {
					logger.Sugar().Fatalf("Failed to start Multi Protocol Listener", zap.Error(err))
				}
			}()
		}),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			if !conf.Logging.StartUpLogEnabled {
				return &fxevent.ZapLogger{Logger: zap.NewNop()}
			}
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
}
