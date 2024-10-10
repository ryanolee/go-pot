package logging

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/ua-parser/uap-go/uaparser" // Updated User-Agent parsing library
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

            // Parse the User-Agent string using ua-parser
            parser := uaparser.NewFromSaved() // Creates a new parser instance
            client := parser.Parse(string(userAgent))

            // Extract browser, OS, and device information
            browserName := client.UserAgent.Family
            browserVersion := client.UserAgent.ToVersionString()
            os := client.Os.ToString()
            device := client.Device.Family

            length := len(userAgent)
            if length > 128 {
                length = 128
            }
            userAgent = userAgent[:length]

            return []zapcore.Field{
                {
                    Key:    "src_ip",
                    Type:   zapcore.StringType,
                    String: c.IP(),
                },
                {
                    Key:    "user-agent",
                    Type:   zapcore.StringType,
                    String: string(userAgent),
                },
                {
                    Key:    "browser",
                    Type:   zapcore.StringType,
                    String: browserName,
                },
                {
                    Key:    "browser-version",
                    Type:   zapcore.StringType,
                    String: browserVersion,
                },
                {
                    Key:    "os",
                    Type:   zapcore.StringType,
                    String: os,
                },
                {
                    Key:    "device",
                    Type:   zapcore.StringType,
                    String: device,
                },
                {
                    Key:    "path",
                    Type:   zapcore.StringType,
                    String: c.Path(),
                },
                // Removed duplicate custom status field
            }
        },
    }))

    return s.logger
}
