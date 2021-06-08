-- name: ExistsTimeseries :one
SELECT COUNT(*) AS count
FROM timeseries
WHERE timeseries.uuid = sqlc.arg(uuid);

-- name: CreateTimeseries :one
WITH t AS (
	INSERT INTO timeseries(
		thing_uuid,
		name,
		si_unit,
		lower_bound,
		upper_bound,
		created_by,
		tags
	) VALUES (
		NULLIF(sqlc.arg(thing_uuid)::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
		sqlc.arg(name),
		sqlc.arg(si_unit),
		sqlc.arg(lower_bound),
		sqlc.arg(upper_bound),
		sqlc.arg(created_by),
		sqlc.arg(tags)
	) RETURNING *
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
		(SELECT uuid FROM grp), 0, 'allow', 'create','timeseries/'||(SELECT uuid FROM t)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'read','timeseries/'||(SELECT uuid FROM t)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'update','timeseries/'||(SELECT uuid FROM t)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'delete','timeseries/'||(SELECT uuid FROM t)||'/%'
	)
)
SELECT *
FROM t LIMIT 1;

-- name: GetTimeseriesByUUID :one
SELECT * FROM timeseries
WHERE uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: GetUnitFromTimeseries :one
SELECT si_unit FROM timeseries
WHERE uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: FindTimeseries :many
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
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT *
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindTimeseriesByTags :many
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
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
AND sqlc.arg(tags) && timeseries.tags
EXCEPT
SELECT *
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
AND sqlc.arg(tags) && timeseries.tags
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindTimeseriesByThing :many
SELECT * FROM timeseries
WHERE sqlc.arg(thing_uuid) = timeseries.thing_uuid
ORDER BY name
;

-- name: FindTimeseriesByUUID :one
SELECT * FROM timeseries
WHERE sqlc.arg(ts_uuid) = timeseries.uuid
LIMIT 1;

-- name: SetTimeseriesThing :execrows
UPDATE timeseries
SET thing_uuid = sqlc.arg(thing_uuid)
WHERE timeseries.uuid = sqlc.arg(uuid);

-- name: SetTimeseriesName :execrows
UPDATE timeseries
SET name = sqlc.arg(name)
WHERE timeseries.uuid = sqlc.arg(uuid);

-- name: SetTimeseriesSiUnit :execrows
UPDATE timeseries
SET si_unit = sqlc.arg(si_unit)
WHERE timeseries.uuid = sqlc.arg(uuid);

-- name: SetTimeseriesLowerBound :execrows
UPDATE timeseries
SET lower_bound = sqlc.arg(lower_bound)
WHERE timeseries.uuid = sqlc.arg(uuid);

-- name: SetTimeseriesUpperBound :execrows
UPDATE timeseries
SET upper_bound = sqlc.arg(upper_bound)
WHERE timeseries.uuid = sqlc.arg(uuid);

-- name: SetTimeseriesTags :execrows
UPDATE timeseries
SET tags = sqlc.arg(tags)
WHERE timeseries.uuid = sqlc.arg(uuid);

-- name: DeleteTimeseries :execrows
DELETE FROM timeseries
WHERE uuid = sqlc.arg(uuid);
