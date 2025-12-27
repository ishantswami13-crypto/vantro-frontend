package expense

import (
	"bytes"
	"fmt"

	"github.com/phpdave11/gofpdf"
)

func BuildMonthlyPDF(sum *MonthlySummary) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Vantro Expense Memory Report", false)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 10, "Vantro Expense Memory")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Report Month: %s", sum.Month))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("User: %s", sum.UserPhone))
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 14)
	pdf.Cell(0, 8, fmt.Sprintf("Total Spend: ₹%.2f", sum.TotalRupees))
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 12)
	pdf.MultiCell(0, 7, "Insight: "+sum.Insight, "", "L", false)
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 13)
	pdf.Cell(0, 8, "Category Breakdown")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.Cell(70, 7, "Category")
	pdf.Cell(50, 7, "Amount")
	pdf.Cell(30, 7, "%")
	pdf.Ln(7)

	pdf.SetFont("Helvetica", "", 11)
	for _, b := range sum.CategoryBreakup {
		pdf.Cell(70, 7, b.Category)
		pdf.Cell(50, 7, fmt.Sprintf("₹%.2f", b.TotalRs))
		pdf.Cell(30, 7, fmt.Sprintf("%.1f%%", b.Percent))
		pdf.Ln(7)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
