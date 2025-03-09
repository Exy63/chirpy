-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
WHERE user_id = COALESCE(sqlc.narg('user_id'), user_id)
ORDER BY created_at COALESCE(sqlc.narg('sort'), ASC);

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;