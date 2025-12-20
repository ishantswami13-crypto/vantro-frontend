package admin

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func RequireAdminAPIKey() fiber.Handler {
	key := strings.TrimSpace(os.Getenv("ADMIN_API_KEY"))
	// If you forget to set it, we hard-fail (safer than accidentally open).
	if key == "" {
		return func(c *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusInternalServerError, "ADMIN_API_KEY not set")
		}
	}

	return func(c *fiber.Ctx) error {
		got := strings.TrimSpace(c.Get("X-Admin-Key"))
		if got == "" || got != key {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid admin key")
		}
		return c.Next()
	}
}
