BEGIN;

DROP VIEW v_alerts;
DROP TABLE alerts;
DROP TYPE alert_status;
DROP TYPE alert_severity;
DROP FUNCTION alert_merge;

COMMIT;
