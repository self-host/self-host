# Performance Test

Specification of hardware (a mid-range 1U server, released back in 2014) used during testing;

```txt
product: PowerEdge R220
vendor: Dell Inc.
*-core
  *-memory
      description: System Memory
      slot: System board or motherboard
      size: 16GiB
      capabilities: ecc
  *-cpu
      product: Intel(R) Xeon(R) CPU E3-1220 v3 @ 3.10GHz
      capacity: 3700MHz
      clock: 100MHz
  *-raid
      description: RAID bus controller
      product: MegaRAID SAS 2008 [Falcon]
    *-disk
        product: PERC H310
        vendor: DELL
        size: 931GiB (999GB)
        subsystem: TOSHIBA mechanical drive
```


# The data

We generate `1000` time series for testing, each with one year (`2019-2020`) of data points every `10 minute`. This period gives us `52 561` data points per time series, a total of `52 561 000` data points.

Generating this data takes about an hour (60 minutes) on the target hardware. This time is due mainly to CPU limitations as the CPU pins to `100%` for the duration of the generation. The `52 561` random data points for each time series takes about 3-4 seconds to generate and insert.

Checking `iotop` reveals that the bottleneck was not related to the mechanical drives as the io-load remained low. It is all CPU related and is likely due to the store procedure `tsdata_insert` or index creation, or both.

We used the following trivial code to generate the data.

```python
#!/usr/bin/python3

import psycopg2

insert_timeseries = """
-- Generate Time Series
WITH new_timeseries AS (
  SELECT
    uuid_generate_v4() AS uuid
  FROM generate_series(1, 1000)
), create_timeseries AS (
  INSERT INTO timeseries(uuid, name, si_unit, lower_bound, upper_bound)
  SELECT uuid,
    'outdoortemp ' || new_timeseries.uuid,
    'C',
    -50,
    50
    FROM new_timeseries
  RETURNING *
) SELECT COUNT(*) FROM create_timeseries;
"""

get_timeseries = """
SELECT uuid FROM timeseries;
"""

generate_tsdata = """
WITH year_range AS (
    SELECT generate_series(
        '2019-01-01'::timestamptz,
        '2020-01-01'::timestamptz, 
        '10 minutes'::interval) AS when
)
SELECT
    COUNT(*) AS count
FROM year_range, tsdata_insert(
    '{0}',
    floor(random() * 100 - 50),
        year_range.when,
        '00000000-0000-1000-8000-000000000000'
) AS tsdata_insert;
"""

conn = psycopg2.connect("dbname='selfhost-test' user='postgres' host='pg13.selfhost' password='mysecretpassword'")

cur = conn.cursor()
cur.execute(insert_timeseries)
conn.commit()

cur.execute(get_timeseries)

count = 0
for x in [x[0] for x in cur.fetchall()]:
    print(count, x)
    cur.execute(generate_tsdata.format(x))
    conn.commit()
    count += 1
```

# Querying data

To query data, we use `curl` wrapped in several different shell scripts and the tool `hyperfine` to help us determine the average execution time.

## Get data

Retrieving one day of data (144 data points) from a single time series;

> hyperfine --warmup 10 "./get_data_1day.sh"

```
Benchmark #1: ./get_data_1day.sh
  Time (mean ± σ):      18.2 ms ±   6.2 ms    [User: 8.7 ms, System: 4.6 ms]
  Range (min … max):     9.6 ms …  50.5 ms    133 runs
```

Retrieving one week of data (~1 000 data points) from a single time series;

> hyperfine --warmup 10 "./get_data_1week.sh"

```
Benchmark #1: ./get_data_1week.sh
  Time (mean ± σ):      25.0 ms ±   7.3 ms    [User: 11.4 ms, System: 5.0 ms]
  Range (min … max):    16.7 ms …  50.6 ms    113 runs
```

Retrieving one month of data (~4 400 data points) from a single time series;

> hyperfine --warmup 10 "./get_data_1month.sh"

```
Benchmark #1: ./get_data_1month.sh
  Time (mean ± σ):      38.0 ms ±   9.5 ms    [User: 12.3 ms, System: 7.4 ms]
  Range (min … max):    23.5 ms …  61.4 ms    80 runs
```

Retrieving one year of data (52 561 data points) from one time series;

> hyperfine --warmup 10 "./get_data_1year.sh"

```
Benchmark #1: ./get_data_1year.sh
  Time (mean ± σ):     171.9 ms ±   5.8 ms    [User: 22.4 ms, System: 22.3 ms]
  Range (min … max):   162.5 ms … 184.4 ms    17 runs
```


## Store data

Insertion of a single datapoint;

> hyperfine --warmup 10 "./store_1datapoint.sh" 

```
Benchmark #1: ./store_1datapoint.sh
  Time (mean ± σ):      21.8 ms ±   7.5 ms    [User: 15.1 ms, System: 6.4 ms]
  Range (min … max):    11.7 ms …  57.4 ms    132 runs
```

Insertion of 10 data points at a time;

> hyperfine --warmup 10 "./store_10datapoints.sh" 

```
Benchmark #1: ./store_10datapoints.sh
  Time (mean ± σ):      25.3 ms ±   7.7 ms    [User: 16.2 ms, System: 6.9 ms]
  Range (min … max):    13.3 ms …  59.7 ms    177 runs
```

Insertion of 100 data points at a time;

> hyperfine --warmup 10 "./store_100datapoints.sh" 

```
Benchmark #1: ./store_100datapoints.sh
  Time (mean ± σ):      45.8 ms ±  10.8 ms    [User: 22.3 ms, System: 11.5 ms]
  Range (min … max):    31.4 ms …  75.5 ms    64 runs
```


## Tools

You can use the following query to track down locations where it can be good to have an index. Do remember that indexes carry with them performance penalties. They are no silver bullet.

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
