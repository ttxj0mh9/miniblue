# ratelimit

Package `ratelimit` provides a simple token-bucket rate limiter and an HTTP
middleware that enforces it.

## Usage

### Create a limiter

```go
import "github.com/moabukar/miniblue/pkg/ratelimit"

// Allow up to 50 requests per second with a burst of 10.
limiter := ratelimit.NewLimiter(50, 10)
```

### Check programmatically

```go
if !limiter.Allow() {
    // request denied
}
```

### HTTP middleware

Wrap any `http.Handler` to automatically reject excess requests with
`429 Too Many Requests`:

```go
http.Handle("/api", limiter.Middleware(myHandler))
```

## Parameters

| Parameter | Description |
|-----------|-------------|
| `rate`    | Sustained requests allowed per second (token refill rate). |
| `burst`   | Maximum number of tokens (requests) that can accumulate. |

## Thread Safety

The `Limiter` is safe for concurrent use from multiple goroutines.
