# Module - "http"

```golang
http := import("http")
```

## Parameters to functions

- `query_args []string`: an array of strings on the format `key=value`.
- `url string`: the target URL.
- `header map[string]string`: a map of HTTP header. For example `{"Content-Type": "application/json"}`.
- `content string`: the body content.

## Return Objects

- `response`: an immutable `map[string]string` with the keys: `StatusCode (int)`, `Body (bytes)`, `ContentLength (int)` and `Header (map[string][]string)`.

## Functions

- `delete(url string, query_args []string, header map[string]string, content string) => response`: execute a DELETE request.
- `get(url string, query_args []string, header map[string]string) => response`: execute a GET request.
- `put(url string, query_args []string, header map[string]string, content string) => response`: execute a PUT request.
- `post(url string, query_args []string, header map[string]string, content string) => response`: execute a POST request.
- `postform(url string, query_args []string, header map[string]string) => response`: execute a POST request, with the query_args transformed into the body and encoded as `application/x-www-form-urlencoded`
- `toformdata(query_args []string) => string`: converts query_args to a request body for use with a POST request with the header `Content-Type: application/x-www-form-urlencoded`.
