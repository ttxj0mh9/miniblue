# healthcheck

The `healthcheck` package provides a lightweight, composable health-check system for miniblue services.

## Features

- Register named component checks (database, cache, queues, etc.)
- Aggregate status: `healthy`, `degraded`, or `unhealthy`
- HTTP handler ready for `/healthz` or `/readyz` endpoints
- Thread-safe check registration and execution

## Usage

```go
import "github.com/moabukar/miniblue/pkg/healthcheck"

checker := healthcheck.NewChecker()

// Register a database check
checker.Register("database", func() healthcheck.ComponentHealth {
    err := db.Ping()
    if err != nil {
        return healthcheck.ComponentHealth{
            Status:    healthcheck.StatusUnhealthy,
            Message:   err.Error(),
            CheckedAt: time.Now(),
        }
    }
    return healthcheck.ComponentHealth{
        Status:    healthcheck.StatusHealthy,
        CheckedAt: time.Now(),
    }
})

// Mount the HTTP handler
http.HandleFunc("/healthz", checker.HTTPHandler())
```

## Response Format

```json
{
  "status": "healthy",
  "components": {
    "database": {
      "status": "healthy",
      "checked_at": "2024-01-15T10:00:00Z"
    }
  },
  "timestamp": "2024-01-15T10:00:00Z"
}
```

## Status Codes

| Status      | HTTP Code |
|-------------|----------|
| healthy     | 200       |
| degraded    | 200       |
| unhealthy   | 503       |
