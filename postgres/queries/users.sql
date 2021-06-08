-- name: ExistsUser :one
SELECT COUNT(*) AS count
FROM users
WHERE users.uuid = sqlc.arg(uuid);

-- name: CreateUser :one
WITH grp AS (
	INSERT INTO groups(name)
	VALUES(sqlc.arg(name))
	RETURNING *
), usr AS (
	INSERT INTO users(name)
	VALUES(sqlc.arg(name))
	RETURNING *
), grp_policies AS (
	INSERT INTO group_policies(group_uuid, priority, effect, action, resource)
	VALUES (
		(SELECT uuid FROM grp), 0, 'allow', 'read','users/me'
	), (

		(SELECT uuid FROM grp), 0, 'allow', 'create','users/'||(SELECT uuid FROM usr)
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'read','users/'||(SELECT uuid FROM usr)
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'update','users/'||(SELECT uuid FROM usr)
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'delete','users/'||(SELECT uuid FROM usr)
	), (

		(SELECT uuid FROM grp), 0, 'allow', 'create','users/'||(SELECT uuid FROM usr)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'read','users/'||(SELECT uuid FROM usr)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'update','users/'||(SELECT uuid FROM usr)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'delete','users/'||(SELECT uuid FROM usr)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'deny', 'update','users/'||(SELECT uuid FROM usr)||'/rate'
	)
), usrgrp AS (
	INSERT INTO user_groups(user_uuid, group_uuid)
	SELECT usr.uuid, grp.uuid
	FROM usr, grp
)
SELECT * FROM usr;

-- name: AddTokenToUser :one
INSERT INTO user_tokens(user_uuid, name, token_hash)
VALUES (sqlc.arg(user_uuid), sqlc.arg(name), sha256(sqlc.arg(secret)))
RETURNING *;

-- name: FindUsers :many
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
), partial_users AS (
	SELECT *
	FROM users
	WHERE 'users/'||users.uuid LIKE ANY(
		(SELECT resource FROM policies WHERE effect = 'allow')
	)
	EXCEPT
	SELECT *
	FROM users
	WHERE 'users/'||users.uuid LIKE ANY(
		(SELECT resource FROM policies WHERE effect = 'deny')
	)
	ORDER BY name
	LIMIT sqlc.arg(arg_limit)::BIGINT
	OFFSET sqlc.arg(arg_offset)::BIGINT
)
SELECT partial_users.*, (COALESCE((
	SELECT json_agg(json_build_object('uuid', groups.uuid, 'name', groups.name))
	FROM user_groups, groups
	WHERE groups.uuid = user_groups.group_uuid
	AND partial_users.uuid = user_groups.user_uuid
), '[]')::text) AS groups
FROM partial_users;

-- name: FindUserByUUID :one
SELECT *
FROM users
WHERE users.uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: FindTokensByUser :many
SELECT uuid, name
FROM user_tokens
WHERE user_tokens.user_uuid = sqlc.arg(uuid);

-- name: GetUserUuidFromToken :one
SELECT user_tokens.user_uuid AS uuid
FROM user_tokens
WHERE user_tokens.token_hash = sha256(sqlc.arg(token))
LIMIT 1;

-- name: AddUserToGroup :exec
INSERT INTO user_groups(user_uuid, group_uuid)
VALUES(
	sqlc.arg(user_uuid)::uuid,
	sqlc.arg(group_uuid)::uuid
);

-- name: RemoveUserFromGroups :execrows
DELETE FROM user_groups
WHERE user_uuid = sqlc.arg(user_uuid)
AND group_uuid = ANY(sqlc.arg(group_uuids)::uuid[]);

-- name: RemoveUserFromAllGroups :execrows
DELETE FROM user_groups
WHERE user_uuid = sqlc.arg(user_uuid);

-- name: DeleteUser :execrows
WITH grp AS (
	DELETE FROM groups
	WHERE name = (
		SELECT name
		FROM users
		WHERE users.uuid = sqlc.arg(uuid)
	)
)
DELETE FROM users
WHERE users.uuid = sqlc.arg(uuid);

-- name: DeleteTokenFromUser :execrows
DELETE FROM user_tokens
WHERE user_tokens.uuid = sqlc.arg(token_uuid)
AND user_tokens.user_uuid = sqlc.arg(user_uuid);
