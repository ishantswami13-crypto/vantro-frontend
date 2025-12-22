package points

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Summary struct {
	PointsTotal        int64   `json:"points_total"`
	Tier               string  `json:"tier"`
	Multiplier         float64 `json:"multiplier"`
	NextTier           string  `json:"next_tier"`
	NextTierMinPoints  int64   `json:"next_tier_min_points"`
	ProgressToNextTier float64 `json:"progress_to_next"`
}

type tierRow struct {
	Name       string
	MinPoints  int64
	Multiplier float64
}

func AwardPointsForTransaction(ctx context.Context, db *pgxpool.Pool, userID string, sourceTxnID *string, amount int64, reason string) (int64, error) {
	if amount <= 0 {
		return 0, nil
	}

	basePoints := amount / 100
	if basePoints <= 0 {
		return 0, nil
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var current int64
	err = tx.QueryRow(ctx, `SELECT points_total FROM points_balance WHERE user_id = $1 FOR UPDATE`, userID).Scan(&current)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		current = 0
	}

	tier, multiplier, err := lookupTier(ctx, tx, current)
	if err != nil {
		return 0, err
	}
	_ = tier // tier not used further for now

	pointsAwarded := int64(math.Floor(float64(basePoints) * multiplier))
	if pointsAwarded == 0 {
		return 0, nil
	}

	_, err = tx.Exec(ctx, `
INSERT INTO points_ledger (user_id, source_txn_id, points_delta, reason, created_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (user_id, source_txn_id) WHERE source_txn_id IS NOT NULL DO NOTHING
`, userID, sourceTxnID, pointsAwarded, reason)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO points_balance (user_id, points_total, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id) DO UPDATE
SET points_total = points_balance.points_total + EXCLUDED.points_total,
    updated_at = NOW()
`, userID, pointsAwarded)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return pointsAwarded, nil
}

func lookupTier(ctx context.Context, tx pgx.Tx, points int64) (string, float64, error) {
	rows, err := tx.Query(ctx, `SELECT tier_name, min_points, multiplier FROM tiers ORDER BY min_points DESC`)
	if err != nil {
		return "STONE", 1.0, err
	}
	defer rows.Close()

	tier := "STONE"
	mult := 1.0
	for rows.Next() {
		var name string
		var minPts int64
		var m float64
		if err := rows.Scan(&name, &minPts, &m); err != nil {
			return "STONE", 1.0, err
		}
		if points >= minPts {
			tier = name
			mult = m
			break
		}
	}
	return tier, mult, nil
}

func GetPointsSummary(ctx context.Context, db *pgxpool.Pool, userID string) (Summary, error) {
	var total int64
	err := db.QueryRow(ctx, `SELECT points_total FROM points_balance WHERE user_id = $1`, userID).Scan(&total)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Summary{}, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		total = 0
	}

	tierRows, err := loadTiers(ctx, db)
	if err != nil {
		return Summary{}, err
	}

	currentTier := tierRows[0]
	nextTier := tierRows[len(tierRows)-1]
	found := false
	for i, t := range tierRows {
		if total >= t.MinPoints {
			currentTier = t
			if i+1 < len(tierRows) {
				nextTier = tierRows[i+1]
			} else {
				nextTier = t
			}
			found = true
			break
		}
	}
	if !found {
		currentTier = tierRows[len(tierRows)-1]
		nextTier = currentTier
	}

	progress := 1.0
	if nextTier.MinPoints > currentTier.MinPoints {
		progress = float64(total-currentTier.MinPoints) / float64(nextTier.MinPoints-currentTier.MinPoints)
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
	}

	return Summary{
		PointsTotal:        total,
		Tier:               currentTier.Name,
		Multiplier:         currentTier.Multiplier,
		NextTier:           nextTier.Name,
		NextTierMinPoints:  nextTier.MinPoints,
		ProgressToNextTier: progress,
	}, nil
}

func loadTiers(ctx context.Context, db *pgxpool.Pool) ([]tierRow, error) {
	rows, err := db.Query(ctx, `SELECT tier_name, min_points, multiplier FROM tiers ORDER BY min_points ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []tierRow
	for rows.Next() {
		var t tierRow
		if err := rows.Scan(&t.Name, &t.MinPoints, &t.Multiplier); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		out = append(out, tierRow{Name: "STONE", MinPoints: 0, Multiplier: 1.0})
	}
	// ensure descending order for lookup convenience
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func SpendPoints(ctx context.Context, db *pgxpool.Pool, userID string, pointsToSpend int64, rewardID int64) (int64, error) {
	if pointsToSpend <= 0 {
		return 0, errors.New("points_to_spend must be > 0")
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var total int64
	err = tx.QueryRow(ctx, `SELECT points_total FROM points_balance WHERE user_id = $1 FOR UPDATE`, userID).Scan(&total)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		total = 0
	}

	var cost int64
	var status string
	err = tx.QueryRow(ctx, `SELECT points_cost, status FROM rewards_catalog WHERE id = $1`, rewardID).Scan(&cost, &status)
	if err != nil {
		return 0, err
	}
	if strings.ToUpper(status) != "ACTIVE" {
		return 0, errors.New("reward not active")
	}

	if total < cost {
		return 0, errors.New("insufficient points")
	}

	var redemptionID int64
	err = tx.QueryRow(ctx, `
INSERT INTO redemptions (user_id, reward_id, points_spent, status, created_at)
VALUES ($1, $2, $3, 'REQUESTED', NOW())
RETURNING id
`, userID, rewardID, cost).Scan(&redemptionID)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO points_ledger (user_id, source_txn_id, points_delta, reason, created_at)
VALUES ($1, NULL, $2, $3, NOW())
`, userID, -cost, "redeem_reward")
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 0, errors.New("insufficient points")
	}

	_, err = tx.Exec(ctx, `
UPDATE points_balance
SET points_total = points_total - $2,
    updated_at = NOW()
WHERE user_id = $1
`, userID, cost)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return redemptionID, nil
}
