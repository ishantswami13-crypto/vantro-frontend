package points

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ishantswami13-crypto/vantro-backend/internal/audit"
)

type Handler struct {
	Pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{Pool: pool}
}

func getUserID(c *fiber.Ctx) string {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func (h *Handler) PointsSummary(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	summary, err := GetPointsSummary(c.UserContext(), h.Pool, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load points")
	}

	return c.JSON(summary)
}

func (h *Handler) PointsLedger(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	rows, err := h.Pool.Query(c.UserContext(), `
SELECT id, points_delta, reason, source_txn_id, created_at::text
FROM points_ledger
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 50
`, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load ledger")
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var (
			id          int64
			delta       int
			reason      string
			sourceTxnID *string
			createdAt   string
		)
		if err := rows.Scan(&id, &delta, &reason, &sourceTxnID, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read ledger")
		}
		items = append(items, fiber.Map{
			"id":            id,
			"points_delta":  delta,
			"reason":        reason,
			"source_txn_id": sourceTxnID,
			"created_at":    createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to read ledger")
	}

	return c.JSON(items)
}

func (h *Handler) Rewards(c *fiber.Ctx) error {
	rows, err := h.Pool.Query(c.UserContext(), `
SELECT id, title, type, points_cost, partner, status, created_at::text
FROM rewards_catalog
ORDER BY id ASC
`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load rewards")
	}
	defer rows.Close()

	type reward struct {
		ID         int64   `json:"id"`
		Title      string  `json:"title"`
		Type       string  `json:"type"`
		PointsCost int64   `json:"points_cost"`
		Partner    *string `json:"partner,omitempty"`
		Status     string  `json:"status"`
		CreatedAt  string  `json:"created_at"`
	}

	var out []reward
	for rows.Next() {
		var r reward
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.PointsCost, &r.Partner, &r.Status, &r.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read rewards")
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to read rewards")
	}

	return c.JSON(out)
}

func (h *Handler) Redeem(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	idemKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	bodyBytes := c.Body()

	var body struct {
		RewardID int64 `json:"reward_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if body.RewardID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "reward_id required")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	requestHash := ""
	if idemKey != "" {
		sum := sha256.Sum256(append([]byte(c.Method()+" "+c.Path()+" "), bodyBytes...))
		requestHash = hex.EncodeToString(sum[:])

		var statusCode int
		var storedBody string
		err := h.Pool.QueryRow(
			ctx,
			`SELECT response_status, response_body
             FROM idempotency_keys
             WHERE owner_id = $1 AND idempotency_key = $2`,
			userID, idemKey,
		).Scan(&statusCode, &storedBody)
		if err == nil {
			c.Status(statusCode)
			c.Type("application/json")
			return c.SendString(storedBody)
		}
		if err != nil && err != pgx.ErrNoRows {
			// fall through on unexpected error
		}
	}

	// fetch cost to pass
	var cost int64
	var status string
	err := h.Pool.QueryRow(ctx, `SELECT points_cost, status FROM rewards_catalog WHERE id = $1`, body.RewardID).
		Scan(&cost, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "reward not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load reward")
	}
	if strings.ToUpper(status) != "ACTIVE" {
		return fiber.NewError(fiber.StatusBadRequest, "reward not active")
	}

	redemptionID, err := SpendPoints(ctx, h.Pool, userID, cost, body.RewardID)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient") {
			return fiber.NewError(fiber.StatusBadRequest, "insufficient points")
		}
		if strings.Contains(err.Error(), "reward not active") {
			return fiber.NewError(fiber.StatusBadRequest, "reward not active")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "redeem failed: "+err.Error())
	}

	resp := fiber.Map{
		"ok":            true,
		"redemption_id": redemptionID,
		"status":        "REQUESTED",
		"points_spent":  cost,
	}

	// Best-effort audit
	ip := strings.TrimSpace(c.IP())
	ua := strings.TrimSpace(c.Get("User-Agent"))
	entry := audit.Entry{
		UserID:     &userID,
		Action:     "reward_redeem",
		EntityType: "reward",
		EntityID:   nil,
		Metadata:   bodyBytes,
	}
	if redemptionID > 0 {
		idStr := strconv.FormatInt(redemptionID, 10)
		entry.EntityID = &idStr
	}
	if ip != "" {
		entry.IP = &ip
	}
	if ua != "" {
		entry.UserAgent = &ua
	}
	_ = audit.Write(ctx, h.Pool, entry)

	if idemKey != "" && requestHash != "" {
		if buf, mErr := json.Marshal(resp); mErr == nil {
			_, _ = h.Pool.Exec(
				ctx,
				`INSERT INTO idempotency_keys (owner_id, endpoint, idempotency_key, request_hash, response_status, response_body)
                 VALUES ($1, $2, $3, $4, $5, $6)
                 ON CONFLICT (owner_id, idempotency_key) DO NOTHING`,
				userID,
				"/redeem",
				idemKey,
				requestHash,
				fiber.StatusOK,
				string(buf),
			)
		}
	}

	return c.JSON(resp)
}
