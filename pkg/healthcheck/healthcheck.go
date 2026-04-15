package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// ComponentHealth holds the health info for a single component
type ComponentHealth struct {
	Status    Status    `json:"status"`
	Message   string    `json:"message,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// HealthResponse is the full health check response payload
type HealthResponse struct {
	Status     Status                     `json:"status"`
	Components map[string]ComponentHealth `json:"components"`
	Timestamp  time.Time                  `json:"timestamp"`
}

// Checker manages health checks for multiple components
type Checker struct {
	mu         sync.RWMutex
	components map[string]func() ComponentHealth
}

// NewChecker creates a new Checker instance
func NewChecker() *Checker {
	return &Checker{
		components: make(map[string]func() ComponentHealth),
	}
}

// Register adds a named health check function
func (c *Checker) Register(name string, fn func() ComponentHealth) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.components[name] = fn
}

// RunChecks executes all registered checks and returns the aggregate result
func (c *Checker) RunChecks() HealthResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := HealthResponse{
		Status:     StatusHealthy,
		Components: make(map[string]ComponentHealth),
		Timestamp:  time.Now(),
	}

	for name, fn := range c.components {
		h := fn()
		result.Components[name] = h
		if h.Status == StatusUnhealthy {
			result.Status = StatusUnhealthy
		} else if h.Status == StatusDegraded && result.Status != StatusUnhealthy {
			result.Status = StatusDegraded
		}
	}

	return result
}

// HTTPHandler returns an http.HandlerFunc for the health endpoint
func (c *Checker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := c.RunChecks()
		w.Header().Set("Content-Type", "application/json")
		if resp.Status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}
