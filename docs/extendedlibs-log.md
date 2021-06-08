# Module - "log"

```golang
log := import("log")
```

Uses [Zap](https://github.com/uber-go/zap) under the hood for logging.

## Functions

- `info(format, args...)`: Info level logging. The first argument must be a String object. See
  [this](https://github.com/d5/tengo/blob/master/docs/formatting.md) for more
  details on formatting.
- `error(format, args...)`: Error level logging. The first argument must be a String object. See
  [this](https://github.com/d5/tengo/blob/master/docs/formatting.md) for more
  details on formatting.

