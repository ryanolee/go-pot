package logging

import (
	"github.com/ryanolee/ryan-pot/config"
	"go.uber.org/zap"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.Level.UnmarshalText([]byte(cfg.Logging.Level))

	return loggerCfg.Build()
}