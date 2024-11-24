package logging

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ua-parser/uap-go/uaparser"
	"go.uber.org/zap"
)

type (
	HttpAccessLogger struct {
		logger      *zap.Logger
		mainLogger  *zap.Logger
		fieldsToLog []string
		loggingMode string
		uaParser    *uaparser.Parser
	}

	HttpAccessLogEntry struct {
		Context        *fiber.Ctx
		Duration       time.Duration
		uaDetails      *uaparser.Client
		phase          string
		resolvedFields map[string]string
	}

	fieldLoggers map[string]func(*HttpAccessLogEntry) string
)

// Lookup table for pulling fields from the http request
var startFieldAccessors fieldLoggers = fieldLoggers{
	// Timestamp and request metadata
	"id": func(entry *HttpAccessLogEntry) string {
		return uuid.New().String()
	},
	"timestamp": func(entry *HttpAccessLogEntry) string {
		return entry.Context.Context().Time().Format(time.RFC3339)
	},
	"status": func(entry *HttpAccessLogEntry) string {
		return strconv.Itoa(entry.Context.Response().StatusCode())
	},
	"src_ip": func(entry *HttpAccessLogEntry) string {
		return entry.Context.IP()
	},
	"method": func(entry *HttpAccessLogEntry) string {
		return entry.Context.Method()
	},
	"path": func(entry *HttpAccessLogEntry) string {
		return entry.Context.Path()
	},
	"qs": func(entry *HttpAccessLogEntry) string {
		return entry.Context.Context().QueryArgs().String()
	},
	"dest_port": func(entry *HttpAccessLogEntry) string {
		host := string(entry.Context.Request().Host())
		_, port, err := net.SplitHostPort(host)
		if err != nil {
			// Set default port based on the scheme (HTTP or HTTPS)
			if strings.EqualFold(entry.Context.Protocol(), "https") {
				port = "443"
			} else {
				port = "80"
			}
		}

		return port
	},
	"type": func(entry *HttpAccessLogEntry) string {
		return "http"
	},
	"host": func(entry *HttpAccessLogEntry) string {
		return string(entry.Context.Request().Host())
	},

	// Parameters derived from the User-Agent header
	"user_agent": func(entry *HttpAccessLogEntry) string {
		return string(entry.Context.Request().Header.UserAgent())
	},
	"browser": func(entry *HttpAccessLogEntry) string {
		if entry.uaDetails != nil {
			return entry.uaDetails.UserAgent.Family
		}
		return ""
	},
	"browser_version": func(entry *HttpAccessLogEntry) string {
		if entry.uaDetails != nil {
			return entry.uaDetails.UserAgent.ToVersionString()
		}
		return ""
	},
	"os": func(entry *HttpAccessLogEntry) string {
		if entry.uaDetails != nil {
			return entry.uaDetails.Os.ToString()
		}
		return ""
	},
	"os_version": func(entry *HttpAccessLogEntry) string {
		if entry.uaDetails != nil {
			return entry.uaDetails.Os.ToVersionString()
		}
		return ""
	},
	"device": func(entry *HttpAccessLogEntry) string {
		if entry.uaDetails != nil {
			return entry.uaDetails.Device.Family
		}
		return ""
	},
	"device_brand": func(entry *HttpAccessLogEntry) string {
		if entry.uaDetails != nil {
			return entry.uaDetails.Device.Brand
		}
		return ""
	},

	// Lifecycle fields
	"phase": func(entry *HttpAccessLogEntry) string {
		return entry.phase
	},
}

// Lookup table for pulling fields from the http response
// with information about the request only available at the end of the request
var endFieldAccessors = fieldLoggers{
	"duration": func(entry *HttpAccessLogEntry) string {
		if entry.Duration == 0 {
			return ""
		}
		return entry.Duration.String()
	},
	"phase": func(entry *HttpAccessLogEntry) string {
		return entry.phase
	},
}

func NewHttpAccessLogger(cfg *config.Config, mainLogger *zap.Logger) (*HttpAccessLogger, error) {
	if cfg.Server.AccessLog.Mode == "none" {
		// Return a no-op logger if access logging is disabled
		return &HttpAccessLogger{
			mainLogger:  mainLogger,
			logger:      zap.NewNop(),
			fieldsToLog: []string{},
			loggingMode: cfg.Server.AccessLog.Mode,
		}, nil
	}

	loggerCfg := zap.NewProductionConfig()

	if cfg.Server.AccessLog.Path != "" {
		loggerCfg.OutputPaths = []string{cfg.Server.AccessLog.Path}
	} else if cfg.Logging.Path != "" {
		loggerCfg.OutputPaths = []string{cfg.Logging.Path}
	} else {
		loggerCfg.OutputPaths = []string{"stdout"}
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		return nil, err
	}

	return &HttpAccessLogger{
		logger:      logger,
		mainLogger:  mainLogger,
		fieldsToLog: cfg.Server.AccessLog.FieldsToLog,
		uaParser:    uaparser.NewFromSaved(),
		loggingMode: cfg.Server.AccessLog.Mode,
	}, nil
}

// Starts logging a http request
func (l *HttpAccessLogger) Start(ctx *fiber.Ctx) *HttpAccessLogEntry {
	if l.loggingMode == "none" {
		return nil
	}

	entry := &HttpAccessLogEntry{
		Context: ctx,
		phase:   "start",
	}

	entry = l.resolveFields(startFieldAccessors, entry)
	if l.loggingMode == "start" || l.loggingMode == "both" {
		l.log(entry)
	}
	return entry
}

// Called when the http request is complete
func (l *HttpAccessLogger) End(entry *HttpAccessLogEntry, duration time.Duration) *HttpAccessLogEntry {
	if l.loggingMode == "none" {
		return nil
	}

	// Update the entry with the duration and log the end entry
	entry.Duration = duration
	entry.phase = "end"

	entry = l.resolveFields(endFieldAccessors, entry)

	if l.loggingMode == "end" || l.loggingMode == "both" {
		l.log(entry)
	}

	return entry
}

func (l *HttpAccessLogger) resolveFields(accessors fieldLoggers, entry *HttpAccessLogEntry) *HttpAccessLogEntry {

	if entry.uaDetails == nil {
		entry.uaDetails = l.uaParser.Parse(string(entry.Context.Request().Header.UserAgent()))
	}

	if entry.resolvedFields == nil {
		entry.resolvedFields = make(map[string]string)
	}

	for _, field := range l.fieldsToLog {
		if accessor, ok := accessors[field]; ok {
			entry.resolvedFields[field] = accessor(entry)
		}
	}

	return entry
}

// Pulls zap fields from the entry
func (l *HttpAccessLogger) getFields(entry *HttpAccessLogEntry) []zap.Field {
	zapFields := make([]zap.Field, 0, len(l.fieldsToLog))
	for _, field := range l.fieldsToLog {
		if resolvedField, ok := entry.resolvedFields[field]; ok {
			zapFields = append(zapFields, zap.String(field, resolvedField))
		} else {
			zapFields = append(zapFields, zap.String(field, ""))
		}
	}

	return zapFields
}

func (l *HttpAccessLogger) log(entry *HttpAccessLogEntry) {
	fields := l.getFields(entry)
	l.logger.Info("", fields...)
}
