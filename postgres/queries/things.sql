-- name: ExistsThing :one
SELECT COUNT(*) AS count
FROM things
WHERE things.uuid = sqlc.arg(uuid);

-- name: CreateThing :one
WITH t AS (
	INSERT INTO things (
		name, type, created_by, tags
	) VALUES (
		sqlc.arg(name),
		sqlc.arg(type),
		sqlc.arg(created_by),
		sqlc.arg(tags)
	)
	RETURNING *
), grp AS (
	SELECT groups.uuid
	FROM groups, user_groups
	WHERE user_groups.group_uuid = groups.uuid
	AND user_groups.user_uuid = (SELECT created_by FROM t)
	AND groups.uuid = (
		SELECT users.uuid
		FROM users
		WHERE users.name = groups.name
	)
	LIMIT 1
), grp_policies AS (
	INSERT INTO group_policies(group_uuid, priority, effect, action, resource)
	VALUES (
		(SELECT uuid FROM grp), 0, 'allow', 'create','things/'||(SELECT uuid FROM t)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'read','things/'||(SELECT uuid FROM t)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'update','things/'||(SELECT uuid FROM t)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'delete','things/'||(SELECT uuid FROM t)||'/%'
	)
)
SELECT *
FROM t LIMIT 1;

-- name: FindThingByUUID :one
SELECT *
FROM things
WHERE things.uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: FindThings :many
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
FROM things
WHERE 'things/'||things.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT *
FROM things
WHERE 'things/'||things.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: SetThingNameByUUID :execrows
UPDATE things
SET name = sqlc.arg(name)
WHERE things.uuid = sqlc.arg(uuid);

-- name: SetThingTypeByUUID :execrows
UPDATE things
SET type = sqlc.arg(type)
WHERE things.uuid = sqlc.arg(uuid);

-- name: SetThingStateByUUID :execrows
UPDATE things
SET state = sqlc.arg(state)
WHERE things.uuid = sqlc.arg(uuid);

-- name: SetThingTags :execrows
UPDATE things
SET tags = sqlc.arg(tags)
WHERE things.uuid = sqlc.arg(uuid);

-- name: DeleteThing :execrows
DELETE FROM things
WHERE things.uuid = sqlc.arg(uuid);
