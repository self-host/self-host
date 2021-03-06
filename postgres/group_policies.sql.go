// Code generated by sqlc. DO NOT EDIT.
// source: group_policies.sql

package postgres

import (
	"context"

	"github.com/google/uuid"
)

const findPoliciesByGroup = `-- name: FindPoliciesByGroup :many
SELECT uuid, group_uuid, priority, effect, action, resource FROM group_policies
WHERE group_policies.group_uuid = $1
ORDER BY priority
`

func (q *Queries) FindPoliciesByGroup(ctx context.Context, uuid uuid.UUID) ([]GroupPolicy, error) {
	rows, err := q.query(ctx, q.findPoliciesByGroupStmt, findPoliciesByGroup, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GroupPolicy{}
	for rows.Next() {
		var i GroupPolicy
		if err := rows.Scan(
			&i.Uuid,
			&i.GroupUuid,
			&i.Priority,
			&i.Effect,
			&i.Action,
			&i.Resource,
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

const findPoliciesByUser = `-- name: FindPoliciesByUser :many
SELECT group_policies.uuid, group_policies.group_uuid, group_policies.priority, group_policies.effect, group_policies.action, group_policies.resource
FROM group_policies, user_groups
WHERE user_groups.group_uuid = group_policies.group_uuid
AND user_groups.user_uuid = $1
ORDER BY priority
`

func (q *Queries) FindPoliciesByUser(ctx context.Context, uuid uuid.UUID) ([]GroupPolicy, error) {
	rows, err := q.query(ctx, q.findPoliciesByUserStmt, findPoliciesByUser, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GroupPolicy{}
	for rows.Next() {
		var i GroupPolicy
		if err := rows.Scan(
			&i.Uuid,
			&i.GroupUuid,
			&i.Priority,
			&i.Effect,
			&i.Action,
			&i.Resource,
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
