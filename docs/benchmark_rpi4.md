# Hardware: Raspberry Pi 4 (8GB)

Specification of hardware (an ordinary RPi4 8GB with SSD via USB 3.0) used during testing;

```txt
product: Raspberry Pi 4 Model B Rev 1.4
vendor: Raspberry Pi Foundation
*-core
  *-cpu
      product: Broadcom BCM2711 SoC
      capacity: 1500MHz
  *-memory
      description: System Memory
      size: 8GiB
  *-usb
      description: Mass storage device
      product: USB3.0 to SATA adapter
      vendor: JMicron
    *-disk
        description: SCSI Disk
        vendor: KINGSTON
        size: 111GiB (120GB)
```


# Populating the database

See [Benchmark details](https://github.com/self-host/self-host/blob/main/docs/benchmark_generating.md) for more info.

Generating this data takes about five hours (300 minutes) on the target hardware. This time is due mainly to CPU limitations as the CPU pins to `100%` for the duration of the generation. The `52 561` random data points for each time series takes about 15 seconds to generate and insert.

Checking `iotop` reveals that the bottleneck was not related to the USB connected SSD as the io-load remained low. It is all CPU related and is likely due to the function `tsdata_insert` or index creation, or both.

By ensuring all partition tables exist beforehand and using direct insert statements instead of using the `tsdata_insert` function, we can cut the time required by about half.

Removing or adding indexes hardly affected the execution time.


# Benchmarks

To aid with benchmarking, we use `curl` wrapped in different shell scripts and the tool `hyperfine` to help us determine the average execution time.

Insertion of a single datapoint;

```
hyperfine --warmup 10 "./insert_data.sh data1.json"
Benchmark #1: ./insert_data.sh data1.json
  Time (mean ± σ):      41.3 ms ±   6.0 ms    [User: 24.2 ms, System: 14.6 ms]
  Range (min … max):    30.1 ms …  53.6 ms    55 runs
```

Insertion of 10 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data10.json"
Benchmark #1: ./insert_data.sh data10.json
  Time (mean ± σ):      48.3 ms ±   9.0 ms    [User: 27.3 ms, System: 15.4 ms]
  Range (min … max):    26.1 ms …  61.3 ms    64 runs
```

Insertion of 100 data points at a time;

```
hyperfine --warmup 10 "./insert_data.sh data100.json"
Benchmark #1: ./insert_data.sh data100.json
  Time (mean ± σ):      88.4 ms ±   1.7 ms    [User: 33.6 ms, System: 21.5 ms]
  Range (min … max):    86.3 ms …  92.9 ms    32 runs
```


## Retrieve data

Retrieving one day of data (144 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-02"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-02
  Time (mean ± σ):      28.3 ms ±   7.3 ms    [User: 13.1 ms, System: 6.3 ms]
  Range (min … max):    12.1 ms …  43.5 ms    100 runs
```

Retrieving one week of data (~1 000 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-01-08"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-01-08
  Time (mean ± σ):      39.9 ms ±   6.6 ms    [User: 16.4 ms, System: 7.9 ms]
  Range (min … max):    21.6 ms …  50.8 ms    92 runs
```

Retrieving one month of data (~4 400 data points) from a single time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2019-02-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2019-02-01
  Time (mean ± σ):      77.6 ms ±   2.5 ms    [User: 19.2 ms, System: 11.8 ms]
  Range (min … max):    68.6 ms …  81.3 ms    37 runs
```

Retrieving one year of data (52 561 data points) from one time series;

```
hyperfine --warmup 10 "./get_data_period.sh 2019-01-01 2020-01-01"
Benchmark #1: ./get_data_period.sh 2019-01-01 2020-01-01
  Time (mean ± σ):     684.3 ms ±  86.3 ms    [User: 15.9 ms, System: 13.0 ms]
  Range (min … max):   497.8 ms … 769.5 ms    10 runs

```
