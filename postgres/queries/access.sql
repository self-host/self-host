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

-- name: CheckUserTokenHasAccessMany :one
WITH usr AS (
	SELECT users.uuid
	FROM users, user_tokens
	WHERE user_tokens.user_uuid = users.uuid
	AND user_tokens.token_hash = sha256(sqlc.arg(token))
	LIMIT 1
), usr_r AS (
	SELECT
		usr.uuid,
		sqlc.arg(action)::policy_action AS action,
		unnest((SELECT sqlc.arg(resources)::TEXT[]))::TEXT AS resource
	FROM usr
)
SELECT
	COALESCE(true = ALL (array_agg(user_has_access(usr_r.uuid, usr_r.action, usr_r.resource))), false)::boolean AS access
FROM usr_r;
