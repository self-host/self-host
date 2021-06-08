-- name: CreateUserToken :one
INSERT INTO user_tokens(user_uuid, name, token_hash)
VALUES(
  sqlc.arg(user_uuid),
  sqlc.arg(name),
  sha256(sqlc.arg(token)::bytea)
)
RETURNING *;
