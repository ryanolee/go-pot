package di

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/core/gossip"
	"github.com/ryanolee/ryan-pot/core/gossip/action"
	"github.com/ryanolee/ryan-pot/core/gossip/handler"
	"github.com/ryanolee/ryan-pot/core/logging"
	"github.com/ryanolee/ryan-pot/core/metrics"
	"github.com/ryanolee/ryan-pot/core/stall"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/protocol/ftp"
	ftpDi "github.com/ryanolee/ryan-pot/protocol/ftp/di"
	"github.com/ryanolee/ryan-pot/protocol/ftp/driver"
	ftpLogging "github.com/ryanolee/ryan-pot/protocol/ftp/logging"
	ftpStall "github.com/ryanolee/ryan-pot/protocol/ftp/stall"
	"github.com/ryanolee/ryan-pot/protocol/ftp/throttle"
	"github.com/ryanolee/ryan-pot/protocol/http"
	httpLogger "github.com/ryanolee/ryan-pot/protocol/http/logging"
	httpStall "github.com/ryanolee/ryan-pot/protocol/http/stall"
	"github.com/ryanolee/ryan-pot/secrets"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	//"go.uber.org/fx/fxevent"
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

				shutdown.Shutdown()
				time.Sleep(time.Second * 30)
				os.Exit(0)
			}()
		}),

		// Start HTTP server
		fx.Invoke(func(c *config.Config, s *http.Server) {
			if conf.Server.Disable {
				zap.L().Info("Http is disabled")
				return
			}
			zap.L().Info("Starting Http server", zap.Int("port", s.ListenPort), zap.String("host", s.ListenHost))
			go func() {
				if err := s.Start(); err != nil {
					zap.L().Sugar().Fatalf("Failed to start Http server", "error", err)
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
