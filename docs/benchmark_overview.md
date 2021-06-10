# Benchmark overview

Hardware used;

- [PowerEdge R220 (mid-range 1U server, released back in 2014)](https://github.com/self-host/self-host/blob/main/docs/benchmark_poweredge_r220.md)
- [RaspberryPi 4 8GB](https://github.com/self-host/self-host/blob/main/docs/benchmark_rpi4.md)
- [Cloud hosted virtual machine](https://github.com/self-host/self-host/blob/main/docs/benchmark_cloud_vm.md)


# Results

|                   |   RPi4   |   R220   | CloudVM* |
|-------------------|:--------:|:--------:|:--------:|
| Get 1day          |  28.3 ms |  21.5 ms |  70.9 ms |
| Get 1week         |  39.9 ms |  27.4 ms |  94.9 ms |
| Get 1month        |  77.6 ms |  43.3 ms | 138.9 ms |
| Get 1year         | 684.3 ms | 174.2 ms | 431.9 ms |
| Insert 1 point    |  41.3 ms |  22.6 ms |  86.2 ms |
| Insert 10 points  |  48.3 ms |  26.8 ms |  88.7 ms |
| Insert 100 points |  88.4 ms |  49.1 ms | 121.2 ms |

* We accessed the machine over the internet.


# Stress test

Using [Locust](https://locust.io/), we have managed to execute four different requests (see below), at a total rate of about 200 requests per second on the `PowerEdge R220` hardware. Extrapolated; `12 000` request per minute and `720 000` request per hour. With a CPU load of about 90% (all cores) and memory consumption of around 20%.

- `GET` on /v2/timeseries/029a530b-8432-4a93-8903-24044c389b50/data?start=2019-01-01T00:00:00Z&end=2019-02-01T00:00:00Z
- `GET` on /v2/timeseries/af2a2145-9c5b-44db-b1f0-a5b92d0230cf/data?start=2019-01-01T00:00:00Z&end=2019-01-01T00:00:00Z
- `GET` on /v2/timeseries/e5b3642d-8d0b-4abf-80bf-33e24d4cb9d8/data?start=2019-01-01T00:00:00Z&end=2019-04-01T00:00:00Z
- `GET` on /v2/users?limit=100

All three of the time series contained `52 561` data points spread over the range `2019-01-01 to 2020-01-01`.

