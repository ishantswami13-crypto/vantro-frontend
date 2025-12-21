package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OnboardingHandler struct {
	DB *pgxpool.Pool
}

type onboardingStepRequest struct {
	Step string `json:"step"`
}

// UpdateStep updates the onboarding_step for the authenticated user.
func (h *OnboardingHandler) UpdateStep(c *fiber.Ctx) error {
	userID := strings.TrimSpace(getUserID(c))
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var body onboardingStepRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	step := strings.TrimSpace(body.Step)
	if step == "" {
		return fiber.NewError(fiber.StatusBadRequest, "step required")
	}

	ctx := userContext(c)
	if _, err := h.DB.Exec(ctx, `UPDATE users SET onboarding_step = $1 WHERE id = $2`, step, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update step")
	}

	return c.JSON(fiber.Map{"ok": true, "step": step})
}

func getUserID(c *fiber.Ctx) string {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	if v := c.Locals("userID"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
