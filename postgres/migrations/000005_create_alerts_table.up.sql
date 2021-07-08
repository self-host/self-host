BEGIN;

CREATE TYPE alert_status AS ENUM (
        'open', 'close', 'expire', 'shelve',
        'acknowledge', 'unknown'
);

CREATE TYPE alert_severity AS ENUM (
        'security', 'critical', 'major',
        'minor', 'warning', 'informational',
        'debug', 'trace', 'indeterminate'
);

CREATE TABLE alerts (
        uuid UUID NOT NULL DEFAULT uuid_generate_v4 () PRIMARY KEY,
        resource TEXT NOT NULL,
        environment TEXT NOT NULL,
        event TEXT NOT NULL,
        severity alert_severity NOT NULL,
        status alert_status NOT NULL,
        service TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
        value TEXT NOT NULL DEFAULT '',
        description TEXT NOT NULL DEFAULT '',
        origin TEXT NOT NULL DEFAULT '',
        tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
        created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        timeout INTEGER NOT NULL DEFAULT 3600,
        rawdata BYTEA NOT NULL DEFAULT ''::BYTEA,
        duplicate INTEGER NOT NULL DEFAULT 0,
        previous_severity alert_severity NOT NULL DEFAULT 'indeterminate'::alert_severity,
        last_receive_time TIMESTAMPTZ
);

CREATE INDEX programs_environment_idx ON alerts(environment);
CREATE INDEX programs_severity_idx ON alerts(severity);
CREATE INDEX programs_status_idx ON alerts(status);

CREATE VIEW v_alerts
AS
 SELECT
    alerts.uuid,
    alerts.resource,
    alerts.environment,
    alerts.event,
    alerts.severity,
    alerts.previous_severity,
    (CASE
        WHEN alerts.status IN ('open'::alert_status, 'acknowledge'::alert_status)
            AND (
                COALESCE(alerts.last_receive_time, alerts.created)
                + make_interval(secs => alerts.timeout)
            ) < now() THEN
                'expire'::alert_status
        ELSE alerts.status
    END)::alert_status AS status,
    alerts.description,
    alerts.value,
    alerts.origin,
    alerts.created,
    alerts.last_receive_time,
    alerts.timeout,
    alerts.duplicate,
    alerts.service,
    alerts.tags,
    alerts.rawdata
 FROM alerts;

CREATE OR REPLACE FUNCTION public.alert_merge(
        resource text,
        environment text,
        event text,
        origin text,
        severity alert_severity,
        status alert_status,
        service text[],
        value text,
        description text,
        tags text[],
        timeout integer,
        rawdata bytea)
    RETURNS alerts.uuid%TYPE
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE PARALLEL UNSAFE
AS $BODY$
DECLARE
	res_id alerts.uuid%TYPE;
BEGIN
    -- first try to update the key
    UPDATE alerts SET
        duplicate = alerts.duplicate + 1,
        last_receive_time = now(),
        previous_severity = alerts.severity,
        severity = alert_merge.severity,
        service = alert_merge.service,
        description = alert_merge.description,
        value = alert_merge.value,
        timeout = alert_merge.timeout,
        tags = alert_merge.tags,
        rawdata = alert_merge.rawdata
    WHERE alerts.status IN (
        'open'::alert_status,
        'acknowledge'::alert_status,
        'shelve'::alert_status
    )
    AND COALESCE(alerts.last_receive_time, alerts.created) + make_interval(secs => alerts.timeout) >= NOW()
    AND alerts.resource = alert_merge.resource
    AND alerts.environment = alert_merge.environment
    AND alerts.event = alert_merge.event
    AND alerts.origin = alert_merge.origin
	RETURNING uuid INTO res_id;
    IF found THEN
        RETURN res_id;
    END IF;

    INSERT INTO alerts (resource, environment, event, severity, status,
                        value, description, origin, service, tags, timeout, rawdata)
    VALUES (
        resource,
        environment,
        event,
        COALESCE(severity, 'indeterminate'::alert_severity),
        COALESCE(status, 'open'::alert_status),
        value,
        description,
        origin,
        service,
        tags,
        timeout,
        rawdata
    ) RETURNING uuid INTO res_id;
	RETURN res_id;
END;
$BODY$;

COMMIT;
