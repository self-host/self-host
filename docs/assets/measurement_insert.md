```sql
CREATE OR REPLACE FUNCTION public.measurement_insert(
	city_id integer,
	logdate timestamp with time zone,
	peaktemp integer,
	unitsales integer)
    RETURNS SETOF measurement 
    LANGUAGE 'plpgsql'
    RETURNS NULL ON NULL INPUT
AS $BODY$
BEGIN
	RETURN QUERY
	INSERT INTO measurement(city_id, logdate, peaktemp, unitsales)
	VALUES(city_id, logdate, peaktemp, unitsales)
	RETURNING *;

EXCEPTION
	WHEN check_violation THEN
	EXECUTE
	'CREATE TABLE measurement_y' || to_char(logdate, 'YYYYmMM') ||
	' PARTITION OF measurement FOR VALUES FROM (''' ||
	date_trunc('month', logdate) ||
	''') TO (''' || 
	(date_trunc('month', logdate) + '1month'::interval) || 
	''')';

	RETURN QUERY
	INSERT INTO measurement(city_id, logdate, peaktemp, unitsales)
	VALUES(city_id, logdate, peaktemp, unitsales) RETURNING *;
END;
$BODY$;
```
