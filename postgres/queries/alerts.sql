-- name: ExistsAlert :one
SELECT COUNT(*) AS count
FROM alerts
WHERE uuid = sqlc.arg(uuid);

-- name: CreateAlert :one
SELECT  alert_merge(
        sqlc.arg(resource)::text,
        sqlc.arg(environment)::text,
        sqlc.arg(event)::text,
        sqlc.arg(origin)::text,
        sqlc.arg(severity)::alert_severity,
        sqlc.arg(status)::alert_status,
	sqlc.arg(service)::text[],
        sqlc.arg(value)::text,
        sqlc.arg(description)::text,
	sqlc.arg(tags)::text[],
	sqlc.arg(timeout)::integer,
	sqlc.arg(rawdata)::bytea
)::UUID AS uuid LIMIT 1;

-- name: FindAlertByUUID :one
SELECT *
FROM v_alerts
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetResource :execrows
UPDATE alerts
SET resource = sqlc.arg(resource)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetEnvironment :execrows
UPDATE alerts
SET environment = sqlc.arg(environment)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetEvent :execrows
UPDATE alerts
SET event = sqlc.arg(event)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetOrigin :execrows
UPDATE alerts
SET origin = sqlc.arg(origin)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetSeverity :execrows
UPDATE alerts
SET previous_severity = (
	SELECT severity
	FROM alerts
	WHERE alerts.uuid = sqlc.arg(uuid)
), severity = sqlc.arg(severity)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetStatus :execrows
UPDATE alerts
SET status = sqlc.arg(status)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetService :execrows
UPDATE alerts
SET service = sqlc.arg(service)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetValue :execrows
UPDATE alerts
SET value = sqlc.arg(value)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetDescription :execrows
UPDATE alerts
SET description = sqlc.arg(description)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetTags :execrows
UPDATE alerts
SET tags = sqlc.arg(tags)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetTimeout :execrows
UPDATE alerts
SET timeout = sqlc.arg(timeout)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetRawdata :execrows
UPDATE alerts
SET rawdata = sqlc.arg(rawdata)
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertIncDuplicate :execrows
UPDATE alerts
SET duplicate = duplicate + 1
WHERE uuid = sqlc.arg(uuid);

-- name: UpdateAlertSetLastReceivedTime :execrows
UPDATE alerts
SET last_receive_time = NOW()
WHERE uuid = sqlc.arg(uuid);

-- name: FindAlerts :many
WITH severity_levels AS (
	SELECT name::alert_severity, level::int FROM (
		VALUES
			('security', 0),
			('critical', 1),
			('major', 2),
			('minor', 3),
			('warning', 4),
			('informational', 5),
			('debug', 6),
			('trace', 7),
			('indeterminate', 8)
	) AS x(name, level)
)
SELECT
	v_alerts.created,
	v_alerts.description,
	v_alerts.duplicate,
	v_alerts.environment,
	v_alerts.event,
	v_alerts.last_receive_time,
	v_alerts.origin,
	v_alerts.previous_severity,
	v_alerts.rawdata,
	v_alerts.resource,
	v_alerts.service,
	v_alerts.severity,
	v_alerts.status,
	v_alerts.tags,
	v_alerts.timeout,
	v_alerts.uuid,
	v_alerts.value
FROM v_alerts, severity_levels
WHERE v_alerts.severity = severity_levels.name
AND (
	NULLIF(sqlc.arg(resource)::TEXT, '') IS NULL
	OR
	sqlc.arg(resource)::TEXT = v_alerts.resource
)
AND (
	NULLIF(sqlc.arg(environment)::TEXT, '') IS NULL
	OR
	sqlc.arg(environment)::TEXT = v_alerts.environment
)
AND (
	NULLIF(sqlc.arg(event)::TEXT, '') IS NULL
	OR
	sqlc.arg(event)::TEXT = v_alerts.event
)
AND (
	NULLIF(sqlc.arg(origin)::TEXT, '') IS NULL
	OR
	sqlc.arg(origin)::TEXT = v_alerts.origin
)
AND (
	NULLIF(sqlc.arg(status)::TEXT, '') IS NULL
	OR
	sqlc.arg(status)::TEXT = v_alerts.status::TEXT
)
AND (
	NULLIF(sqlc.arg(severity_le)::TEXT, '') IS NULL
	OR
	(
		SELECT level
		FROM severity_levels
		WHERE name::TEXT = sqlc.arg(severity_le)::TEXT
	) <= severity_levels.level
)
AND (
	NULLIF(sqlc.arg(severity_ge)::TEXT, '') IS NULL
	OR
	(
		SELECT level
		FROM severity_levels
		WHERE name::TEXT = sqlc.arg(severity_ge)::TEXT
	) >= severity_levels.level
)
AND (
	NULLIF(sqlc.arg(severity_eq)::TEXT, '') IS NULL
	OR
	(
		SELECT level
		FROM severity_levels
		WHERE name::TEXT = sqlc.arg(severity_eq)::TEXT
	) = severity_levels.level
)
AND (
	NULLIF(sqlc.arg(service)::TEXT[], ARRAY[]::TEXT[]) IS NULL
	OR
	sqlc.arg(service)::TEXT[] && v_alerts.service
)
AND (
	NULLIF(sqlc.arg(tags)::TEXT[], ARRAY[]::TEXT[]) IS NULL
	OR
	sqlc.arg(tags)::TEXT[] && v_alerts.tags
)
ORDER BY resource, environment, event
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: DeleteAlert :execrows
DELETE FROM alerts
WHERE alerts.uuid = sqlc.arg(uuid);
