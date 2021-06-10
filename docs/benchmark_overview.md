# Benchmark overview

Hardwares used;

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

*The machine was access over the internet.

