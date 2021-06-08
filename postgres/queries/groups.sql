-- name: ExistsGroup :one
SELECT COUNT(*) AS count
FROM groups
WHERE groups.uuid = sqlc.arg(uuid);

-- name: CreateGroup :one
INSERT INTO groups(name)
VALUES(sqlc.arg(name))
RETURNING *;

-- name: FindGroups :many
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
)
SELECT *
FROM groups
WHERE 'groups/'||groups.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT *
FROM groups
WHERE 'groups/'||groups.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindGroupByUuid :one
SELECT * FROM groups
WHERE uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: FindGroupsByUser :many
SELECT groups.*
FROM groups, user_groups
WHERE groups.uuid = user_groups.group_uuid
AND user_groups.user_uuid = sqlc.arg(uuid);

-- name: DeleteGroup :execrows
DELETE FROM groups
WHERE groups.uuid = sqlc.arg(uuid);

-- name: SetGroupNameByUUID :execrows
UPDATE groups
SET name = sqlc.arg(name)
WHERE groups.uuid = sqlc.arg(uuid);
