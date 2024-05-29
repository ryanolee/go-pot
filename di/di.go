package di

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/http"
	"github.com/ryanolee/ryan-pot/http/gossip"
	"github.com/ryanolee/ryan-pot/http/gossip/action"
	"github.com/ryanolee/ryan-pot/http/gossip/handler"
	"github.com/ryanolee/ryan-pot/http/logging"
	"github.com/ryanolee/ryan-pot/http/metrics"
	"github.com/ryanolee/ryan-pot/http/stall"
	"github.com/ryanolee/ryan-pot/secrets"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// Creates the dependency injection container for the application
func CreateContainer(conf *config.Config) *fx.App {
	return fx.New(
		fx.Supply(conf),
		fx.Provide(
			// Logging
			logging.NewLogger,

			// Metrics
			metrics.NewTimeoutWatcher,
			metrics.NewTelemetry,
			metrics.NewRecast,

			// Generators
			generator.NewConfigGeneratorCollection,
			secrets.NewSecretGeneratorCollection,

			// Stallers
			stall.NewHttpStallerPool,
			stall.NewHttpStallerFactory,

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
				logging.NewServerLogger,
				fx.As(new(logging.IServerLogger)),
			),
		),
		// Resolve circular dependencies
		fx.Invoke(func(config *config.Config, watcher *metrics.TimeoutWatcher, dispatcher action.IBroadcastActionDispatcher) {
			if(config.Cluster.Enabled && config.TimeoutWatcher.Enabled) {
				watcher.SetActionDispatcher(dispatcher)
			}
		}),

		fx.Invoke(func(shutdown fx.Shutdowner){
			go func() {
				shutdownChannel := make(chan os.Signal, 1)
		
				signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		
				<-shutdownChannel
				
				shutdown.Shutdown()
				time.Sleep(time.Second * 30)
				os.Exit(0)
			}()
		}),

		fx.Invoke(func(s *http.Server) {
			zap.L().Info("Starting server", zap.Int("port", s.ListenPort),  zap.String("host", s.ListenHost))
			go s.Start()
		}),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
}
