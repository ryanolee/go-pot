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

	if !conf.FtpServer.Enabled && conf.Server.Disable {
		fmt.Print("Both FTP and HTTP servers are disabled. There is nothing to do. Exiting.")
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
		fx.Invoke(func(shutdown fx.Shutdowner) {
			go func() {
				shutdownChannel := make(chan os.Signal, 1)

				signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

				<-shutdownChannel
				zap.L().Info("Shutting down...")
				err := shutdown.Shutdown()
				if err != nil {
					zap.L().Sugar().Fatalf("Error shutting down, Forcing shutdown", zap.Error(err))
				}

				time.Sleep(time.Second * 30)
				zap.L().Info("Deadline has passed after 30 seconds, Forcing shutdown.")
				zap.L().Sync()
				os.Exit(0)
			}()
		}),

		// Start HTTP server
		fx.Invoke(func(c *config.Config, s *http.Server) {
			zap.L().Info("HTTP Server Enabled: ", zap.Bool("enabled", !conf.Server.Disable))
			if conf.Server.Disable {
				zap.L().Info("Http is disabled")
				return
			}
			zap.L().Info("Starting Http server", zap.Int("port", s.ListenPort), zap.String("host", s.ListenHost))
			go func() {
				if err := s.Start(); err != nil {
					zap.L().Fatal("Failed to start Http server", zap.Error(err))
				}
			}()
		}),

		// Start Ftp server
		fx.Invoke(func(c *config.Config, s *ftpserver.FtpServer) {
			if !conf.FtpServer.Enabled {
				zap.L().Info("Ftp is disabled")
				return
			}
			zap.L().Info("Starting Ftp server", zap.Int("port", c.FtpServer.Port), zap.String("host", c.FtpServer.Host), zap.String("passive_port_range", c.FtpServer.PassivePortRange))
			go func() {
				if err := s.ListenAndServe(); err != nil {
					zap.L().Sugar().Fatalf("Failed to start Ftp server", "error", err)
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
