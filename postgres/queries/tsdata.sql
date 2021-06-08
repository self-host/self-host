-- name: GetTsDataRange :many
SELECT	ts_uuid,
	value,
	ts
FROM tsdata
WHERE ts_uuid = ANY(sqlc.arg(ts_uuids)::uuid[])
AND ts BETWEEN sqlc.arg(start) AND sqlc.arg(stop)
ORDER BY ts ASC;

-- name: CreateTsData :one
SELECT
	COUNT(*) AS count
FROM tsdata_insert(
  sqlc.arg(ts_uuid),
  sqlc.arg(value),
  sqlc.arg(ts),
  sqlc.arg(created_by)
) AS tsdata_insert;

-- name: DeleteAllTsData :execrows
DELETE FROM tsdata
WHERE ts_uuid = ANY(sqlc.arg(ts_uuid));

-- name: DeleteTsDataRange :execrows
DELETE FROM tsdata
WHERE ts_uuid = ANY(sqlc.arg(ts_uuids)::uuid[])
AND ts BETWEEN sqlc.arg(start) AND sqlc.arg(stop)
AND (sqlc.arg(ge_null)::boolean = true OR tsdata.value >= sqlc.arg(ge))
AND (sqlc.arg(le_null)::boolean = true OR tsdata.value <= sqlc.arg(le))
;
