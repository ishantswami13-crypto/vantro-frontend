package expense

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) CreateExpense(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var req CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	req.VendorName = strings.TrimSpace(req.VendorName)
	if req.VendorName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "vendor_name required")
	}
	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}

	spentOn, err := time.Parse("2006-01-02", req.SpentOn)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "spent_on must be YYYY-MM-DD")
	}

	exp := &Expense{
		UserID:     userID,
		VendorName: req.VendorName,
		Amount:     req.Amount,
		Currency:   "INR",
		SpentOn:    spentOn,
		Note:       req.Note,
	}

	id, err := h.Repo.InsertExpense(userContext(c), exp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add expense: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(CreateExpenseResponse{
		ID:      id,
		Message: "expense added",
	})
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
