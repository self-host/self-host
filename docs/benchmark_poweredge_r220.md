# Hardware: PowerEdge R220

Specification of hardware (a mid-range 1U server, released back in 2014) used during testing;

```txt
product: PowerEdge R220
vendor: Dell Inc.
*-core
  *-cpu
      product: Intel(R) Xeon(R) CPU E3-1220 v3 @ 3.10GHz
      capacity: 3700MHz
  *-memory
      description: System Memory
      slot: System board or motherboard
      size: 16GiB
      capabilities: ecc
  *-raid
      description: RAID bus controller
      product: MegaRAID SAS 2008 [Falcon]
    *-disk
        product: PERC H310
        vendor: DELL
        size: 931GiB (999GB)
        subsystem: TOSHIBA mechanical drive
```


# Populating the database

See [Benchmark details](https://github.com/self-host/self-host/blob/main/docs/benchmark_generating.md) for more info.

Generating this data takes about an hour (60 minutes) on the target hardware. This time is due mainly to CPU limitations as the CPU pins to `100%` for the duration of the generation. The `52 561` random data points for each time series takes about 3-4 seconds to generate and insert.

Checking `iotop` reveals that the bottleneck was not related to the mechanical drives as the io-load remained low. It is all CPU related and is likely due to the function `tsdata_insert` or index creation, or both.

By ensuring all partition tables exist beforehand and using direct insert statements instead of using the `tsdata_insert` function, we can cut the time required by about half.

Removing or adding indexes hardly affected the execution time.


# Benchmarks

To aid with benchmarking, we use `curl` wrapped in different shell scripts and the tool `hyperfine` to help us determine the average execution time.


## Store data

Insertion of a single datapoint;

```
hyperfine --warmup 10 "./insert_data.sh data1.json"
Benchmark #1: ./insert_data.sh data1.json
  Time (mean ± σ):      22.6 ms ±   8.8 ms    [User: 16.2 ms, System: 7.4 ms]
  Range (min … max):    13.3 ms …  59.2 ms    151 runs
```

Insertion of 10 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data10.json"
Benchmark #1: ./insert_data.sh data10.json
  Time (mean ± σ):      26.8 ms ±  10.0 ms    [User: 18.0 ms, System: 8.8 ms]
  Range (min … max):    13.5 ms …  61.8 ms    141 runs
```

Insertion of 100 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data100.json"
Benchmark #1: ./insert_data.sh data100.json
  Time (mean ± σ):      49.1 ms ±   8.3 ms    [User: 25.1 ms, System: 14.0 ms]
  Range (min … max):    38.1 ms …  74.7 ms    57 runs
```


## Retrieve data

Retrieving one day of data (144 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-02"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-02
  Time (mean ± σ):      21.5 ms ±   8.5 ms    [User: 11.2 ms, System: 5.2 ms]
  Range (min … max):    12.5 ms …  44.2 ms    130 runs
```

Retrieving one week of data (~1 000 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-08"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-08
  Time (mean ± σ):      27.4 ms ±   7.2 ms    [User: 12.8 ms, System: 6.2 ms]
  Range (min … max):    17.1 ms …  51.5 ms    104 runs
```

Retrieving one month of data (~4 400 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-02-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-02-01
  Time (mean ± σ):      43.3 ms ±   8.0 ms    [User: 17.0 ms, System: 8.0 ms]
  Range (min … max):    32.2 ms …  61.6 ms    78 runs
```

Retrieving one year of data (52 561 data points) from one time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2020-01-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2020-01-01
  Time (mean ± σ):     174.2 ms ±   4.9 ms    [User: 22.7 ms, System: 22.8 ms]
  Range (min … max):   163.8 ms … 183.2 ms    16 runs
```
