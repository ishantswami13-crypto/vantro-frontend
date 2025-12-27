package whatsapp

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ishantswami13-crypto/vantro-backend/internal/billing"
	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
)

// Twilio sends x-www-form-urlencoded with fields like:
// From=whatsapp:+91xxxx  Body=250 food pizza  ProfileName=... etc.
// We respond with TwiML XML (simple). Twilio accepts MessagingResponse XML.
func InboundHandler(expStore *expense.Store, billStore *billing.Store, rp *billing.RazorpayClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		from := c.FormValue("From") // "whatsapp:+91..."
		body := strings.TrimSpace(c.FormValue("Body"))

		phone := normalizeTwilioWhatsAppFrom(from)
		if phone == "" {
			writeTwiML(c, "Phone not detected. Try again.")
			return nil
		}

		cmd := strings.ToLower(strings.TrimSpace(body))

		switch cmd {
		case "help", "h", "menu":
			writeTwiML(c, "Commands:\n- total\n- report\nFormat:\n- 250 food pizza\n- 180 uber\n- 99 coffee")
			return nil

		case "total":
			now := time.Now()
			sum, err := expStore.MonthlySummary(c.Context(), phone, now.Year(), int(now.Month()))
			if err != nil {
				writeTwiML(c, "Server error. Try again.")
				return nil
			}
			writeTwiML(c, "This month total: Rs "+format2(sum.TotalRupees)+"\nTop: "+sum.TopCategory+"\nTxns: "+itoa64(sum.Transactions))
			return nil

		case "report":
			active, err := billStore.IsActive(c.Context(), phone)
			if err != nil {
				writeTwiML(c, "Server error. Try again.")
				return nil
			}
			if active {
				writeTwiML(c, "You are active.\nUse the app/API to download PDF report.\n(Next update: I will send it here.)")
				return nil
			}

			link, err := rp.CreateMonthlyLink(c.Context(), phone)
			if err != nil {
				writeTwiML(c, "Payment link error. Try again.")
				return nil
			}

			writeTwiML(c, "Unlock monthly PDF + insights (Rs 199):\n"+link.ShortURL+"\nAfter payment, reply: total")
			return nil
		}

		if body == "" {
			writeTwiML(c, "Send like: 250 food pizza\nType: help")
			return nil
		}

		_, err := expStore.AddExpense(c.Context(), expense.AddExpenseRequest{
			UserPhone: phone,
			Text:      body,
			Source:    "whatsapp",
		})
		if err != nil {
			writeTwiML(c, "Could not log that. Try: 250 food coffee\nType: help")
			return nil
		}

		writeTwiML(c, "Logged. Try: total | report")
		return nil
	}
}

func normalizeTwilioWhatsAppFrom(from string) string {
	from = strings.TrimSpace(from)
	from = strings.TrimPrefix(from, "whatsapp:")
	return strings.TrimSpace(from)
}

func writeTwiML(c *fiber.Ctx, msg string) {
	c.Set("Content-Type", "application/xml")
	c.Status(http.StatusOK)
	_ = c.SendString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Response><Message>" + escapeXML(msg) + "</Message></Response>")
}

func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}

func format2(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}
