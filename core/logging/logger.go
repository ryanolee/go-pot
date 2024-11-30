package logging

import (
	"github.com/ryanolee/ryan-pot/config"
	"go.uber.org/zap"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	loggerCfg := zap.NewProductionConfig()
	if err := loggerCfg.Level.UnmarshalText([]byte(cfg.Logging.Level)); err != nil {
		return nil, err
	}

	return loggerCfg.Build()
}
