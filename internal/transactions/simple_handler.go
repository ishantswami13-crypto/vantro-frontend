package transactions

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type SimpleHandler struct {
	Repo *SimpleRepo
}

func NewSimpleHandler(repo *SimpleRepo) *SimpleHandler {
	return &SimpleHandler{Repo: repo}
}

func (h *SimpleHandler) Create(c *fiber.Ctx) error {
	userID := strings.TrimSpace(getUserFromCtx(c))
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var body CreateTxnRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	tx, err := h.Repo.Create(c.UserContext(), userID, body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(tx)
}

func (h *SimpleHandler) List(c *fiber.Ctx) error {
	userID := strings.TrimSpace(getUserFromCtx(c))
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	items, err := h.Repo.List(c.UserContext(), userID, 50)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load transactions")
	}
	return c.JSON(fiber.Map{"items": items})
}

func getUserFromCtx(c *fiber.Ctx) string {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}
