package reports

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type Store struct {
	DB *sql.DB
}

var ErrNotFound = errors.New("not found")

func newToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Store) Create(ctx context.Context, phone, month, filePath string, ttl time.Duration) (string, time.Time, error) {
	token, err := newToken(24)
	if err != nil {
		return "", time.Time{}, err
	}
	expires := time.Now().Add(ttl)

	const q = `
		INSERT INTO reports (user_phone, month, token, file_path, expires_at)
		VALUES ($1, $2, $3, $4, $5);
	`
	if _, err := s.DB.ExecContext(ctx, q, phone, month, token, filePath, expires); err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (s *Store) GetByToken(ctx context.Context, token string) (string, time.Time, error) {
	const q = `SELECT file_path, expires_at FROM reports WHERE token = $1;`

	var path string
	var exp time.Time
	if err := s.DB.QueryRowContext(ctx, q, token).Scan(&path, &exp); err != nil {
		if err == sql.ErrNoRows {
			return "", time.Time{}, ErrNotFound
		}
		return "", time.Time{}, err
	}
	return path, exp, nil
}
