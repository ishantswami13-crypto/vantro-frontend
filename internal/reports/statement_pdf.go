package reports

import (
	"bytes"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/phpdave11/gofpdf"
)

func (h *Handler) StatementPDF(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	if uidVal == nil {
		uidVal = c.Locals("userID")
	}
	userID, _ := uidVal.(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))
	if from == "" || to == "" {
		end := time.Now()
		start := end.AddDate(0, 0, -29)
		from = start.Format("2006-01-02")
		to = end.Format("2006-01-02")
	}

	if _, err := time.Parse("2006-01-02", from); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "from must be YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "to must be YYYY-MM-DD")
	}

	ctx := c.UserContext()
	rows, err := h.Pool.Query(ctx, `
SELECT type, id, title, amount, currency, date, created_at
FROM (
  SELECT 'income' AS type,
         id::text AS id,
         client_name AS title,
         amount::bigint AS amount,
         COALESCE(currency,'INR') AS currency,
         received_on::text AS date,
         created_at::text AS created_at
  FROM incomes
  WHERE user_id=$1 AND deleted_at IS NULL AND received_on BETWEEN $2::date AND $3::date

  UNION ALL

  SELECT 'expense' AS type,
         id::text AS id,
         vendor_name AS title,
         amount::bigint AS amount,
         COALESCE(currency,'INR') AS currency,
         spent_on::text AS date,
         created_at::text AS created_at
  FROM expenses
  WHERE user_id=$1 AND deleted_at IS NULL AND spent_on BETWEEN $2::date AND $3::date
) t
ORDER BY date DESC, created_at DESC
LIMIT 2000
`, userID, from, to)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed statement: "+err.Error())
	}
	defer rows.Close()

	type row struct {
		Type      string
		ID        string
		Title     string
		Amount    int64
		Currency  string
		Date      string
		CreatedAt string
	}

	var items []row
	currency := "INR"
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Type, &r.ID, &r.Title, &r.Amount, &r.Currency, &r.Date, &r.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "scan statement: "+err.Error())
		}
		if strings.TrimSpace(r.Currency) != "" {
			currency = r.Currency
		}
		items = append(items, r)
	}

	var totalIncome int64
	if err := h.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0)
		FROM incomes
		WHERE user_id=$1 AND deleted_at IS NULL AND received_on BETWEEN $2::date AND $3::date
	`, userID, from, to).Scan(&totalIncome); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "totals income: "+err.Error())
	}

	var totalExpense int64
	if err := h.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0)
		FROM expenses
		WHERE user_id=$1 AND deleted_at IS NULL AND spent_on BETWEEN $2::date AND $3::date
	`, userID, from, to).Scan(&totalExpense); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "totals expense: "+err.Error())
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 48)
	pdf.SetTextColor(235, 235, 235)
	pdf.Text(25, 140, "VANTRO")

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 10, "VANTRO Statement")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 6, "Period: "+from+" to "+to)
	pdf.Ln(5)
	pdf.Cell(0, 6, "User: "+maskID(userID))
	pdf.Ln(10)

	pdf.SetDrawColor(200, 200, 200)
	pdf.SetFillColor(248, 248, 248)
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "B", 11)

	sumW := []float64{62, 62, 62}
	pdf.CellFormat(sumW[0], 10, "Income ("+currency+")", "1", 0, "C", true, 0, "")
	pdf.CellFormat(sumW[1], 10, "Expense ("+currency+")", "1", 0, "C", true, 0, "")
	pdf.CellFormat(sumW[2], 10, "Balance ("+currency+")", "1", 1, "C", true, 0, "")

	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(sumW[0], 10, formatMoney(totalIncome), "1", 0, "C", false, 0, "")
	pdf.CellFormat(sumW[1], 10, formatMoney(totalExpense), "1", 0, "C", false, 0, "")
	pdf.CellFormat(sumW[2], 10, formatMoney(totalIncome-totalExpense), "1", 1, "C", false, 0, "")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetTextColor(20, 20, 20)

	colW := []float64{22, 26, 92, 30, 20}
	pdf.CellFormat(colW[0], 8, "TYPE", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[1], 8, "DATE", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[2], 8, "TITLE", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colW[3], 8, "AMOUNT", "1", 0, "R", true, 0, "")
	pdf.CellFormat(colW[4], 8, "ID", "1", 1, "C", true, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(30, 30, 30)

	maxRows := 200
	for i, it := range items {
		if i >= maxRows {
			pdf.SetFont("Helvetica", "I", 9)
			pdf.CellFormat(0, 8, "…truncated (too many rows)", "1", 1, "C", false, 0, "")
			break
		}

		typ := strings.ToUpper(it.Type)
		date := it.Date
		title := it.Title
		amt := formatMoneySigned(it.Amount, it.Type)

		if pdf.GetY() > 270 {
			pdf.AddPage()
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetFillColor(245, 245, 245)
			pdf.CellFormat(colW[0], 8, "TYPE", "1", 0, "C", true, 0, "")
			pdf.CellFormat(colW[1], 8, "DATE", "1", 0, "C", true, 0, "")
			pdf.CellFormat(colW[2], 8, "TITLE", "1", 0, "L", true, 0, "")
			pdf.CellFormat(colW[3], 8, "AMOUNT", "1", 0, "R", true, 0, "")
			pdf.CellFormat(colW[4], 8, "ID", "1", 1, "C", true, 0, "")
			pdf.SetFont("Helvetica", "", 9)
		}

		pdf.CellFormat(colW[0], 8, typ, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[1], 8, date, "1", 0, "C", false, 0, "")

		x := pdf.GetX()
		y := pdf.GetY()

		pdf.MultiCell(colW[2], 8, trimTo(title, 90), "1", "L", false)
		usedH := pdf.GetY() - y
		pdf.SetXY(x+colW[2], y)

		pdf.CellFormat(colW[3], usedH, amt, "1", 0, "R", false, 0, "")
		pdf.CellFormat(colW[4], usedH, shortID(it.ID), "1", 1, "C", false, 0, "")
	}

	pdf.SetY(-18)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 10, "Generated by VANTRO • "+time.Now().Format(time.RFC3339), "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "pdf build failed: "+err.Error())
	}

	filename := "vantro-statement-" + from + "-to-" + to + ".pdf"
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Send(buf.Bytes())
}

func shortID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func maskID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 8 {
		return id
	}
	return id[:4] + "…" + id[len(id)-4:]
}

func trimTo(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func formatMoney(n int64) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	return sign + withCommas(n)
}

func formatMoneySigned(n int64, typ string) string {
	if strings.ToLower(typ) == "expense" && n > 0 {
		return "-" + withCommas(n)
	}
	return withCommas(n)
}

func withCommas(n int64) string {
	s := []byte{}
	str := []byte(strings.TrimSpace(intToStr(n)))
	l := len(str)
	for i := 0; i < l; i++ {
		s = append(s, str[i])
		rem := l - i - 1
		if rem > 0 && rem%3 == 0 {
			s = append(s, ',')
		}
	}
	return string(s)
}

func intToStr(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
