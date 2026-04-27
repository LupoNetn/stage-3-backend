-- name: CreateUser :one
INSERT INTO users (
    id, github_id, username, email, avatar_url, role, is_active, last_login_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetUserByGithubID :one
SELECT * FROM users
WHERE github_id = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: UpdateRefreshToken :exec
UPDATE users
SET refresh_token = $2
WHERE id = $1;
