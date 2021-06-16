// Code generated by sqlc. DO NOT EDIT.
// source: groups.sql

package postgres

import (
	"context"

	"github.com/google/uuid"
)

const createGroup = `-- name: CreateGroup :one
INSERT INTO groups(name)
VALUES($1)
RETURNING uuid, name
`

func (q *Queries) CreateGroup(ctx context.Context, name string) (Group, error) {
	row := q.queryRow(ctx, q.createGroupStmt, createGroup, name)
	var i Group
	err := row.Scan(&i.Uuid, &i.Name)
	return i, err
}

const deleteGroup = `-- name: DeleteGroup :execrows
DELETE FROM groups
WHERE groups.uuid = $1
`

func (q *Queries) DeleteGroup(ctx context.Context, uuid uuid.UUID) (int64, error) {
	result, err := q.exec(ctx, q.deleteGroupStmt, deleteGroup, uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const existsGroup = `-- name: ExistsGroup :one
SELECT COUNT(*) AS count
FROM groups
WHERE groups.uuid = $1
`

func (q *Queries) ExistsGroup(ctx context.Context, uuid uuid.UUID) (int64, error) {
	row := q.queryRow(ctx, q.existsGroupStmt, existsGroup, uuid)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const findGroupByUuid = `-- name: FindGroupByUuid :one
SELECT uuid, name FROM groups
WHERE uuid = $1
LIMIT 1
`

func (q *Queries) FindGroupByUuid(ctx context.Context, uuid uuid.UUID) (Group, error) {
	row := q.queryRow(ctx, q.findGroupByUuidStmt, findGroupByUuid, uuid)
	var i Group
	err := row.Scan(&i.Uuid, &i.Name)
	return i, err
}

const findGroups = `-- name: FindGroups :many
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
SELECT uuid, name
FROM groups
WHERE 'groups/'||groups.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT uuid, name
FROM groups
WHERE 'groups/'||groups.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT $2::BIGINT
OFFSET $1::BIGINT
`

type FindGroupsParams struct {
	ArgOffset int64
	ArgLimit  int64
	Token     []byte
}

func (q *Queries) FindGroups(ctx context.Context, arg FindGroupsParams) ([]Group, error) {
	rows, err := q.query(ctx, q.findGroupsStmt, findGroups, arg.ArgOffset, arg.ArgLimit, arg.Token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Group{}
	for rows.Next() {
		var i Group
		if err := rows.Scan(&i.Uuid, &i.Name); err != nil {
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

const findGroupsByUser = `-- name: FindGroupsByUser :many
SELECT groups.uuid, groups.name
FROM groups, user_groups
WHERE groups.uuid = user_groups.group_uuid
AND user_groups.user_uuid = $1
`

func (q *Queries) FindGroupsByUser(ctx context.Context, uuid uuid.UUID) ([]Group, error) {
	rows, err := q.query(ctx, q.findGroupsByUserStmt, findGroupsByUser, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Group{}
	for rows.Next() {
		var i Group
		if err := rows.Scan(&i.Uuid, &i.Name); err != nil {
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

const setGroupNameByUUID = `-- name: SetGroupNameByUUID :execrows
UPDATE groups
SET name = $1
WHERE groups.uuid = $2
`

type SetGroupNameByUUIDParams struct {
	Name string
	Uuid uuid.UUID
}

func (q *Queries) SetGroupNameByUUID(ctx context.Context, arg SetGroupNameByUUIDParams) (int64, error) {
	result, err := q.exec(ctx, q.setGroupNameByUUIDStmt, setGroupNameByUUID, arg.Name, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
