# metrics

A lightweight in-process metrics package for **miniblue**.

## Features

- **Counter** — monotonically increasing integer counter (e.g. request count).
- **Gauge** — float64 value that can be set arbitrarily (e.g. CPU usage).
- **Registry** — thread-safe store for named counters and gauges.
- **HTTP handler** — exposes all metrics as plain text on any HTTP endpoint.

## Usage

```go
import "github.com/moabukar/miniblue/pkg/metrics"

reg := metrics.NewRegistry()

// Increment a counter
reg.Counter("http_requests").Inc()

// Increment by a specific value
reg.Counter("bytes_received").Add(512)

// Set a gauge
reg.Gauge("goroutines").Set(float64(runtime.NumGoroutine()))

// Expose over HTTP
http.Handle("/metrics", reg.HTTPHandler())
```

## Output format

The HTTP handler returns plain text in the following format:

```
counter_http_requests 42
counter_bytes_received 512
gauge_goroutines 8
uptime_seconds 123.456
```

> **Note:** Metrics are sorted alphabetically in the output, which makes it easier to diff snapshots.

## Running tests

```bash
go test ./pkg/metrics/...
```
