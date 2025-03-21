// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: resfresh_tokens.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :one
WITH revoke_tokens AS (
    UPDATE refresh_tokens
    SET 
        revoked_at = NOW(),
        updated_at = NOW()
    WHERE user_id = $2 
        AND revoked_at IS NULL
)
INSERT INTO refresh_tokens (
    token,
    user_id,
    created_at,
    updated_at,
    expires_at,
    revoked_at
) VALUES (
    $1,
    $2,
    NOW(),
    NOW(),
    NOW() + INTERVAL '60 days',
    NULL
) RETURNING token, user_id, created_at, updated_at, expires_at, revoked_at
`

type CreateRefreshTokenParams struct {
	Token  string
	UserID uuid.UUID
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, createRefreshToken, arg.Token, arg.UserID)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const getUserFromRefreshToken = `-- name: GetUserFromRefreshToken :one
SELECT token, user_id, created_at, updated_at, expires_at, revoked_at FROM refresh_tokens WHERE token = $1 AND revoked_at IS NULL
`

func (q *Queries) GetUserFromRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getUserFromRefreshToken, token)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const revokeRefreshTokens = `-- name: RevokeRefreshTokens :exec
UPDATE refresh_tokens
  SET revoked_at = NOW(), updated_at = NOW()
  WHERE user_id = $1
`

func (q *Queries) RevokeRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, revokeRefreshTokens, userID)
	return err
}
