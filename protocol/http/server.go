package http

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/protocol/detect"
	"github.com/ryanolee/go-pot/protocol/http/logging"
	"github.com/ryanolee/go-pot/protocol/http/stall"
)

type (
	Server struct {
		App        *fiber.App
		ListenPort int
		ListenHost string

		// Custom listener for the server (Overriding the default listener)
		Listener       net.Listener
		stallerFactory *stall.HttpStallerFactory
	}
)

func NewServer(
	lf fx.Lifecycle,
	shutdown fx.Shutdowner,
	cfg *config.Config,
	logging logging.IServerLogger,
	stallerFactory *stall.HttpStallerFactory,
	ls *detect.MultiProtocolListener,
	logger *zap.Logger,
) *Server {
	// Only enable the trusted proxy check if we have trusted proxies
	trustedProxyCheck := len(cfg.Server.TrustedProxies) > 0

	var listener net.Listener = nil
	if ls.ProtocolEnabled("http") {
		listener = ls.GetListenerForProtocol("http")
	}

	server := &Server{
		App: fiber.New(fiber.Config{
			IdleTimeout:             time.Second * 15,
			ReduceMemoryUsage:       true,
			DisableStartupMessage:   true,
			Network:                 cfg.Server.Network,
			EnableIPValidation:      true,
			ProxyHeader:             cfg.Server.ProxyHeader,
			TrustedProxies:          cfg.Server.TrustedProxies,
			EnableTrustedProxyCheck: trustedProxyCheck,
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				// All is always ok even if we have an error. Just log it and return an empty response
				logger.Error("Error in request", zap.Error(err))
				return c.Status(fiber.StatusOK).SendString("{}")
			},
		}),

		ListenPort: cfg.Server.Port,
		ListenHost: cfg.Server.Host,

		Listener: listener,

		stallerFactory: stallerFactory,
	}

	lf.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Sugar().Info("Http Shutting down server")
			return server.App.ShutdownWithTimeout(time.Second * 5)
		},
	})

	return server
}

func (s *Server) Start() error {
	// Setup routes
	s.App.Get("/robots.txt", func(c *fiber.Ctx) error {
		return c.SendString("User-agent: *\nDisallow: /")
	})

	s.App.Get("/*", func(c *fiber.Ctx) error {
		staller, err := s.stallerFactory.FromFiberContext(c)
		if err != nil {
			return err
		}

		// Set the correct content type based on the context
		c.Response().Header.SetContentType(staller.GetContentType())

		return staller.StallContextBuffer(c)
	})

	if s.Listener != nil {
		s.App.Listener(s.Listener)
		return s.Start()
	}

	return s.App.Listen(fmt.Sprintf("%s:%d", s.ListenHost, s.ListenPort))
}
