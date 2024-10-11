package logging

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/ua-parser/uap-go/uaparser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"strings"
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
            parser := uaparser.NewFromSaved()
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

            // Extract destination port from the request host
            host := string(c.Request().Host())
            _, port, err := net.SplitHostPort(host)
            if err != nil {
                // Set default port based on the scheme (HTTP or HTTPS)
                if strings.EqualFold(c.Protocol(), "https") {
                    port = "443"
                } else {
                    port = "80"
                }
            }

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
                {
                    Key:    "dest_port",
                    Type:   zapcore.StringType,
                    String: port,
                },
            }
        },
    }))

    return s.logger
}
