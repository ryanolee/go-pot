package logging

import (
	"github.com/ryanolee/ryan-pot/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	loggerCfg := zap.NewProductionConfig()

	loggerCfg.Level.UnmarshalText([]byte(cfg.Logging.Level))

	// Add file logging
	loggerCfg.OutputPaths = []string{
		"/opt/go-pot/log/go-pot.json", // Path to your log file
		"stderr",           // Also output to stderr (optional)
	}

	// Modify the encoder configuration to change the timestamp format
	loggerCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339)) // ISO 8601 format
	})

	// Change the field name for the timestamp
	loggerCfg.EncoderConfig.TimeKey = "timestamp"

	// Build the logger
	return loggerCfg.Build()
}
