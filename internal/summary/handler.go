package summary

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Repo Repo
}

func (h Handler) GetSummary(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	month := strings.TrimSpace(c.Query("month")) // YYYY-MM or empty

	s, err := h.Repo.GetByUser(userContext(c), userID, month)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch summary: "+err.Error())
	}

	return c.JSON(s)
}

func extractUserID(c *fiber.Ctx) (string, error) {
	val := c.Locals("user_id")
	if val == nil {
		val = c.Locals("userID")
	}
	if val == nil {
		return "", errors.New("user id missing")
	}
	if uid, ok := val.(string); ok && strings.TrimSpace(uid) != "" {
		return uid, nil
	}
	return "", errors.New("user id missing")
}

func userContext(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}
