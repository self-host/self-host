-- name: ExistsDataset :one
SELECT COUNT(*) AS count
FROM datasets
WHERE datasets.uuid = sqlc.arg(uuid);

-- name: CreateDataset :one
WITH ds AS (
	INSERT INTO datasets (name, format, content, size, belongs_to, created_by, updated_by, tags)
	VALUES(
		sqlc.arg(name)::text,
		sqlc.arg(format)::text,
		sqlc.arg(content)::bytea,
		length(sqlc.arg(content))::integer,
		NULLIF(sqlc.arg(belongs_to)::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
		sqlc.arg(created_by)::uuid,
		sqlc.arg(created_by)::uuid,
		sqlc.arg(tags)
	)
	RETURNING
		uuid,
		name,
		format,
		size,
		belongs_to,
		created,
		updated,
		created_by,
		updated_by,
		tags
), grp AS (
	SELECT groups.uuid
	FROM groups, user_groups
	WHERE user_groups.group_uuid = groups.uuid
	AND user_groups.user_uuid = (SELECT created_by FROM ds)
	AND groups.uuid = (
		SELECT users.uuid
		FROM users
		WHERE users.name = groups.name
	)
	LIMIT 1
), grp_policies AS (
	INSERT INTO group_policies(group_uuid, priority, effect, action, resource)
	VALUES (
		(SELECT uuid FROM grp), 0, 'allow', 'create','datasets/'||(SELECT uuid FROM ds)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'read','datasets/'||(SELECT uuid FROM ds)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'update','datasets/'||(SELECT uuid FROM ds)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'delete','datasets/'||(SELECT uuid FROM ds)||'/%'
	)
)
SELECT *
FROM ds LIMIT 1;

-- name: FindDatasets :many
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
SELECT
	uuid,
	name,
	format,
	size,
	belongs_to,
	created,
	updated,
	created_by,
	updated_by,
	tags
FROM datasets
WHERE 'datasets/'||datasets.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT
	uuid,
	name,
	format,
	size,
	belongs_to,
	created,
	updated,
	created_by,
	updated_by,
	tags
FROM datasets
WHERE 'datasets/'||datasets.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindDatasetsByTags :many
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
SELECT
	uuid,
	name,
	format,
	size,
	belongs_to,
	created,
	updated,
	created_by,
	updated_by,
	tags
FROM datasets
WHERE 'datasets/'||datasets.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
AND sqlc.arg(tags) && datasets.tags
EXCEPT
SELECT
	uuid,
	name,
	format,
	size,
	belongs_to,
	created,
	updated,
	created_by,
	updated_by,
	tags
FROM datasets
WHERE 'datasets/'||datasets.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
AND sqlc.arg(tags) && datasets.tags
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindDatasetByUUID :one
SELECT
	uuid,
	name,
	format,
	size,
	belongs_to,
	created,
	updated,
	created_by,
	updated_by,
	tags
FROM datasets
WHERE datasets.uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: FindDatasetByThing :many
SELECT
	uuid,
	name,
	format,
	size,
	belongs_to,
	created,
	updated,
	created_by,
	updated_by,
	tags
FROM datasets
WHERE datasets.belongs_to = sqlc.arg(thing_uuid)
ORDER BY name
;

-- name: GetDatasetContentByUUID :one
SELECT format, content
FROM datasets
WHERE datasets.uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: SetDatasetNameByUUID :execrows
UPDATE datasets
SET name = sqlc.arg(name)
WHERE datasets.uuid = sqlc.arg(uuid);

-- name: SetDatasetFormatByUUID :execrows
UPDATE datasets
SET format = sqlc.arg(format)
WHERE datasets.uuid = sqlc.arg(uuid);

-- name: SetDatasetContentByUUID :execrows
UPDATE datasets
SET content = sqlc.arg(content)
WHERE datasets.uuid = sqlc.arg(uuid);

-- name: SetDatasetTags :execrows
UPDATE datasets
SET tags = sqlc.arg(tags)
WHERE datasets.uuid = sqlc.arg(uuid);

-- name: DeleteDataset :execrows
DELETE FROM datasets
WHERE datasets.uuid = sqlc.arg(uuid);
