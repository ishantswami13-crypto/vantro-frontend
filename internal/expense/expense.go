package expense

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Store struct {
	DB *sql.DB
}

type Expense struct {
	ID          int64     `json:"id"`
	UserPhone   string    `json:"user_phone"`
	AmountPaise int64     `json:"amount_paise"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Note        string    `json:"note,omitempty"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
}

type AddExpenseRequest struct {
	UserPhone string `json:"user_phone"`
	// amount in rupees (e.g. 250.50). We'll convert to paise.
	AmountRupees float64 `json:"amount_rupees"`
	Category     string  `json:"category,omitempty"`
	Note         string  `json:"note,omitempty"`
	Source       string  `json:"source,omitempty"` // manual by default
	// Optional: raw text like "250 food pizza"
	Text string `json:"text,omitempty"`
}

type MonthlySummary struct {
	UserPhone       string           `json:"user_phone"`
	Month           string           `json:"month"` // YYYY-MM
	TotalPaise      int64            `json:"total_paise"`
	TotalRupees     float64          `json:"total_rupees"`
	TopCategory     string           `json:"top_category"`
	CategoryBreakup []CategoryBucket `json:"category_breakup"`
	Insight         string           `json:"insight"`
	Transactions    int64            `json:"transactions"`
}

type CategoryBucket struct {
	Category   string  `json:"category"`
	TotalPaise int64   `json:"total_paise"`
	TotalRs    float64 `json:"total_rupees"`
	Percent    float64 `json:"percent"`
}

var (
	ErrBadRequest = errors.New("bad request")
)

// ---------------------------
// Categorization (rule-based)
// ---------------------------

func normalizeCategory(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return "MISC"
	}
	// keep it simple
	switch s {
	case "FOOD", "TRANSPORT", "FIXED", "SHOPPING", "HEALTH", "ENTERTAINMENT", "BILLS", "MISC":
		return s
	default:
		return "MISC"
	}
}

func categorizeFromText(text string) (amountRupees float64, category string, note string, ok bool) {
	// Accept formats like:
	// "250 food pizza"
	// "uber 180"
	// "Spent 99 coffee"
	// We'll try to extract first number as amount, rest as note/category guess.

	t := strings.TrimSpace(text)
	if t == "" {
		return 0, "", "", false
	}

	// find first number (integer or decimal)
	re := regexp.MustCompile(`(\d+(\.\d+)?)`)
	m := re.FindStringSubmatch(t)
	if len(m) < 2 {
		return 0, "", "", false
	}

	// parse amount
	var amt float64
	_, err := fmtSscanfFloat(m[1], &amt)
	if err != nil || amt <= 0 {
		return 0, "", "", false
	}

	// remove the first number from text to get remaining tokens
	idx := strings.Index(t, m[1])
	rest := strings.TrimSpace(t[:idx] + t[idx+len(m[1]):])
	restLower := strings.ToLower(rest)

	// rule-based categories
	cat := "MISC"
	switch {
	case containsAny(restLower, "zomato", "swiggy", "food", "pizza", "burger", "coffee", "chai", "tea", "restaurant", "dinner", "lunch", "breakfast"):
		cat = "FOOD"
	case containsAny(restLower, "uber", "ola", "auto", "metro", "bus", "cab", "rapido", "petrol", "fuel"):
		cat = "TRANSPORT"
	case containsAny(restLower, "rent", "emi", "loan", "school fee", "fees", "insurance"):
		cat = "FIXED"
	case containsAny(restLower, "bill", "electricity", "gas", "water", "recharge", "wifi", "broadband"):
		cat = "BILLS"
	case containsAny(restLower, "netflix", "prime", "hotstar", "movie", "game", "spotify"):
		cat = "ENTERTAINMENT"
	case containsAny(restLower, "doctor", "medicine", "pharmacy", "gym", "protein"):
		cat = "HEALTH"
	case containsAny(restLower, "amazon", "flipkart", "shopping", "clothes", "shoes"):
		cat = "SHOPPING"
	}

	return amt, cat, rest, true
}

func containsAny(s string, words ...string) bool {
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

// Small helper: avoid importing fmt just for Sscanf issues in some setups
func fmtSscanfFloat(str string, out *float64) (int, error) {
	// minimal parse
	// NOTE: using strconv is more standard; keeping it direct & safe:
	return 1, parseFloat(str, out)
}

func parseFloat(str string, out *float64) error {
	// use strconv for safety
	// kept in a separate function so it’s easy to change.
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	*out = f
	return nil
}

// ---------------------------
// Store methods
// ---------------------------

func (s *Store) AddExpense(ctx context.Context, req AddExpenseRequest) (*Expense, error) {
	req.UserPhone = strings.TrimSpace(req.UserPhone)
	if req.UserPhone == "" {
		return nil, ErrBadRequest
	}

	amountRupees := req.AmountRupees
	category := req.Category
	note := strings.TrimSpace(req.Note)

	// If text is provided, try parse from it (nice UX)
	if strings.TrimSpace(req.Text) != "" {
		amt, cat, parsedNote, ok := categorizeFromText(req.Text)
		if ok {
			amountRupees = amt
			category = cat
			// only set note if not explicitly given
			if note == "" {
				note = parsedNote
			}
		}
	}

	if amountRupees <= 0 {
		return nil, ErrBadRequest
	}
	if req.Source == "" {
		req.Source = "manual"
	}

	amountPaise := int64(amountRupees * 100.0)
	// protect against float rounding weirdness:
	if amountPaise <= 0 {
		return nil, ErrBadRequest
	}

	category = normalizeCategory(category)

	const q = `
        INSERT INTO expenses (user_phone, amount_paise, currency, category, note, source)
        VALUES ($1, $2, 'INR', $3, $4, $5)
        RETURNING id, user_phone, amount_paise, currency, category, note, source, created_at;
    `

	var e Expense
	err := s.DB.QueryRowContext(ctx, q, req.UserPhone, amountPaise, category, note, req.Source).
		Scan(&e.ID, &e.UserPhone, &e.AmountPaise, &e.Currency, &e.Category, &e.Note, &e.Source, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

type ListExpensesParams struct {
	UserPhone string
	Limit     int
}

func (s *Store) ListExpenses(ctx context.Context, p ListExpensesParams) ([]Expense, error) {
	p.UserPhone = strings.TrimSpace(p.UserPhone)
	if p.UserPhone == "" {
		return nil, ErrBadRequest
	}
	if p.Limit <= 0 || p.Limit > 200 {
		p.Limit = 50
	}

	const q = `
        SELECT id, user_phone, amount_paise, currency, category, note, source, created_at
        FROM expenses
        WHERE user_phone = $1
        ORDER BY created_at DESC
        LIMIT $2;
    `

	rows, err := s.DB.QueryContext(ctx, q, p.UserPhone, p.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Expense
	for rows.Next() {
		var e Expense
		if err := rows.Scan(&e.ID, &e.UserPhone, &e.AmountPaise, &e.Currency, &e.Category, &e.Note, &e.Source, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func monthRange(year int, month int) (time.Time, time.Time, error) {
	if month < 1 || month > 12 {
		return time.Time{}, time.Time{}, ErrBadRequest
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

func (s *Store) MonthlySummary(ctx context.Context, userPhone string, year int, month int) (*MonthlySummary, error) {
	userPhone = strings.TrimSpace(userPhone)
	if userPhone == "" {
		return nil, ErrBadRequest
	}
	start, end, err := monthRange(year, month)
	if err != nil {
		return nil, err
	}

	const q = `
        SELECT category, COALESCE(SUM(amount_paise),0) AS total_paise, COUNT(*) AS txns
        FROM expenses
        WHERE user_phone = $1 AND created_at >= $2 AND created_at < $3
        GROUP BY category;
    `

	rows, err := s.DB.QueryContext(ctx, q, userPhone, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rowAgg struct {
		Cat   string
		Paise int64
		Txns  int64
	}
	var aggs []rowAgg
	var total int64
	var txns int64

	for rows.Next() {
		var r rowAgg
		if err := rows.Scan(&r.Cat, &r.Paise, &r.Txns); err != nil {
			return nil, err
		}
		aggs = append(aggs, r)
		total += r.Paise
		txns += r.Txns
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build buckets
	buckets := make([]CategoryBucket, 0, len(aggs))
	for _, a := range aggs {
		pct := 0.0
		if total > 0 {
			pct = (float64(a.Paise) / float64(total)) * 100.0
		}
		buckets = append(buckets, CategoryBucket{
			Category:   a.Cat,
			TotalPaise: a.Paise,
			TotalRs:    float64(a.Paise) / 100.0,
			Percent:    pct,
		})
	}

	sort.Slice(buckets, func(i, j int) bool { return buckets[i].TotalPaise > buckets[j].TotalPaise })

	topCat := "NONE"
	if len(buckets) > 0 {
		topCat = buckets[0].Category
	}

	insight := buildInsight(buckets, total)

	sum := &MonthlySummary{
		UserPhone:       userPhone,
		Month:           start.Format("2006-01"),
		TotalPaise:      total,
		TotalRupees:     float64(total) / 100.0,
		TopCategory:     topCat,
		CategoryBreakup: buckets,
		Insight:         insight,
		Transactions:    txns,
	}
	return sum, nil
}

func buildInsight(buckets []CategoryBucket, total int64) string {
	if total <= 0 {
		return "No spends logged this month. Calm wallet energy."
	}
	if len(buckets) == 0 {
		return "No category data this month."
	}
	top := buckets[0]
	// simple “roast but helpful”
	if top.Percent >= 45 {
		return "Reality check: almost half your money went to " + strings.ToLower(top.Category) + ". Tiny tweaks here = big savings."
	}
	if top.Category == "MISC" && top.Percent >= 25 {
		return "A lot is going into MISC. Add 1–2 words in your message so Vantro can categorize better."
	}
	return "Good signal: your spend is fairly spread out. Keep logging — trends show up after 2–3 months."
}
