package money

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidMoney = errors.New("invalid money amount")
)

// RupeesToPaise converts a rupee value (like 12.34) to paise as int64 safely.
// Use ONLY when you must parse user-entered decimal rupees.
// Prefer sending paise directly from frontend.
func RupeesToPaise(rupees float64) (int64, error) {
	if math.IsNaN(rupees) || math.IsInf(rupees, 0) {
		return 0, ErrInvalidMoney
	}
	if rupees < 0 {
		return 0, ErrInvalidMoney
	}
	// Prevent overflow: int64 max ~9e18 => rupees max ~9e16
	if rupees > 9e16 {
		return 0, fmt.Errorf("%w: too large", ErrInvalidMoney)
	}
	paise := int64(math.Round(rupees * 100.0))
	if paise < 0 {
		return 0, ErrInvalidMoney
	}
	return paise, nil
}

func PaiseToRupeesString(paise int64) string {
	// Lightweight formatting without float: ₹123.45 style string
	sign := ""
	if paise < 0 {
		sign = "-"
		paise = -paise
	}
	rs := paise / 100
	ps := paise % 100
	return fmt.Sprintf("%s%d.%02d", sign, rs, ps)
}
