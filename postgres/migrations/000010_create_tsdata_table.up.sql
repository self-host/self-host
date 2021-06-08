BEGIN;

CREATE TABLE tsdata (
  ts_uuid UUID REFERENCES timeseries(uuid) NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  ts TIMESTAMPTZ NOT NULL,
  created_by UUID REFERENCES users(uuid) ON DELETE SET NULL,

  UNIQUE(ts_uuid, ts)
) PARTITION BY RANGE (ts);

CREATE INDEX tsdata_created_by_idx ON tsdata(created_by);

--
-- Functions to insert data into tsdata as a partition
--

CREATE FUNCTION public.tsdata_insert(
	ts_uuid UUID,
	value DOUBLE PRECISION,
	ts TIMESTAMPTZ,
	created_by UUID)
    RETURNS SETOF tsdata
    LANGUAGE 'plpgsql'
    RETURNS NULL ON NULL INPUT
AS $BODY$
BEGIN
	RETURN QUERY
	INSERT INTO tsdata(ts_uuid, value, ts, created_by)
	VALUES(ts_uuid, value, ts, created_by)
	RETURNING *;

EXCEPTION
	WHEN check_violation THEN
	EXECUTE
	'CREATE TABLE IF NOT EXISTS' ||
	' tsdata_y' || to_char(ts, 'YYYYmMM') ||
	' PARTITION OF tsdata FOR VALUES FROM (''' ||
	date_trunc('month', ts) ||
	''') TO (''' ||
	(date_trunc('month', ts) + '1month'::interval) ||
	''')';

	RETURN QUERY
	INSERT INTO tsdata(ts_uuid, value, ts, created_by)
	VALUES(ts_uuid, value, ts, created_by)
	RETURNING *;
END;
$BODY$;

COMMIT;
