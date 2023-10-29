package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/http/encoder"
	"github.com/ryanolee/ryan-pot/http/stall"
	"github.com/ryanolee/ryan-pot/secrets"
)

type (
	ServerConfig struct {
		Port  int
		Debug bool
	}
)

func Serve(cfg ServerConfig) error {
	app := fiber.New(fiber.Config{
		IdleTimeout:       time.Second * 15,
		ReduceMemoryUsage: true,
	})
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
	//app.Use(logger.New())

	if cfg.Debug {
		app.Use(pprof.New())
	}

	//rand := rand.NewSeededRand(324234)
	//gen, err := generator.NewRobotsTxtGenerator(rand)
	//if err != nil {
	//	return err
	//}

	confGenerators, err := generator.NewConfigGeneratorCollection()
	if err != nil {
		return err
	}

	// Connection Pool
	pool := stall.NewHttpStallerPool(stall.HttpStallerPoolOptions{
		MaximumConnections: 200,
	})
	pool.Start()

	// Secret Generators
	secretGenerators := secrets.NewSecretGeneratorCollection()

	//app.Get("/robots.txt", func(c *fiber.Ctx) error {
	//	g := NewHttpStaller(&HttpStallerOptions{
	//		Generator: gen,
	//	})
	//	return g.StallContextBuffer(c)
	//})

	app.Get("/*", func(c *fiber.Ctx) error {
		encoder := encoder.GetEncoderForPath(c.Path())
		c.Response().Header.SetContentType(encoder.ContentType())
		generator := generator.NewConfigGenerator(encoder, confGenerators, secretGenerators)
		staller := stall.NewHttpStaller(&stall.HttpStallerOptions{
			Generator:    generator,
			Request:      c,
			TransferRate: time.Millisecond * 1,
		})
		err := pool.Register(staller)
		if err != nil {
			return err
		}

		return staller.StallContextBuffer(c)
	})

	return app.Listen(fmt.Sprintf(":%d", cfg.Port))
}
