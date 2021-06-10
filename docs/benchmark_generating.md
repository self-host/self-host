# Populating the database

We generate `1000` time series for testing, each with one year (`2019-2020`) of data points every `10 minute`. This period gives us `52 561` data points per time series, a total of `52 561 000` data points.

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

# Data sets and shell scripts used for benchmarking inserts

```
get_data_period.sh 2019-01-01 2019-02-01
```

**get_data_period.sh**

```bash
#!/bin/bash
# get_data_period.sh

HOST='127.0.0.1:8080'
FROM=$1
TO=$2

# change to the target time series
TIMESERIES='017bcd5e-30cb-448c-97fe-e563fd82bb8b'

curl 'http://'$HOST'/v2/timeseries/'$TIMESERIES'/data?start='$FROM'T00%3A00%3A00Z&end='$TO'T00%3A00%3A00Z' -H 'accept: application/json' -H 'Authorization: Basic dGVzdDpyb290'
```


```
insert_data.sh data_1point.json
```

**insert_data.sh**

Please edit where appropriate before use.

```bash
#!/bin/bash
# insert_data.sh

HOST='127.0.0.1:8080'

# change to the target time series
TIMESERIES='017bcd5e-30cb-448c-97fe-e563fd82bb8b'

# Take a JSON file and replace DATETIME with the current time in RFC3339 format
# Note: change +02:00 to your time zone.
DATA=$(cat $1 | sed "s/DATETIME/`date --rfc-3339=second | date --rfc-3339=ns | sed 's/ /T/; s/\(\....\).*\([+-]\)/\1\2/g' | sed 's/+02:00//g'`/g")

# POST the result to the target Self-host server
curl -X 'POST' \
  'http://'$HOST'/v2/timeseries/'$TIMESERIES'/data?unit=C' \
  -H 'accept: */*' \
  -H 'Authorization: Basic dGVzdDpyb290' \
  -H 'Content-Type: application/json' \
  -d "$DATA"
```

**data_1point.json**

```json
[
  {"v": 3.14, "ts": "DATETIME00+02:00"}
]
```

**data_10points.json**

```json
[
  {"v": 3.14, "ts": "DATETIME00+02:00"},
  {"v": 3.14, "ts": "DATETIME01+02:00"},
  {"v": 3.14, "ts": "DATETIME02+02:00"},
  {"v": 3.14, "ts": "DATETIME03+02:00"},
  {"v": 3.14, "ts": "DATETIME04+02:00"},
  {"v": 3.14, "ts": "DATETIME05+02:00"},
  {"v": 3.14, "ts": "DATETIME06+02:00"},
  {"v": 3.14, "ts": "DATETIME07+02:00"},
  {"v": 3.14, "ts": "DATETIME08+02:00"},
  {"v": 3.14, "ts": "DATETIME09+02:00"}
]
```

**data_100points.json**

```json
[
  {"v": 3.14, "ts": "DATETIME00+02:00"},
  {"v": 3.14, "ts": "DATETIME01+02:00"},
  {"v": 3.14, "ts": "DATETIME02+02:00"},
  {"v": 3.14, "ts": "DATETIME03+02:00"},
  {"v": 3.14, "ts": "DATETIME04+02:00"},
  {"v": 3.14, "ts": "DATETIME05+02:00"},
  {"v": 3.14, "ts": "DATETIME06+02:00"},
  {"v": 3.14, "ts": "DATETIME07+02:00"},
  {"v": 3.14, "ts": "DATETIME08+02:00"},
  {"v": 3.14, "ts": "DATETIME09+02:00"},
  {"v": 3.14, "ts": "DATETIME10+02:00"},
  {"v": 3.14, "ts": "DATETIME11+02:00"},
  {"v": 3.14, "ts": "DATETIME12+02:00"},
  {"v": 3.14, "ts": "DATETIME13+02:00"},
  {"v": 3.14, "ts": "DATETIME14+02:00"},
  {"v": 3.14, "ts": "DATETIME15+02:00"},
  {"v": 3.14, "ts": "DATETIME16+02:00"},
  {"v": 3.14, "ts": "DATETIME17+02:00"},
  {"v": 3.14, "ts": "DATETIME18+02:00"},
  {"v": 3.14, "ts": "DATETIME19+02:00"},
  {"v": 3.14, "ts": "DATETIME20+02:00"},
  {"v": 3.14, "ts": "DATETIME21+02:00"},
  {"v": 3.14, "ts": "DATETIME22+02:00"},
  {"v": 3.14, "ts": "DATETIME23+02:00"},
  {"v": 3.14, "ts": "DATETIME24+02:00"},
  {"v": 3.14, "ts": "DATETIME25+02:00"},
  {"v": 3.14, "ts": "DATETIME26+02:00"},
  {"v": 3.14, "ts": "DATETIME27+02:00"},
  {"v": 3.14, "ts": "DATETIME28+02:00"},
  {"v": 3.14, "ts": "DATETIME29+02:00"},
  {"v": 3.14, "ts": "DATETIME30+02:00"},
  {"v": 3.14, "ts": "DATETIME31+02:00"},
  {"v": 3.14, "ts": "DATETIME32+02:00"},
  {"v": 3.14, "ts": "DATETIME33+02:00"},
  {"v": 3.14, "ts": "DATETIME34+02:00"},
  {"v": 3.14, "ts": "DATETIME35+02:00"},
  {"v": 3.14, "ts": "DATETIME36+02:00"},
  {"v": 3.14, "ts": "DATETIME37+02:00"},
  {"v": 3.14, "ts": "DATETIME38+02:00"},
  {"v": 3.14, "ts": "DATETIME39+02:00"},
  {"v": 3.14, "ts": "DATETIME40+02:00"},
  {"v": 3.14, "ts": "DATETIME41+02:00"},
  {"v": 3.14, "ts": "DATETIME42+02:00"},
  {"v": 3.14, "ts": "DATETIME43+02:00"},
  {"v": 3.14, "ts": "DATETIME44+02:00"},
  {"v": 3.14, "ts": "DATETIME45+02:00"},
  {"v": 3.14, "ts": "DATETIME46+02:00"},
  {"v": 3.14, "ts": "DATETIME47+02:00"},
  {"v": 3.14, "ts": "DATETIME48+02:00"},
  {"v": 3.14, "ts": "DATETIME49+02:00"},
  {"v": 3.14, "ts": "DATETIME50+02:00"},
  {"v": 3.14, "ts": "DATETIME51+02:00"},
  {"v": 3.14, "ts": "DATETIME52+02:00"},
  {"v": 3.14, "ts": "DATETIME53+02:00"},
  {"v": 3.14, "ts": "DATETIME54+02:00"},
  {"v": 3.14, "ts": "DATETIME55+02:00"},
  {"v": 3.14, "ts": "DATETIME56+02:00"},
  {"v": 3.14, "ts": "DATETIME57+02:00"},
  {"v": 3.14, "ts": "DATETIME58+02:00"},
  {"v": 3.14, "ts": "DATETIME59+02:00"},
  {"v": 3.14, "ts": "DATETIME60+02:00"},
  {"v": 3.14, "ts": "DATETIME61+02:00"},
  {"v": 3.14, "ts": "DATETIME62+02:00"},
  {"v": 3.14, "ts": "DATETIME63+02:00"},
  {"v": 3.14, "ts": "DATETIME64+02:00"},
  {"v": 3.14, "ts": "DATETIME65+02:00"},
  {"v": 3.14, "ts": "DATETIME66+02:00"},
  {"v": 3.14, "ts": "DATETIME67+02:00"},
  {"v": 3.14, "ts": "DATETIME68+02:00"},
  {"v": 3.14, "ts": "DATETIME69+02:00"},
  {"v": 3.14, "ts": "DATETIME70+02:00"},
  {"v": 3.14, "ts": "DATETIME71+02:00"},
  {"v": 3.14, "ts": "DATETIME72+02:00"},
  {"v": 3.14, "ts": "DATETIME73+02:00"},
  {"v": 3.14, "ts": "DATETIME74+02:00"},
  {"v": 3.14, "ts": "DATETIME75+02:00"},
  {"v": 3.14, "ts": "DATETIME76+02:00"},
  {"v": 3.14, "ts": "DATETIME77+02:00"},
  {"v": 3.14, "ts": "DATETIME78+02:00"},
  {"v": 3.14, "ts": "DATETIME79+02:00"},
  {"v": 3.14, "ts": "DATETIME80+02:00"},
  {"v": 3.14, "ts": "DATETIME81+02:00"},
  {"v": 3.14, "ts": "DATETIME82+02:00"},
  {"v": 3.14, "ts": "DATETIME83+02:00"},
  {"v": 3.14, "ts": "DATETIME84+02:00"},
  {"v": 3.14, "ts": "DATETIME85+02:00"},
  {"v": 3.14, "ts": "DATETIME86+02:00"},
  {"v": 3.14, "ts": "DATETIME87+02:00"},
  {"v": 3.14, "ts": "DATETIME88+02:00"},
  {"v": 3.14, "ts": "DATETIME89+02:00"},
  {"v": 3.14, "ts": "DATETIME90+02:00"},
  {"v": 3.14, "ts": "DATETIME91+02:00"},
  {"v": 3.14, "ts": "DATETIME92+02:00"},
  {"v": 3.14, "ts": "DATETIME93+02:00"},
  {"v": 3.14, "ts": "DATETIME94+02:00"},
  {"v": 3.14, "ts": "DATETIME95+02:00"},
  {"v": 3.14, "ts": "DATETIME96+02:00"},
  {"v": 3.14, "ts": "DATETIME97+02:00"},
  {"v": 3.14, "ts": "DATETIME98+02:00"},
  {"v": 3.14, "ts": "DATETIME99+02:00"}
]
```
