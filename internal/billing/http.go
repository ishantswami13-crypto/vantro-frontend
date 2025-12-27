package billing

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
	"github.com/ishantswami13-crypto/vantro-backend/internal/reports"
)

type pdfSender interface {
	SendWhatsAppPDF(ctx context.Context, toPhone, caption, pdfURL string) error
}

func CreatePaymentLinkHandler(store *Store, rp *RazorpayClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateLinkRequest
		if err := c.BodyParser(&req); err != nil || req.UserPhone == "" {
			return c.Status(fiber.StatusBadRequest).SendString("user_phone required")
		}

		link, err := rp.CreateMonthlyLink(c.Context(), req.UserPhone)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("razorpay error")
		}

		// Optional: auto-activate. Better to activate via webhook; keep commented.
		// _ = store.ActivateFor30Days(c.Context(), req.UserPhone, link.ID)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"payment_link_id": link.ID,
			"short_url":       link.ShortURL,
			"status":          link.Status,
		})
	}
}

func RazorpayWebhookHandler(
	billStore *Store,
	expStore *expense.Store,
	repStore *reports.Store,
	tw pdfSender,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Body()
		sig := c.Get("X-Razorpay-Signature")
		secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")

		if secret == "" || sig == "" || !VerifyWebhookSignature(raw, sig, secret) {
			return c.Status(fiber.StatusUnauthorized).SendString("invalid signature")
		}

		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad payload")
		}

		evt, _ := payload["event"].(string)
		if evt != "payment_link.paid" {
			return c.Status(fiber.StatusOK).SendString("ignored")
		}

		phone := ""
		plID := ""

		if p, ok := payload["payload"].(map[string]any); ok {
			if pl, ok := p["payment_link"].(map[string]any); ok {
				if ent, ok := pl["entity"].(map[string]any); ok {
					if id, ok := ent["id"].(string); ok {
						plID = id
					}
					if cust, ok := ent["customer"].(map[string]any); ok {
						if cphone, ok := cust["contact"].(string); ok {
							phone = cphone
						}
					}
				}
			}
		}

		if phone == "" {
			return c.Status(fiber.StatusBadRequest).SendString("phone not found")
		}

		if err := billStore.ActivateFor30Days(c.Context(), phone, plID); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("db error")
		}

		now := time.Now()
		sum, err := expStore.MonthlySummary(c.Context(), phone, now.Year(), int(now.Month()))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("summary error")
		}

		pdfBytes, err := expense.BuildMonthlyPDF(sum)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("pdf error")
		}

		dir := filepath.Join("data", "reports", safePhone(phone))
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("file error")
		}
		filePath := filepath.Join(dir, "vantro-"+sum.Month+".pdf")
		if writeErr := os.WriteFile(filePath, pdfBytes, 0o644); writeErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("file error")
		}

		token, _, err := repStore.Create(c.Context(), phone, sum.Month, filePath, 7*24*time.Hour)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("report token error")
		}

		base := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/")
		pdfURL := base + "/r/" + token

		if tw != nil {
			_ = tw.SendWhatsAppPDF(c.Context(), phone, "Your Vantro Expense Memory report: "+sum.Month, pdfURL)
		}

		return c.Status(fiber.StatusOK).SendString("ok")
	}
}

func safePhone(p string) string {
	out := make([]rune, 0, len(p))
	for _, r := range p {
		if (r >= '0' && r <= '9') || r == '+' {
			out = append(out, r)
		}
	}
	return string(out)
}
