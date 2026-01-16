package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Entry struct {
	UserID     *string
	Action     string
	EntityType string
	EntityID   *string
	IP         *string
	UserAgent  *string
	Metadata   []byte
}

// Write records an audit entry; failures are returned so callers can ignore if needed.
func Write(ctx context.Context, db *pgxpool.Pool, e Entry) error {
	if db == nil {
		return nil
	}

	var metadata interface{}
	if len(e.Metadata) > 0 {
		raw := json.RawMessage(e.Metadata)
		metadata = raw
	}

	_, err := db.Exec(ctx, `
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, ip, user_agent, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, e.UserID, e.Action, e.EntityType, e.EntityID, e.IP, e.UserAgent, metadata)

	return err
}
