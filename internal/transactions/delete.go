package transactions

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (r *Repo) DeleteIncomeByID(ctx context.Context, userID, id string) error {
	ct, err := r.Pool.Exec(ctx, `UPDATE incomes SET deleted_at = $3 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID, time.Now().UTC())
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *Repo) DeleteExpenseByID(ctx context.Context, userID, id string) error {
	ct, err := r.Pool.Exec(ctx, `UPDATE expenses SET deleted_at = $3 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID, time.Now().UTC())
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *Repo) UndoIncomeByID(ctx context.Context, userID, id string) error {
	ct, err := r.Pool.Exec(ctx, `UPDATE incomes SET deleted_at = NULL WHERE id = $1 AND user_id = $2 AND deleted_at IS NOT NULL`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *Repo) UndoExpenseByID(ctx context.Context, userID, id string) error {
	ct, err := r.Pool.Exec(ctx, `UPDATE expenses SET deleted_at = NULL WHERE id = $1 AND user_id = $2 AND deleted_at IS NOT NULL`, id, userID)
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
