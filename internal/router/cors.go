package router

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CorsMiddleware configures CORS based on CORS_ORIGIN (defaults to *).
func CorsMiddleware() fiber.Handler {
	origin := strings.TrimSpace(os.Getenv("CORS_ORIGIN"))
	if origin == "" {
		origin = "*"
	}

	return cors.New(cors.Config{
		AllowOrigins:     origin,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: false,
	})
}
