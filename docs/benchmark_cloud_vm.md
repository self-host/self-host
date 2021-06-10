# Hardware: Cloud VM

Specification of hardware (a hosted Database and single Droplet at DigitalOcean) used during testing;

```txt
product: Droplet
vendor: DigitalOcean
*-core
  *-cpu
      product: DO-Premium-Intel (?)
      capacity: 2GHz
  *-memory
      description: System Memory
      size: 8GiB
      capabilities: ecc
  *-virtio
       description: Virtual I/O device
       size: 160GiB (171GB)

product: DO Database
vendor: DigitalOcean
*-core:
  *-cpu
      product: Intel Xeon (Unknown)
      cores: 6vCPU
  *-memory
      description: System Memory
      size: 16GiB
  *-disk
      description: Virtual I/O device
      size: 270GB
```


# Populating the database

See [Benchmark details](https://github.com/self-host/self-host/blob/main/docs/benchmark_generating.md) for more info.

Generating this data takes about an hour (60 minutes) on the target hardware. This time is due mainly to CPU limitations as the CPU pins to `100%` for the duration of the generation. The `52 561` random data points for each time series takes about 15 seconds to generate and insert.

Checking DB stats reveals that the mean insert time was `1388.3088ms`. It seems to be CPU related and is likely due to the function `tsdata_insert` or index creation, or both.

By ensuring all partition tables exist beforehand and using direct insert statements instead of using the `tsdata_insert` function, we can cut the time required by about half.

Removing or adding indexes hardly affected the execution time.


# Benchmarks

To aid with benchmarking, we use `curl` wrapped in different shell scripts and the tool `hyperfine` to help us determine the average execution time.

**NOTE:** the following request where done over the internet and has an inherit delay of having to traverse the internet.


## Store data

Insertion of a single datapoint;

```
hyperfine --warmup 10 "./insert_data.sh data1.json"
Benchmark #1: ./insert_data.sh data1.json
  Time (mean ± σ):      86.2 ms ±   1.1 ms    [User: 30.7 ms, System: 23.5 ms]
  Range (min … max):    84.7 ms …  90.0 ms    33 runs
```

Insertion of 10 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data10.json"
Benchmark #1: ./insert_data.sh data10.json
  Time (mean ± σ):      88.7 ms ±   6.8 ms    [User: 31.9 ms, System: 21.2 ms]
  Range (min … max):    58.8 ms … 103.5 ms    33 runs
```

Insertion of 100 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data100.json"
Benchmark #1: ./insert_data.sh data100.json
  Time (mean ± σ):     121.2 ms ±  12.8 ms    [User: 32.3 ms, System: 22.5 ms]
  Range (min … max):    96.4 ms … 172.4 ms    24 runs
```


## Retrieve data

Retrieving one day of data (144 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-02"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-02
  Time (mean ± σ):      70.9 ms ±   5.8 ms    [User: 15.1 ms, System: 9.9 ms]
  Range (min … max):    53.7 ms …  80.2 ms    39 runs
```

Retrieving one week of data (~1 000 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-08"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-08
  Time (mean ± σ):      94.9 ms ±   4.6 ms    [User: 16.7 ms, System: 10.1 ms]
  Range (min … max):    84.0 ms …  99.8 ms    30 runs
```

Retrieving one month of data (~4 400 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-02-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-02-01
  Time (mean ± σ):     138.9 ms ±  10.5 ms    [User: 12.1 ms, System: 8.2 ms]
  Range (min … max):   127.4 ms … 154.7 ms    22 runs
```

Retrieving one year of data (52 561 data points) from one time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2020-01-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2020-01-01
  Time (mean ± σ):     431.9 ms ±  14.5 ms    [User: 24.8 ms, System: 12.8 ms]
  Range (min … max):   411.3 ms … 455.0 ms    10 runs
```
