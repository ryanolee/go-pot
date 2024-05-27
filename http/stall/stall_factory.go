package stall

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/http/encoder"
	"github.com/ryanolee/ryan-pot/http/metrics"
	"github.com/ryanolee/ryan-pot/secrets"
	"go.uber.org/zap"
)

type HttpStallerFactory struct {
	// Services
	pool *HttpStallerPool
	telemetry *metrics.Telemetry
	timeoutWatcher *metrics.TimeoutWatcher
	secretsGenerators *secrets.SecretGeneratorCollection
	configGenerators *generator.ConfigGeneratorCollection

	// Config
	bytesPerSecond int
}

func NewHttpStallerFactory(
	config *config.Config,
	pool *HttpStallerPool,
	timeoutWatcher *metrics.TimeoutWatcher,
	telemetry *metrics.Telemetry,
	secretsGeneratorCollection *secrets.SecretGeneratorCollection,
	configGeneratorCollection *generator.ConfigGeneratorCollection,
) *HttpStallerFactory {
	return &HttpStallerFactory{
		pool: pool,
		telemetry: telemetry,
		timeoutWatcher: timeoutWatcher,
		secretsGenerators: secretsGeneratorCollection,
		configGenerators: configGeneratorCollection,

		bytesPerSecond: config.Staller.BytesPerSecond,
	}
}

func (f *HttpStallerFactory) FromFiberContext(c *fiber.Ctx) (*HttpStaller, error) {
	encoder := encoder.GetEncoderForPath(c.Path())
	gen := generator.NewConfigGenerator(encoder, f.configGenerators, f.secretsGenerators)
	ip := c.IP()
	opts := &HttpStallerOptions{
		Request: c,
		Generator: gen,
		TransferRate: time.Second / time.Duration(f.bytesPerSecond),
		Timeout: f.timeoutWatcher.GetTimeout(ip),
		ContentType: encoder.ContentType(),
		OnTimeout: func(stl *HttpStaller) {
			zap.L().Sugar().Infow("Timeout", "ip", ip, "duration", stl.GetElapsedTime())
			f.timeoutWatcher.RecordResponse(ip, stl.GetElapsedTime(), false)
		},
		OnClose: func(stl *HttpStaller) {
			zap.L().Sugar().Infow("Timeout", "ip", ip, "duration", stl.GetElapsedTime())
			f.timeoutWatcher.RecordResponse(ip, stl.GetElapsedTime(), true)
		},
		Telemetry: f.telemetry,
	}
	staller := NewHttpStaller(opts)
	if err := f.pool.Register(staller); err != nil {
		return nil, err
	}

	return staller, nil
}

