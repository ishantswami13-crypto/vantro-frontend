package transactions

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{Repo: repo}
}

func getUserID(c *fiber.Ctx) (string, bool) {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s, true
		}
	}
	if v := c.Locals("userID"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s, true
		}
	}
	return "", false
}

func (h *Handler) ListLatest(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := userContext(c)

	items, err := h.Repo.ListLatest(ctx, userID, 50)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load transactions: "+err.Error())
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) GetSummary(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := userContext(c)

	s, err := h.Repo.GetSummary(ctx, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to compute summary: "+err.Error())
	}

	return c.JSON(s)
}

func userContext(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}
