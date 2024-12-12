package stall

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/core/metrics"
	"github.com/ryanolee/go-pot/core/stall"
	"github.com/ryanolee/go-pot/generator"
	"github.com/ryanolee/go-pot/generator/encoder"
	"github.com/ryanolee/go-pot/protocol/http/logging"
	"github.com/ryanolee/go-pot/secrets"
)

type HttpStallerFactory struct {
	// Services
	pool              *stall.StallerPool
	telemetry         *metrics.Telemetry
	timeoutWatcher    *metrics.TimeoutWatcher
	secretsGenerators *secrets.SecretGeneratorCollection
	configGenerators  *generator.ConfigGeneratorCollection

	// Logger
	logger *logging.HttpAccessLogger

	// Config
	bytesPerSecond int
}

func NewHttpStallerFactory(
	config *config.Config,
	pool *stall.StallerPool,
	timeoutWatcher *metrics.TimeoutWatcher,
	telemetry *metrics.Telemetry,
	secretsGeneratorCollection *secrets.SecretGeneratorCollection,
	configGeneratorCollection *generator.ConfigGeneratorCollection,
	logger *logging.HttpAccessLogger,
) *HttpStallerFactory {
	return &HttpStallerFactory{
		pool:              pool,
		telemetry:         telemetry,
		timeoutWatcher:    timeoutWatcher,
		secretsGenerators: secretsGeneratorCollection,
		configGenerators:  configGeneratorCollection,
		logger:            logger,

		bytesPerSecond: config.Staller.BytesPerSecond,
	}
}

func (f *HttpStallerFactory) FromFiberContext(c *fiber.Ctx) (*HttpStaller, error) {
	entry := f.logger.Start(c)

	encoderInstance := encoder.GetEncoderForPath(c.Path())
	gen := generator.GetGeneratorForEncoder(encoderInstance, f.configGenerators, f.secretsGenerators)
	ip := c.IP()
	identifier := "http-" + ip
	opts := &HttpStallerOptions{
		Request:      c,
		Generator:    gen,
		TransferRate: time.Second / time.Duration(f.bytesPerSecond),
		Timeout:      f.timeoutWatcher.GetTimeout(identifier),
		ContentType:  encoderInstance.ContentType(),
		OnTimeout: func(stl *HttpStaller) {
			f.logger.End(entry, stl.GetElapsedTime())
			f.timeoutWatcher.RecordResponse(identifier, stl.GetElapsedTime(), false)
		},
		OnClose: func(stl *HttpStaller) {
			f.logger.End(entry, stl.GetElapsedTime())
			f.timeoutWatcher.RecordResponse(identifier, stl.GetElapsedTime(), true)
		},
		Telemetry: f.telemetry,
	}
	staller := NewHttpStaller(opts)
	if err := f.pool.Register(staller); err != nil {
		return nil, err
	}

	return staller, nil
}
