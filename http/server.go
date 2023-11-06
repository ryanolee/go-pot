package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/http/encoder"
	"github.com/ryanolee/ryan-pot/http/gossip"
	"github.com/ryanolee/ryan-pot/http/logging"
	"github.com/ryanolee/ryan-pot/http/metrics"
	"github.com/ryanolee/ryan-pot/http/stall"
	"github.com/ryanolee/ryan-pot/secrets"
)

type (
	ServerConfig struct {
		Port  int
		Debug bool
	}
)

func Serve(cfg ServerConfig) error {
	// Setup server
	app := fiber.New(fiber.Config{
		IdleTimeout:       time.Second * 15,
		ReduceMemoryUsage: true,
	})

	// Setup logging
	logger := logging.UseLogger(app)
	zap.ReplaceGlobals(logger)

	// Initialize generators
	confGenerators, err := generator.NewConfigGeneratorCollection()
	secretGenerators := secrets.NewSecretGeneratorCollection()

	if err != nil {
		panic(err)
	}

	// Connection Pool
	pool := stall.NewHttpStallerPool(stall.HttpStallerPoolOptions{
		MaximumConnections: 200,
	})
	pool.Start()

	// Timeout watcher
	watcher := metrics.NewTimeoutWatcher()

	// Meberlist setup
	actionHandler := gossip.NewBroadcastActionHandler(watcher)
	client, err := gossip.NewMemberList(&gossip.MemberlistOpts{
		OnAction:     actionHandler.Handle,
		SuppressLogs: true,
	})

	if err != nil {
		zap.L().Sugar().Errorw("Failed to create memberlist client", "error", err)
	} else {
		watcher.BindActionDispatcher(client)
	}

	// Setup routes
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		return c.SendString("User-agent: *\nDisallow: /")
	})

	app.Get("/*", func(c *fiber.Ctx) error {
		encoder := encoder.GetEncoderForPath(c.Path())
		c.Response().Header.SetContentType(encoder.ContentType())
		generator := generator.NewConfigGenerator(encoder, confGenerators, secretGenerators)
		ipAddress := c.IP()
		timeout := watcher.GetTimeout(ipAddress)

		staller := stall.NewHttpStaller(&stall.HttpStallerOptions{
			Generator:    generator,
			Request:      c,
			TransferRate: time.Millisecond * 75,
			Timeout:      timeout,
			OnTimeout: func(s *stall.HttpStaller) {
				logger.Sugar().Infow("Timeout", "ip", ipAddress, "duration", s.GetElapsedTime())
				watcher.RecordResponse(ipAddress, s.GetElapsedTime(), false)
			},
			OnClose: func(s *stall.HttpStaller) {
				logger.Sugar().Infow("Timeout", "ip", ipAddress, "duration", s.GetElapsedTime())
				watcher.RecordResponse(ipAddress, s.GetElapsedTime(), true)
			},
		})
		err := pool.Register(staller)
		if err != nil {
			return err
		}

		return staller.StallContextBuffer(c)
	})

	return app.Listen(fmt.Sprintf(":%d", cfg.Port))
}
