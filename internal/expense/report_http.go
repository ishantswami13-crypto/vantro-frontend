package expense

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type subscriptionChecker interface {
	IsActive(ctx context.Context, phone string) (bool, error)
}

func MonthlyPDFHandler(expStore *Store, checker subscriptionChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		phone := c.Query("phone")
		yearStr := c.Query("year")
		monthStr := c.Query("month")

		if phone == "" || yearStr == "" || monthStr == "" {
			return c.Status(fiber.StatusBadRequest).SendString("phone, year, month required")
		}

		year, err1 := strconv.Atoi(yearStr)
		month, err2 := strconv.Atoi(monthStr)
		if err1 != nil || err2 != nil {
			return c.Status(fiber.StatusBadRequest).SendString("year/month invalid")
		}

		active, err := checker.IsActive(c.Context(), phone)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("server error")
		}
		if !active {
			return c.Status(402).SendString("subscription required for PDF report")
		}

		sum, err := expStore.MonthlySummary(c.Context(), phone, year, month)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("server error")
		}

		pdfBytes, err := BuildMonthlyPDF(sum)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("pdf error")
		}

		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "attachment; filename=vantro-expense-report-"+sum.Month+".pdf")
		return c.Status(fiber.StatusOK).Send(pdfBytes)
	}
}
