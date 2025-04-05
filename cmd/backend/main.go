package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/szehnder/bma-calculator/pkg/backend"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	app := fiber.New()

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Connect to MongoDB
	err := backend.ConnectDB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}

	// Setup routes
	backend.SetupRoutes(app)

	log.Info().Msg("Server listening on http://localhost:8080")
	log.Fatal().Err(app.Listen(":8080")).Msg("")
}
