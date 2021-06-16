// Code generated by sqlc. DO NOT EDIT.
// source: things.sql

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const createThing = `-- name: CreateThing :one
WITH t AS (
	INSERT INTO things (
		name, type, created_by, tags
	) VALUES (
		$1,
		$2,
		$3,
		$4
	)
	RETURNING uuid, name, type, state, created_by, tags
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
SELECT uuid, name, type, state, created_by, tags
FROM t LIMIT 1
`

type CreateThingParams struct {
	Name      string
	Type      sql.NullString
	CreatedBy uuid.UUID
	Tags      []string
}

type CreateThingRow struct {
	Uuid      uuid.UUID
	Name      string
	Type      sql.NullString
	State     ThingState
	CreatedBy uuid.UUID
	Tags      []string
}

func (q *Queries) CreateThing(ctx context.Context, arg CreateThingParams) (CreateThingRow, error) {
	row := q.queryRow(ctx, q.createThingStmt, createThing,
		arg.Name,
		arg.Type,
		arg.CreatedBy,
		pq.Array(arg.Tags),
	)
	var i CreateThingRow
	err := row.Scan(
		&i.Uuid,
		&i.Name,
		&i.Type,
		&i.State,
		&i.CreatedBy,
		pq.Array(&i.Tags),
	)
	return i, err
}

const deleteThing = `-- name: DeleteThing :execrows
DELETE FROM things
WHERE things.uuid = $1
`

func (q *Queries) DeleteThing(ctx context.Context, uuid uuid.UUID) (int64, error) {
	result, err := q.exec(ctx, q.deleteThingStmt, deleteThing, uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const existsThing = `-- name: ExistsThing :one
SELECT COUNT(*) AS count
FROM things
WHERE things.uuid = $1
`

func (q *Queries) ExistsThing(ctx context.Context, uuid uuid.UUID) (int64, error) {
	row := q.queryRow(ctx, q.existsThingStmt, existsThing, uuid)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const findThingByUUID = `-- name: FindThingByUUID :one
SELECT uuid, name, type, state, created_by, tags
FROM things
WHERE things.uuid = $1
LIMIT 1
`

func (q *Queries) FindThingByUUID(ctx context.Context, uuid uuid.UUID) (Thing, error) {
	row := q.queryRow(ctx, q.findThingByUUIDStmt, findThingByUUID, uuid)
	var i Thing
	err := row.Scan(
		&i.Uuid,
		&i.Name,
		&i.Type,
		&i.State,
		&i.CreatedBy,
		pq.Array(&i.Tags),
	)
	return i, err
}

const findThings = `-- name: FindThings :many
WITH usr AS (
	SELECT users.uuid
	FROM users, user_tokens
	WHERE user_tokens.user_uuid = users.uuid
	AND user_tokens.token_hash = sha256($3)
	LIMIT 1
), policies AS (
	SELECT group_policies.effect, group_policies.priority, group_policies.resource
	FROM group_policies, user_groups
	WHERE user_groups.group_uuid = group_policies.group_uuid
	AND user_groups.user_uuid = (SELECT uuid FROM usr)
	AND action = 'read'
)
SELECT uuid, name, type, state, created_by, tags
FROM things
WHERE 'things/'||things.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT uuid, name, type, state, created_by, tags
FROM things
WHERE 'things/'||things.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT $2::BIGINT
OFFSET $1::BIGINT
`

type FindThingsParams struct {
	ArgOffset int64
	ArgLimit  int64
	Token     []byte
}

func (q *Queries) FindThings(ctx context.Context, arg FindThingsParams) ([]Thing, error) {
	rows, err := q.query(ctx, q.findThingsStmt, findThings, arg.ArgOffset, arg.ArgLimit, arg.Token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Thing{}
	for rows.Next() {
		var i Thing
		if err := rows.Scan(
			&i.Uuid,
			&i.Name,
			&i.Type,
			&i.State,
			&i.CreatedBy,
			pq.Array(&i.Tags),
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findThingsByTags = `-- name: FindThingsByTags :many
WITH usr AS (
	SELECT users.uuid
	FROM users, user_tokens
	WHERE user_tokens.user_uuid = users.uuid
	AND user_tokens.token_hash = sha256($3)
	LIMIT 1
), policies AS (
	SELECT group_policies.effect, group_policies.priority, group_policies.resource
	FROM group_policies, user_groups
	WHERE user_groups.group_uuid = group_policies.group_uuid
	AND user_groups.user_uuid = (SELECT uuid FROM usr)
	AND action = 'read'
)
SELECT uuid, name, type, state, created_by, tags
FROM things
WHERE 'things/'||things.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
AND $4 && things.tags
EXCEPT
SELECT uuid, name, type, state, created_by, tags
FROM things
WHERE 'things/'||things.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
AND $4 && things.tags
ORDER BY name
LIMIT $2::BIGINT
OFFSET $1::BIGINT
`

type FindThingsByTagsParams struct {
	ArgOffset int64
	ArgLimit  int64
	Token     []byte
	Tags      interface{}
}

func (q *Queries) FindThingsByTags(ctx context.Context, arg FindThingsByTagsParams) ([]Thing, error) {
	rows, err := q.query(ctx, q.findThingsByTagsStmt, findThingsByTags,
		arg.ArgOffset,
		arg.ArgLimit,
		arg.Token,
		arg.Tags,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Thing{}
	for rows.Next() {
		var i Thing
		if err := rows.Scan(
			&i.Uuid,
			&i.Name,
			&i.Type,
			&i.State,
			&i.CreatedBy,
			pq.Array(&i.Tags),
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setThingNameByUUID = `-- name: SetThingNameByUUID :execrows
UPDATE things
SET name = $1
WHERE things.uuid = $2
`

type SetThingNameByUUIDParams struct {
	Name string
	Uuid uuid.UUID
}

func (q *Queries) SetThingNameByUUID(ctx context.Context, arg SetThingNameByUUIDParams) (int64, error) {
	result, err := q.exec(ctx, q.setThingNameByUUIDStmt, setThingNameByUUID, arg.Name, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setThingStateByUUID = `-- name: SetThingStateByUUID :execrows
UPDATE things
SET state = $1
WHERE things.uuid = $2
`

type SetThingStateByUUIDParams struct {
	State ThingState
	Uuid  uuid.UUID
}

func (q *Queries) SetThingStateByUUID(ctx context.Context, arg SetThingStateByUUIDParams) (int64, error) {
	result, err := q.exec(ctx, q.setThingStateByUUIDStmt, setThingStateByUUID, arg.State, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setThingTags = `-- name: SetThingTags :execrows
UPDATE things
SET tags = $1
WHERE things.uuid = $2
`

type SetThingTagsParams struct {
	Tags []string
	Uuid uuid.UUID
}

func (q *Queries) SetThingTags(ctx context.Context, arg SetThingTagsParams) (int64, error) {
	result, err := q.exec(ctx, q.setThingTagsStmt, setThingTags, pq.Array(arg.Tags), arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setThingTypeByUUID = `-- name: SetThingTypeByUUID :execrows
UPDATE things
SET type = $1
WHERE things.uuid = $2
`

type SetThingTypeByUUIDParams struct {
	Type sql.NullString
	Uuid uuid.UUID
}

func (q *Queries) SetThingTypeByUUID(ctx context.Context, arg SetThingTypeByUUIDParams) (int64, error) {
	result, err := q.exec(ctx, q.setThingTypeByUUIDStmt, setThingTypeByUUID, arg.Type, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
