package logging

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	IServerLogger interface {
		Use(app *fiber.App) *zap.Logger
	}

	ServerLogger struct {
		logger *zap.Logger
	}
)

func NewServerLogger(logger *zap.Logger) *ServerLogger {
	return &ServerLogger{
		logger: logger,
	}
}

func (s *ServerLogger) Use(app *fiber.App) *zap.Logger {
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: s.logger,
		FieldsFunc: func(c *fiber.Ctx) []zapcore.Field {
			userAgent := c.Request().Header.UserAgent()
			length := len(userAgent)
			if length > 128 {
				length = 128
			}
			userAgent = userAgent[:length]
			return []zapcore.Field{
				{
					Key:    "ip",
					Type:   zapcore.StringType,
					String: c.IP(),
				},
				{
					Key:    "user-agent",
					Type:   zapcore.StringType,
					String: string(userAgent),
				},
				{
					Key:    "path",
					Type:   zapcore.StringType,
					String: c.Path(),
				},
				{
					Key:     "status",
					Type:    zapcore.Int8Type,
					Integer: int64(c.Response().StatusCode()),
				},
				{
					Key:    "path",
					Type:   zapcore.StringType,
					String: c.Path(),
				},
			}
		},
	}))

	return s.logger
}
