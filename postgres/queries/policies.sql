-- name: ExistsPolicy :one
SELECT COUNT(*) AS count
FROM group_policies
WHERE group_policies.uuid = sqlc.arg(uuid);

-- name: CreatePolicy :one
INSERT INTO group_policies(group_uuid, priority, effect, action, resource)
VALUES (
	sqlc.arg(group_uuid),
	sqlc.arg(priority),
	sqlc.arg(effect),
	sqlc.arg(action),
	sqlc.arg(resource)
)
RETURNING *;

-- name: FindPolicies :many
WITH usr AS (
	SELECT users.uuid
	FROM users, user_tokens
	WHERE user_tokens.user_uuid = users.uuid
	AND user_tokens.token_hash = sha256(sqlc.arg(token))
	LIMIT 1
), policies AS (
	SELECT group_policies.effect, group_policies.priority, group_policies.resource
	FROM group_policies, user_groups
	WHERE user_groups.group_uuid = group_policies.group_uuid
	AND user_groups.user_uuid = (SELECT uuid FROM usr)
	AND action = 'read'
), f_group_policies AS (
	SELECT * FROM group_policies
	WHERE
		sqlc.arg(group_uuids)::uuid[] IS NULL
	OR
		group_policies.group_uuid = ANY(sqlc.arg(group_uuids)::uuid[])
)
SELECT *
FROM f_group_policies
WHERE 'policies/'||f_group_policies.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT *
FROM f_group_policies
WHERE 'policies/'||f_group_policies.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY resource DESC, effect ASC, action DESC, priority ASC
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindPolicyByUUID :one
SELECT *
FROM group_policies
WHERE group_policies.uuid = sqlc.arg(uuid);

-- name: SetPolicyGroup :execrows
UPDATE group_policies
SET group_uuid = sqlc.arg(group_uuid)
WHERE group_policies.uuid = sqlc.arg(uuid);

-- name: SetPolicyPriority :execrows
UPDATE group_policies
SET priority = sqlc.arg(priority)
WHERE group_policies.uuid = sqlc.arg(uuid);

-- name: SetPolicyEffect :execrows
UPDATE group_policies
SET effect = sqlc.arg(effect)
WHERE group_policies.uuid = sqlc.arg(uuid);

-- name: SetPolicyAction :execrows
UPDATE group_policies
SET action = sqlc.arg(action)
WHERE group_policies.uuid = sqlc.arg(uuid);

-- name: SetPolicyResource :execrows
UPDATE group_policies
SET resource = sqlc.arg(resource)
WHERE group_policies.uuid = sqlc.arg(uuid);

-- name: DeletePolicyByUUID :execrows
DELETE
FROM group_policies
WHERE group_policies.uuid = sqlc.arg(uuid);
