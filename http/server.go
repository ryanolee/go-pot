package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	//	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/ryanolee/ryan-pot/generator"
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

	randGen, err := generator.NewConfigGenerator()
	if err != nil {
		return err
	}

	pool := NewHttpStallerPool(HttpStallerPoolOptions{
		maximumConnections: 200,
	})
	pool.Start()

	//app.Get("/robots.txt", func(c *fiber.Ctx) error {
	//	g := NewHttpStaller(&HttpStallerOptions{
	//		Generator: gen,
	//	})
	//	return g.StallContextBuffer(c)
	//})

	app.Get("/", func(c *fiber.Ctx) error {
		staller := NewHttpStaller(&HttpStallerOptions{
			Generator:    randGen,
			Request:      c,
			TransferRate: time.Millisecond * 75,
		})
		pool.Register(staller)
		return staller.StallContextBuffer(c)
	})

	return app.Listen(fmt.Sprintf(":%d", cfg.Port))
}
