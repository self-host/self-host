# Rate control

The API server (Aapije) uses rate-control to mitigate DoS queries, help the
infrastructure to scale (by denying each client too many requests).

The solution used in Aapije is called the `token-bucket` principle and the
implementation is based on the library `golang.org/x/time/rate`.

For more details, we recommend that you read the Wikipedia [article](https://en.wikipedia.org/wiki/Token_bucket) on the subject.


# Configuration

The configuration is pretty straightforward.

We need to modify the configuration file for the Aapije service software.

```yaml
--
-- aapije.conf.yaml
--
rate_control:
  req_per_hour: 600
  maxburst: 10
  cleanup: "20minutes"
```

The options are;

- req_per_hour: Number of requests allowed for each client. Grouped by domain and API key.
- maxburst: The maximum allowed "burst" (consecutive request). Higher values allow more events to happen at once.
- cleanup: A timer after which the rate controller will reset.


# HTTP codes and headers

The error code `HTTP 429 Too Many Requests` will be returned when a client exceeds the rate limit. Along with the
HTTP header `X-RateLimit-Limit`, where the value is the maximum overall event rate (a rate-per-hour value).
