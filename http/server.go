package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
//	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/ryanolee/ryan-pot/generator"
	"github.com/ryanolee/ryan-pot/rand"
)

type (
	ServerConfig struct {
		Port int
		Debug bool
	}
)

func Serve(cfg ServerConfig) error {
	app := fiber.New()
	//app.Use(logger.New())

	if(cfg.Debug) {
		app.Use(pprof.New())
	}

	rand := rand.NewSeededRand(324234)
	generator, err := generator.NewRobotsTxtGenerator(rand)
	if err != nil {
		return err
	}
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		g := NewHttpStaller(&HttpStallerOptions{
			Generator: generator,
		})
		return g.StallContextBuffer(c)
		//return c.SendString(generator.Generate())
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return nil
	})

	return app.Listen(fmt.Sprintf(":%d", cfg.Port))
}