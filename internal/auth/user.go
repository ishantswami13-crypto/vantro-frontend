package auth

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// UserIDFromRequest extracts the user ID from the temporary X-User-Id header.
// Example header: X-User-Id: 11111111-1111-1111-1111-111111111111
func UserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := r.Header.Get("X-User-Id")
	if raw == "" {
		return uuid.Nil, errors.New("missing X-User-Id header")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid X-User-Id UUID")
	}
	return id, nil
}
