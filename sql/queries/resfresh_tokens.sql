-- name: CreateRefreshToken :one
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
) RETURNING *;


-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1 AND revoked_at IS NULL;


-- name: RevokeRefreshTokens :exec
UPDATE refresh_tokens
  SET revoked_at = NOW(), updated_at = NOW()
  WHERE user_id = $1;
