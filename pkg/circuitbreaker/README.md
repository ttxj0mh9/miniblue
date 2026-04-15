# circuitbreaker

A lightweight circuit breaker package for miniblue.

## Overview

The circuit breaker prevents cascading failures by tracking consecutive errors and temporarily blocking calls when a failure threshold is reached.

## States

| State | Description |
|-------|-------------|
| **Closed** | Normal operation — calls pass through. |
| **Open** | Failure threshold exceeded — calls are rejected immediately with `ErrCircuitOpen`. |
| **Half-Open** | Reset timeout elapsed — one probe call is allowed to test recovery. |

## Usage

```go
import "github.com/moabukar/miniblue/pkg/circuitbreaker"

// Create a breaker: open after 5 failures, attempt reset after 10 seconds.
cb := circuitbreaker.New(5, 10*time.Second)

err := cb.Execute(func() error {
    return callExternalService()
})
if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
    // circuit is open, serve fallback
}
```

## API

### `New(maxFailures int, resetTimeout time.Duration) *Breaker`

Creates a new circuit breaker.

### `(*Breaker) Execute(fn func() error) error`

Runs `fn` if the circuit allows it. Tracks success/failure and transitions state.

### `(*Breaker) State() State`

Returns the current `State` (`StateClosed`, `StateOpen`, or `StateHalfOpen`).
