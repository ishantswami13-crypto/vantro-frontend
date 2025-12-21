package transactions

import (
	"context"
	"errors"
	"strings"
)

func (r *Repo) DeleteIncomeByID(ctx context.Context, userID, id string) error {
	ct, err := r.Pool.Exec(ctx, `DELETE FROM incomes WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *Repo) DeleteExpenseByID(ctx context.Context, userID, id string) error {
	ct, err := r.Pool.Exec(ctx, `DELETE FROM expenses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

func normalizeType(t string) string {
	t = strings.TrimSpace(strings.ToLower(t))
	if t == "income" || t == "expense" {
		return t
	}
	return ""
}
