-- name: CheckUserTokenHasAccess :one
WITH usr AS (
  SELECT users.uuid
  FROM users, user_tokens
  WHERE user_tokens.user_uuid = users.uuid
  AND user_tokens.token_hash = sha256(sqlc.arg(token))
  LIMIT 1
) SELECT COALESCE((SELECT user_has_access(
  usr.uuid,
  sqlc.arg(action)::policy_action,
  sqlc.arg(resource)::TEXT
) FROM usr LIMIT 1), false)::boolean AS access;


