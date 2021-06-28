-- name: GetTsDataRange :many
SELECT	ts_uuid,
	value,
	ts
FROM tsdata
WHERE ts_uuid = ANY(sqlc.arg(ts_uuids)::uuid[])
AND ts BETWEEN sqlc.arg(start) AND sqlc.arg(stop)
ORDER BY ts ASC;

-- name: GetTsDataRangeAgg :many
WITH tsdata_trunc AS (
	SELECT
       	ts_uuid,
       	value,
	CASE
		WHEN sqlc.arg(truncate)::text = 'minute5' THEN
		  (date_trunc('hour', ts) + date_part('minute', ts)::int / 5 * interval '5 min') AT time zone sqlc.arg(timezone)::text
		WHEN sqlc.arg(truncate)::text = 'minute10' THEN
		  (date_trunc('hour', ts) + date_part('minute', ts)::int / 10 * interval '10 min') AT time zone sqlc.arg(timezone)::text
		WHEN sqlc.arg(truncate)::text = 'minute15' THEN
		  (date_trunc('hour', ts) + date_part('minute', ts)::int / 15 * interval '15 min') AT time zone sqlc.arg(timezone)::text
		WHEN sqlc.arg(truncate)::text = 'minute20' THEN
		  (date_trunc('hour', ts) + date_part('minute', ts)::int / 20 * interval '20 min') AT time zone sqlc.arg(timezone)::text
		WHEN sqlc.arg(truncate)::text = 'minute30' THEN
		  (date_trunc('hour', ts) + date_part('minute', ts)::int / 30 * interval '30 min') AT time zone sqlc.arg(timezone)::text
		ELSE
		  date_trunc(sqlc.arg(truncate)::text, ts AT time zone sqlc.arg(timezone)::text) AT time zone sqlc.arg(timezone)::text
	END AS ts
	FROM tsdata
	WHERE ts_uuid = ANY(sqlc.arg(ts_uuids)::uuid[])
	AND ts BETWEEN (sqlc.arg(start)::timestamptz AT time zone sqlc.arg(timezone)::text)
		AND (sqlc.arg(stop)::timestamptz AT time zone sqlc.arg(timezone)::text)
)
SELECT
        ts_uuid::uuid,
	(CASE
		WHEN sqlc.arg(aggregate)::text = 'avg'::text THEN AVG(value)
		WHEN sqlc.arg(aggregate)::text = 'min'::text THEN MIN(value)
		WHEN sqlc.arg(aggregate)::text = 'max'::text THEN MAX(value)
		WHEN sqlc.arg(aggregate)::text = 'count'::text THEN COUNT(value)
		WHEN sqlc.arg(aggregate)::text = 'sum'::text THEN SUM(value)
	END)::DOUBLE PRECISION AS value,
        ts::timestamptz
FROM tsdata_trunc
GROUP BY ts_uuid, ts
ORDER BY ts ASC;

-- name: CreateTsData :execrows
INSERT INTO tsdata(ts_uuid, value, ts, created_by)
VALUES (
	sqlc.arg(ts_uuid),
	sqlc.arg(value),
	sqlc.arg(ts),
	sqlc.arg(created_by)
);

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
