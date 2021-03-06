// Code generated by sqlc. DO NOT EDIT.
// source: timeseries.sql

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const createTimeseries = `-- name: CreateTimeseries :one
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
		NULLIF($1::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
		$2,
		$3,
		$4,
		$5,
		$6,
		$7
	) RETURNING uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags
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
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags
FROM t LIMIT 1
`

type CreateTimeseriesParams struct {
	ThingUuid  uuid.UUID
	Name       string
	SiUnit     string
	LowerBound sql.NullFloat64
	UpperBound sql.NullFloat64
	CreatedBy  uuid.UUID
	Tags       []string
}

type CreateTimeseriesRow struct {
	Uuid       uuid.UUID
	ThingUuid  uuid.UUID
	Name       string
	SiUnit     string
	LowerBound sql.NullFloat64
	UpperBound sql.NullFloat64
	CreatedBy  uuid.UUID
	Tags       []string
}

func (q *Queries) CreateTimeseries(ctx context.Context, arg CreateTimeseriesParams) (CreateTimeseriesRow, error) {
	row := q.queryRow(ctx, q.createTimeseriesStmt, createTimeseries,
		arg.ThingUuid,
		arg.Name,
		arg.SiUnit,
		arg.LowerBound,
		arg.UpperBound,
		arg.CreatedBy,
		pq.Array(arg.Tags),
	)
	var i CreateTimeseriesRow
	err := row.Scan(
		&i.Uuid,
		&i.ThingUuid,
		&i.Name,
		&i.SiUnit,
		&i.LowerBound,
		&i.UpperBound,
		&i.CreatedBy,
		pq.Array(&i.Tags),
	)
	return i, err
}

const deleteTimeseries = `-- name: DeleteTimeseries :execrows
DELETE FROM timeseries
WHERE uuid = $1
`

func (q *Queries) DeleteTimeseries(ctx context.Context, uuid uuid.UUID) (int64, error) {
	result, err := q.exec(ctx, q.deleteTimeseriesStmt, deleteTimeseries, uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const existsTimeseries = `-- name: ExistsTimeseries :one
SELECT COUNT(*) AS count
FROM timeseries
WHERE timeseries.uuid = $1
`

func (q *Queries) ExistsTimeseries(ctx context.Context, uuid uuid.UUID) (int64, error) {
	row := q.queryRow(ctx, q.existsTimeseriesStmt, existsTimeseries, uuid)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const findTimeseries = `-- name: FindTimeseries :many
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
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT $2::BIGINT
OFFSET $1::BIGINT
`

type FindTimeseriesParams struct {
	ArgOffset int64
	ArgLimit  int64
	Token     []byte
}

func (q *Queries) FindTimeseries(ctx context.Context, arg FindTimeseriesParams) ([]Timeseries, error) {
	rows, err := q.query(ctx, q.findTimeseriesStmt, findTimeseries, arg.ArgOffset, arg.ArgLimit, arg.Token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Timeseries{}
	for rows.Next() {
		var i Timeseries
		if err := rows.Scan(
			&i.Uuid,
			&i.ThingUuid,
			&i.Name,
			&i.SiUnit,
			&i.LowerBound,
			&i.UpperBound,
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

const findTimeseriesByTags = `-- name: FindTimeseriesByTags :many
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
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
AND $4 && timeseries.tags
EXCEPT
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags
FROM timeseries
WHERE 'timeseries/'||timeseries.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
AND $4 && timeseries.tags
ORDER BY name
LIMIT $2::BIGINT
OFFSET $1::BIGINT
`

type FindTimeseriesByTagsParams struct {
	ArgOffset int64
	ArgLimit  int64
	Token     []byte
	Tags      interface{}
}

func (q *Queries) FindTimeseriesByTags(ctx context.Context, arg FindTimeseriesByTagsParams) ([]Timeseries, error) {
	rows, err := q.query(ctx, q.findTimeseriesByTagsStmt, findTimeseriesByTags,
		arg.ArgOffset,
		arg.ArgLimit,
		arg.Token,
		arg.Tags,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Timeseries{}
	for rows.Next() {
		var i Timeseries
		if err := rows.Scan(
			&i.Uuid,
			&i.ThingUuid,
			&i.Name,
			&i.SiUnit,
			&i.LowerBound,
			&i.UpperBound,
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

const findTimeseriesByThing = `-- name: FindTimeseriesByThing :many
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags FROM timeseries
WHERE $1 = timeseries.thing_uuid
ORDER BY name
`

func (q *Queries) FindTimeseriesByThing(ctx context.Context, thingUuid interface{}) ([]Timeseries, error) {
	rows, err := q.query(ctx, q.findTimeseriesByThingStmt, findTimeseriesByThing, thingUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Timeseries{}
	for rows.Next() {
		var i Timeseries
		if err := rows.Scan(
			&i.Uuid,
			&i.ThingUuid,
			&i.Name,
			&i.SiUnit,
			&i.LowerBound,
			&i.UpperBound,
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

const findTimeseriesByUUID = `-- name: FindTimeseriesByUUID :one
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags FROM timeseries
WHERE $1 = timeseries.uuid
LIMIT 1
`

func (q *Queries) FindTimeseriesByUUID(ctx context.Context, tsUuid interface{}) (Timeseries, error) {
	row := q.queryRow(ctx, q.findTimeseriesByUUIDStmt, findTimeseriesByUUID, tsUuid)
	var i Timeseries
	err := row.Scan(
		&i.Uuid,
		&i.ThingUuid,
		&i.Name,
		&i.SiUnit,
		&i.LowerBound,
		&i.UpperBound,
		&i.CreatedBy,
		pq.Array(&i.Tags),
	)
	return i, err
}

const getTimeseriesByUUID = `-- name: GetTimeseriesByUUID :one
SELECT uuid, thing_uuid, name, si_unit, lower_bound, upper_bound, created_by, tags FROM timeseries
WHERE uuid = $1
LIMIT 1
`

func (q *Queries) GetTimeseriesByUUID(ctx context.Context, uuid uuid.UUID) (Timeseries, error) {
	row := q.queryRow(ctx, q.getTimeseriesByUUIDStmt, getTimeseriesByUUID, uuid)
	var i Timeseries
	err := row.Scan(
		&i.Uuid,
		&i.ThingUuid,
		&i.Name,
		&i.SiUnit,
		&i.LowerBound,
		&i.UpperBound,
		&i.CreatedBy,
		pq.Array(&i.Tags),
	)
	return i, err
}

const getUnitFromTimeseries = `-- name: GetUnitFromTimeseries :one
SELECT si_unit FROM timeseries
WHERE uuid = $1
LIMIT 1
`

func (q *Queries) GetUnitFromTimeseries(ctx context.Context, uuid uuid.UUID) (string, error) {
	row := q.queryRow(ctx, q.getUnitFromTimeseriesStmt, getUnitFromTimeseries, uuid)
	var si_unit string
	err := row.Scan(&si_unit)
	return si_unit, err
}

const setTimeseriesLowerBound = `-- name: SetTimeseriesLowerBound :execrows
UPDATE timeseries
SET lower_bound = $1
WHERE timeseries.uuid = $2
`

type SetTimeseriesLowerBoundParams struct {
	LowerBound sql.NullFloat64
	Uuid       uuid.UUID
}

func (q *Queries) SetTimeseriesLowerBound(ctx context.Context, arg SetTimeseriesLowerBoundParams) (int64, error) {
	result, err := q.exec(ctx, q.setTimeseriesLowerBoundStmt, setTimeseriesLowerBound, arg.LowerBound, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setTimeseriesName = `-- name: SetTimeseriesName :execrows
UPDATE timeseries
SET name = $1
WHERE timeseries.uuid = $2
`

type SetTimeseriesNameParams struct {
	Name string
	Uuid uuid.UUID
}

func (q *Queries) SetTimeseriesName(ctx context.Context, arg SetTimeseriesNameParams) (int64, error) {
	result, err := q.exec(ctx, q.setTimeseriesNameStmt, setTimeseriesName, arg.Name, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setTimeseriesSiUnit = `-- name: SetTimeseriesSiUnit :execrows
UPDATE timeseries
SET si_unit = $1
WHERE timeseries.uuid = $2
`

type SetTimeseriesSiUnitParams struct {
	SiUnit string
	Uuid   uuid.UUID
}

func (q *Queries) SetTimeseriesSiUnit(ctx context.Context, arg SetTimeseriesSiUnitParams) (int64, error) {
	result, err := q.exec(ctx, q.setTimeseriesSiUnitStmt, setTimeseriesSiUnit, arg.SiUnit, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setTimeseriesTags = `-- name: SetTimeseriesTags :execrows
UPDATE timeseries
SET tags = $1
WHERE timeseries.uuid = $2
`

type SetTimeseriesTagsParams struct {
	Tags []string
	Uuid uuid.UUID
}

func (q *Queries) SetTimeseriesTags(ctx context.Context, arg SetTimeseriesTagsParams) (int64, error) {
	result, err := q.exec(ctx, q.setTimeseriesTagsStmt, setTimeseriesTags, pq.Array(arg.Tags), arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setTimeseriesThing = `-- name: SetTimeseriesThing :execrows
UPDATE timeseries
SET thing_uuid = $1
WHERE timeseries.uuid = $2
`

type SetTimeseriesThingParams struct {
	ThingUuid uuid.UUID
	Uuid      uuid.UUID
}

func (q *Queries) SetTimeseriesThing(ctx context.Context, arg SetTimeseriesThingParams) (int64, error) {
	result, err := q.exec(ctx, q.setTimeseriesThingStmt, setTimeseriesThing, arg.ThingUuid, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setTimeseriesUpperBound = `-- name: SetTimeseriesUpperBound :execrows
UPDATE timeseries
SET upper_bound = $1
WHERE timeseries.uuid = $2
`

type SetTimeseriesUpperBoundParams struct {
	UpperBound sql.NullFloat64
	Uuid       uuid.UUID
}

func (q *Queries) SetTimeseriesUpperBound(ctx context.Context, arg SetTimeseriesUpperBoundParams) (int64, error) {
	result, err := q.exec(ctx, q.setTimeseriesUpperBoundStmt, setTimeseriesUpperBound, arg.UpperBound, arg.Uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
