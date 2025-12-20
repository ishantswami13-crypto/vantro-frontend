package transactions

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{Repo: repo}
}

func getUserID(c *fiber.Ctx) (string, bool) {
	// Your JWT middleware sets one of these (you already used both in income handler)
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

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	if req.Type != "income" && req.Type != "expense" {
		return fiber.NewError(fiber.StatusBadRequest, "type must be income or expense")
	}

	if req.Amount < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be >= 0")
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = c.Context()
	}

	id, err := h.Repo.Create(ctx, userID, req.Type, req.Amount, req.Note)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create transaction: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
	})
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = c.Context()
	}

	items, err := h.Repo.List(ctx, userID, 50)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list transactions: "+err.Error())
	}

	// Normalize created_at for frontend if you want (optional)
	out := make([]fiber.Map, 0, len(items))
	for _, t := range items {
		out = append(out, fiber.Map{
			"id":         t.ID,
			"type":       t.Type,
			"amount":     t.Amount,
			"note":       t.Note,
			"created_at": t.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(out)
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = c.Context()
	}

	s, err := h.Repo.Summary(ctx, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to compute summary: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"income":  s.Income,
		"expense": s.Expense,
		"net":     s.Net,
	})
}
