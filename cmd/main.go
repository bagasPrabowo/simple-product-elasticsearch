package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"simple-product-elasticsearch/internal/common"
	"simple-product-elasticsearch/internal/config"
	"simple-product-elasticsearch/internal/rest"
	"time"

	_ "simple-product-elasticsearch/docs"

	"github.com/gofiber/fiber/v3"
	fiberlog "github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

//	@title			Elasticsearch Example
//	@version		1.0
//	@description	This is a sample Simple Elasticsearch Product.
//	@host			localhost:8080
//	@BasePath		/
//	@accept			json
//	@produce		json
//	@schemes		http
func main() {
	cfg := config.Init()
	es := config.InitES(cfg)

	// create new fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			// Retrive the error code from fiber
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Set the status code
			c.Status(code)

			// Return a JSON response with the default error message
			return c.JSON(common.ErrorResponse{
				Error: err.Error(),
			})
		},
	})

	app.Use(logger.New(logger.Config{}), recover.New())
	rest.RegisterRoute(cfg, app, es)

	// Create a channel to listen for OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Start the server
	go func() {
		port := fmt.Sprintf(":%s", cfg.ServerPort)
		if err := app.Listen(port); err != nil {
			fiberlog.Errorf("Error starting server: %s\n", err)
			quit <- os.Interrupt
		}
	}()

	// Wait for the OS
	<-quit
	fiberlog.Info("Shutting down server...")

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down the server
	if err := app.ShutdownWithContext(ctx); err != nil {
		fiberlog.Errorf("Error shutting down server: %s\n", err)
	}

	fiberlog.Infof("Server stopped")
}
