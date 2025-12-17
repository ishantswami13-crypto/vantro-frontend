package expense

import (
	"context"
	"errors"
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

type createExpenseReq struct {
	VendorName string `json:"vendor_name"`
	Amount     int64  `json:"amount"`
	SpentOn    string `json:"spent_on"` // YYYY-MM-DD
	Note       string `json:"note"`
}

func (h *Handler) AddExpense(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	var req createExpenseReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	req.VendorName = strings.TrimSpace(req.VendorName)
	if req.VendorName == "" || req.Amount <= 0 || strings.TrimSpace(req.SpentOn) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "vendor_name, amount, spent_on required")
	}

	spentOn, err := time.Parse("2006-01-02", req.SpentOn)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "spent_on must be YYYY-MM-DD")
	}

	var notePtr *string
	if strings.TrimSpace(req.Note) != "" {
		n := req.Note
		notePtr = &n
	}

	id, err := h.Repo.AddExpense(userContext(c), userID, req.VendorName, req.Amount, spentOn, notePtr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add expense: "+err.Error())
	}

	return c.JSON(fiber.Map{"id": id, "message": "expense added"})
}

func (h *Handler) ListExpenses(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	items, err := h.Repo.ListExpensesByUser(userContext(c), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list expenses: "+err.Error())
	}

	return c.JSON(items)
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
