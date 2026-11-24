package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
)

const idempotencyTTL = 24 * time.Hour

// idempotent runs fn in a transaction and stores the JSON response under key.
// A replay with the same key and request hash returns the stored value; a
// different hash for the same key is a conflict. An empty key disables it.
func idempotent[T any](
	ctx context.Context,
	db *sql.DB,
	key, requestHash string,
	status int,
	fn func(*sql.Tx) (T, error),
) (T, bool, error) {
	var zero T

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return zero, false, err
	}
	defer tx.Rollback()

	if key != "" {
		var storedHash, response string
		var storedStatus int
		cutoff := formatTime(time.Now().UTC().Add(-idempotencyTTL))
		err := tx.QueryRowContext(ctx, `
			SELECT request_hash, status, response
			FROM idempotency_keys
			WHERE key = ? AND created_at >= ?
		`, key, cutoff).Scan(&storedHash, &storedStatus, &response)
		switch {
		case err == nil:
			if storedHash != requestHash {
				return zero, false, domain.ErrConflict
			}
			var value T
			if err := json.Unmarshal([]byte(response), &value); err != nil {
				return zero, false, err
			}
			return value, true, nil
		case errors.Is(err, sql.ErrNoRows):
			// First request with this key.
		default:
			return zero, false, err
		}
	}

	value, err := fn(tx)
	if err != nil {
		return zero, false, err
	}

	if key != "" {
		body, err := json.Marshal(value)
		if err != nil {
			return zero, false, err
		}
		now := time.Now().UTC()
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM idempotency_keys WHERE created_at < ?`,
			formatTime(now.Add(-idempotencyTTL)),
		); err != nil {
			return zero, false, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO idempotency_keys (key, request_hash, status, response, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, key, requestHash, status, string(body), formatTime(now)); err != nil {
			if isUnique(err) {
				return zero, false, domain.ErrConflict
			}
			return zero, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return zero, false, err
	}
	return value, false, nil
}
