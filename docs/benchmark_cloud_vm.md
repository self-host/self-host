# Hardware: Cloud VM

Specification of hardware (an ordinary RPi4 8GB with SSD via USB 3.0) used during testing;

```txt
product: VC2
vendor: Vultr
*-core
  *-cpu
      product: Intel Xeon Processor (Cascadelake)
      capacity: 2GHz
  *-memory
      description: System Memory
      size: 8GiB
      capabilities: ecc
  *-virtio
       description: Virtual I/O device
       size: 160GiB (171GB)
```


# Populating the database

See [Benchmark details](https://github.com/self-host/self-host/blob/main/docs/benchmark_generating.md) for more info.

Generating this data takes about an hour and a half (90 minutes) on the target hardware. This time is due mainly to CPU limitations as the CPU pins to `100%` for the duration of the generation. The `52 561` random data points for each time series takes about 15 seconds to generate and insert.

Checking `iotop` reveals that the bottleneck was not related to the virtual I/O drive as the io-load remained low. It is all CPU related and is likely due to the function `tsdata_insert` or index creation, or both.

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
  Time (mean ± σ):      89.1 ms ±   1.7 ms    [User: 33.2 ms, System: 27.0 ms]
  Range (min … max):    85.3 ms …  91.5 ms    34 runs
```

Insertion of 10 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data10.json"
Benchmark #1: ./insert_data.sh data10.json
  Time (mean ± σ):      90.0 ms ±   4.4 ms    [User: 33.1 ms, System: 26.0 ms]
  Range (min … max):    67.8 ms …  94.5 ms    34 runs
```

Insertion of 100 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data100.json"
Benchmark #1: ./insert_data.sh data100.json
  Time (mean ± σ):     127.3 ms ±   6.6 ms    [User: 33.7 ms, System: 27.4 ms]
  Range (min … max):   121.4 ms … 154.2 ms    24 runs
```


## Retrieve data

Retrieving one day of data (144 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-02"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-02
  Time (mean ± σ):      74.8 ms ±   3.6 ms    [User: 19.5 ms, System: 9.6 ms]
  Range (min … max):    60.8 ms …  82.5 ms    38 runs
```

Retrieving one week of data (~1 000 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-08"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-08
  Time (mean ± σ):     106.3 ms ±   6.3 ms    [User: 19.4 ms, System: 11.2 ms]
  Range (min … max):    94.0 ms … 112.7 ms    27 runs
```

Retrieving one month of data (~4 400 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-02-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-02-01
  Time (mean ± σ):     164.1 ms ±   5.0 ms    [User: 25.0 ms, System: 13.4 ms]
  Range (min … max):   151.5 ms … 169.2 ms    18 runs
```

Retrieving one year of data (52 561 data points) from one time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2020-01-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2020-01-01
  Time (mean ± σ):     467.9 ms ±  37.3 ms    [User: 22.1 ms, System: 18.7 ms]
  Range (min … max):   432.1 ms … 540.2 ms    10 runs
```