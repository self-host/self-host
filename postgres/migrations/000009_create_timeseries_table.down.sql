BEGIN;

DROP INDEX IF EXISTS timeseries_tags_idx;

DROP TABLE timeseries;

COMMIT;
