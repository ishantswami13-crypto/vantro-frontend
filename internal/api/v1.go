package api

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	DB *sql.DB
}

type createTxnRequest struct {
	Amount    int64  `json:"amount"`
	Direction string `json:"direction"`
	Note      string `json:"note"`
}

func (s *Server) CreateTransaction(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return jsonErr(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var body createTxnRequest
	if err := c.BodyParser(&body); err != nil {
		return jsonErr(c, fiber.StatusBadRequest, "invalid body")
	}
	body.Direction = strings.ToUpper(strings.TrimSpace(body.Direction))
	if body.Direction != "IN" && body.Direction != "OUT" {
		return jsonErr(c, fiber.StatusBadRequest, "direction must be IN or OUT")
	}
	if body.Amount <= 0 {
		return jsonErr(c, fiber.StatusBadRequest, "amount must be greater than zero")
	}

	ctx := c.UserContext()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	var id int64
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO transactions_v1 (user_id, amount, direction, note)
		VALUES ($1, $2, $3, NULLIF($4,''))
		RETURNING id, created_at
	`, userID, body.Amount, body.Direction, body.Note).Scan(&id, &createdAt); err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	pointsAwarded := int64(0)
	if body.Direction == "OUT" {
		pointsAwarded = body.Amount / 100
		if pointsAwarded > 0 {
			src := strconv.FormatInt(id, 10)
			if err := awardPointsTx(ctx, tx, userID, src, pointsAwarded, "earn"); err != nil {
				return jsonErr(c, fiber.StatusInternalServerError, err.Error())
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"id":             id,
		"amount":         body.Amount,
		"direction":      body.Direction,
		"created_at":     createdAt.Format(time.RFC3339),
		"points_awarded": pointsAwarded,
	})
}

func (s *Server) ListTransactions(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return jsonErr(c, fiber.StatusUnauthorized, "unauthorized")
	}
	rows, err := s.DB.QueryContext(c.UserContext(), `
		SELECT id, amount, direction, COALESCE(note,''), created_at
		FROM transactions_v1
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	type item struct {
		ID        int64  `json:"id"`
		Amount    int64  `json:"amount"`
		Direction string `json:"direction"`
		Note      string `json:"note"`
		CreatedAt string `json:"created_at"`
	}
	var out []item
	for rows.Next() {
		var it item
		var created time.Time
		if err := rows.Scan(&it.ID, &it.Amount, &it.Direction, &it.Note, &created); err != nil {
			return jsonErr(c, fiber.StatusInternalServerError, err.Error())
		}
		it.CreatedAt = created.Format(time.RFC3339)
		out = append(out, it)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (s *Server) PointsSummary(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return jsonErr(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var total int64
	err := s.DB.QueryRowContext(c.UserContext(), `
		SELECT points_total FROM points_balance WHERE user_id = $1
	`, userID).Scan(&total)
	if errors.Is(err, sql.ErrNoRows) {
		total = 0
	} else if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	tiers, err := loadTiers(c.UserContext(), s.DB)
	if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	cur := tiers[0]
	next := tiers[len(tiers)-1]
	for i, t := range tiers {
		if total >= t.Min {
			cur = t
			if i+1 < len(tiers) {
				next = tiers[i+1]
			} else {
				next = t
			}
		}
	}
	progress := 1.0
	if next.Min > cur.Min {
		progress = float64(total-cur.Min) / float64(next.Min-cur.Min)
		progress = math.Min(1, math.Max(0, progress))
	}

	return c.JSON(fiber.Map{
		"points_total":         total,
		"tier":                 cur.Name,
		"multiplier":           cur.Multiplier,
		"next_tier":            next.Name,
		"next_tier_min_points": next.Min,
		"progress_to_next":     progress,
	})
}

func (s *Server) PointsLedger(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return jsonErr(c, fiber.StatusUnauthorized, "unauthorized")
	}

	rows, err := s.DB.QueryContext(c.UserContext(), `
		SELECT id, source_txn_id, points_delta, reason, created_at
		FROM points_ledger
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	type row struct {
		ID        int64   `json:"id"`
		SourceTxn *string `json:"source_txn_id"`
		Delta     int     `json:"points_delta"`
		Reason    string  `json:"reason"`
		CreatedAt string  `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		var created time.Time
		if err := rows.Scan(&r.ID, &r.SourceTxn, &r.Delta, &r.Reason, &created); err != nil {
			return jsonErr(c, fiber.StatusInternalServerError, err.Error())
		}
		r.CreatedAt = created.Format(time.RFC3339)
		out = append(out, r)
	}
	return c.JSON(out)
}

func (s *Server) Rewards(c *fiber.Ctx) error {
	rows, err := s.DB.QueryContext(c.UserContext(), `
		SELECT id, title, type, points_cost, COALESCE(partner,''), status, created_at
		FROM rewards_catalog
		ORDER BY id ASC
	`)
	if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	type reward struct {
		ID         int64  `json:"id"`
		Title      string `json:"title"`
		Type       string `json:"type"`
		PointsCost int64  `json:"points_cost"`
		Partner    string `json:"partner"`
		Status     string `json:"status"`
		CreatedAt  string `json:"created_at"`
	}
	var out []reward
	for rows.Next() {
		var r reward
		var created time.Time
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.PointsCost, &r.Partner, &r.Status, &created); err != nil {
			return jsonErr(c, fiber.StatusInternalServerError, err.Error())
		}
		r.CreatedAt = created.Format(time.RFC3339)
		out = append(out, r)
	}
	return c.JSON(out)
}

func (s *Server) Redeem(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return jsonErr(c, fiber.StatusUnauthorized, "unauthorized")
	}
	var body struct {
		RewardID int64 `json:"reward_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.RewardID <= 0 {
		return jsonErr(c, fiber.StatusBadRequest, "invalid body")
	}

	ctx := c.UserContext()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	var cost int64
	var status string
	if err := tx.QueryRowContext(ctx, `
		SELECT points_cost, status FROM rewards_catalog WHERE id = $1
	`, body.RewardID).Scan(&cost, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return jsonErr(c, fiber.StatusNotFound, "reward not found")
		}
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}
	if strings.ToUpper(status) != "ACTIVE" {
		return jsonErr(c, fiber.StatusBadRequest, "reward not active")
	}

	var bal int64
	err = tx.QueryRowContext(ctx, `
		SELECT points_total FROM points_balance WHERE user_id = $1 FOR UPDATE
	`, userID).Scan(&bal)
	if errors.Is(err, sql.ErrNoRows) {
		bal = 0
	} else if err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	if bal < cost {
		return jsonErr(c, fiber.StatusBadRequest, "insufficient points")
	}

	var redemptionID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO redemptions (user_id, reward_id, points_spent, status)
		VALUES ($1, $2, $3, 'REQUESTED')
		RETURNING id
	`, userID, body.RewardID, cost).Scan(&redemptionID); err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	if err := insertLedger(ctx, tx, userID, nil, -int(cost), "redemption"); err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	if err := updateBalance(ctx, tx, userID, -cost); err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return jsonErr(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"ok":            true,
		"redemption_id": redemptionID,
		"status":        "REQUESTED",
		"points_spent":  cost,
	})
}

// Helpers

func getUserID(c *fiber.Ctx) string {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

type tier struct {
	Name       string
	Min        int64
	Multiplier float64
}

func loadTiers(ctx context.Context, db *sql.DB) ([]tier, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT tier_name, min_points, multiplier
		FROM tiers
		ORDER BY min_points ASC
	`)
	if err != nil {
		return []tier{{Name: "STONE", Min: 0, Multiplier: 1.0}, {Name: "SILVER", Min: 2000, Multiplier: 1.05}, {Name: "OBSIDIAN", Min: 10000, Multiplier: 1.1}}, nil
	}
	defer rows.Close()
	var out []tier
	for rows.Next() {
		var t tier
		if err := rows.Scan(&t.Name, &t.Min, &t.Multiplier); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		out = []tier{{Name: "STONE", Min: 0, Multiplier: 1.0}, {Name: "SILVER", Min: 2000, Multiplier: 1.05}, {Name: "OBSIDIAN", Min: 10000, Multiplier: 1.1}}
	}
	return out, nil
}

func awardPointsTx(ctx context.Context, tx *sql.Tx, userID string, sourceTxnID string, points int64, reason string) error {
	if points <= 0 {
		return nil
	}
	if err := insertLedger(ctx, tx, userID, &sourceTxnID, int(points), reason); err != nil {
		return err
	}
	return updateBalance(ctx, tx, userID, points)
}

func insertLedger(ctx context.Context, tx *sql.Tx, userID string, sourceTxnID *string, delta int, reason string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO points_ledger (user_id, source_txn_id, points_delta, reason)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, source_txn_id, reason) WHERE source_txn_id IS NOT NULL AND reason = 'earn' DO NOTHING
	`, userID, sourceTxnID, delta, reason)
	return err
}

func updateBalance(ctx context.Context, tx *sql.Tx, userID string, delta int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO points_balance (user_id, points_total, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id) DO UPDATE
		SET points_total = points_balance.points_total + EXCLUDED.points_total,
		    updated_at = now()
	`, userID, delta)
	return err
}

func jsonErr(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(fiber.Map{"error": msg})
}
