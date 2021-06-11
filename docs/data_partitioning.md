# Data partitioning (Time Series)

In PostgreSQL, there are two (primary) ways to handle data partitioning. Using INHERITANCE or using PARTITIONS. The former is the older solution (from around version 7), and the latter was introduced in PostgreSQL 10.

Choosing one or the other is a choice between different drawbacks. Neither solutions are a complete automated solution as they both lack support for automatic creation of partition tables. The partition table must exist before we can insert any data into it. 

But, there are ways to achieve an automated solution anyway. Although, with some limitations.

When using INHERITANCE, it is possible to use a before trigger on the parent table. This trigger then takes care of creating the table if it is missing and inserting data to it. The problem with this is that the trigger has to return NULL; otherwise, the new row will also be inserted into the parent table, resulting in duplicate rows. When the trigger returns NULL, it will appear as if the INSERT statement returned no new rows.

When using PARTITIONS, it is impossible to assign a trigger to the parent table; we can only assign triggers to the partitions, which will not help us. Thus, the only logical solution is to use a separate function to insert the data to the correct table. This function is very similar to the before trigger one would use when using INHERITANCE.

Below is an example of such a function;

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
    INSERT INTO measurement(
        city_id, logdate, peaktemp, unitsales)
    VALUES(city_id, logdate, peaktemp, unitsales)
    RETURNING *;

EXCEPTION
    WHEN check_violation THEN
    EXECUTE
    'CREATE TABLE IF NOT EXISTS' || 
    ' measurement_y' || to_char(logdate, 'YYYYmMM') ||
    ' PARTITION OF measurement FOR VALUES FROM (''' ||
    date_trunc('month', logdate) ||
    ''') TO (''' || 
    (date_trunc('month', logdate) + '1month'::interval) || 
    ''')';

    RETURN QUERY
    INSERT INTO measurement(
        city_id, logdate, peaktemp, unitsales)
    VALUES(city_id, logdate, peaktemp, unitsales)
    RETURNING *;
END;
$BODY$;
```

What this function does is that it tries to inserts the data. If it fails because the table does not exist, it creates the table and tries to insert the data once again. This way, we should avoid any overhead by establishing that the table exists before inserting data.
