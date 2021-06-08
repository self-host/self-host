# Performance Test

```sql
-- Generate Time Series
WITH new_timeseries AS (
  SELECT
    uuid_generate_v4() AS uuid
  FROM generate_series(1, 100)
), create_timeseries AS (
  INSERT INTO timeseries(uuid, thing_uuid, label, si_unit, user_uuid, group_uuid, ugo)
  SELECT uuid,
    'b71e5916-03ed-4d1a-8dd5-1bfc1ddd0487',
    'outdoortemp',
    'C',
    '6369c1c6-cc72-43ac-8061-a77b552e3f67',
    'aeddbffe-23b0-4ebf-bf41-2f25648f3b34',
    15 FROM new_timeseries
  RETURNING *
) SELECT * FROM create_timeseries;
```

```sql
# Helper to determine missing indexes

SELECT c.conrelid::regclass AS "table",
       /* list of key column names in order */
       string_agg(a.attname, ',' ORDER BY x.n) AS columns,
       pg_catalog.pg_size_pretty(
          pg_catalog.pg_relation_size(c.conrelid)
       ) AS size,
       c.conname AS constraint,
       c.confrelid::regclass AS referenced_table
FROM pg_catalog.pg_constraint c
   /* enumerated key column numbers per foreign key */
   CROSS JOIN LATERAL
      unnest(c.conkey) WITH ORDINALITY AS x(attnum, n)
   /* name for each key column */
   JOIN pg_catalog.pg_attribute a
      ON a.attnum = x.attnum
         AND a.attrelid = c.conrelid
WHERE NOT EXISTS
        /* is there a matching index for the constraint? */
        (SELECT 1 FROM pg_catalog.pg_index i
         WHERE i.indrelid = c.conrelid
           /* it must not be a partial index */
           AND i.indpred IS NULL
           /* the first index columns must be the same as the
              key columns, but order doesn't matter */
           AND (i.indkey::smallint[])[0:cardinality(c.conkey)-1]
               OPERATOR(pg_catalog.@>) c.conkey)
  AND c.contype = 'f'
GROUP BY c.conrelid, c.conname, c.confrelid
ORDER BY pg_catalog.pg_relation_size(c.conrelid) DESC;
```