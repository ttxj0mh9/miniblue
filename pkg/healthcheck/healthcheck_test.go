package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	c := NewChecker()
	if c == nil {
		t.Fatal("expected non-nil Checker")
	}
	if len(c.components) != 0 {
		t.Errorf("expected 0 components, got %d", len(c.components))
	}
}

func TestRegisterAndRunChecks_Healthy(t *testing.T) {
	c := NewChecker()
	c.Register("db", func() ComponentHealth {
		return ComponentHealth{Status: StatusHealthy, CheckedAt: time.Now()}
	})

	resp := c.RunChecks()
	if resp.Status != StatusHealthy {
		t.Errorf("expected healthy, got %s", resp.Status)
	}
	if _, ok := resp.Components["db"]; !ok {
		t.Error("expected 'db' component in response")
	}
}

func TestRunChecks_Unhealthy(t *testing.T) {
	c := NewChecker()
	c.Register("cache", func() ComponentHealth {
		return ComponentHealth{Status: StatusUnhealthy, Message: "connection refused", CheckedAt: time.Now()}
	})

	resp := c.RunChecks()
	if resp.Status != StatusUnhealthy {
		t.Errorf("expected unhealthy, got %s", resp.Status)
	}
}

func TestRunChecks_Degraded(t *testing.T) {
	c := NewChecker()
	c.Register("queue", func() ComponentHealth {
		return ComponentHealth{Status: StatusDegraded, Message: "high latency", CheckedAt: time.Now()}
	})
	c.Register("api", func() ComponentHealth {
		return ComponentHealth{Status: StatusHealthy, CheckedAt: time.Now()}
	})

	resp := c.RunChecks()
	if resp.Status != StatusDegraded {
		t.Errorf("expected degraded, got %s", resp.Status)
	}
}

func TestHTTPHandler_Returns200WhenHealthy(t *testing.T) {
	c := NewChecker()
	c.Register("svc", func() ComponentHealth {
		return ComponentHealth{Status: StatusHealthy, CheckedAt: time.Now()}
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rw := httptest.NewRecorder()
	c.HTTPHandler()(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}

	var body HealthResponse
	if err := json.NewDecoder(rw.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status != StatusHealthy {
		t.Errorf("expected healthy in body, got %s", body.Status)
	}
}

func TestHTTPHandler_Returns503WhenUnhealthy(t *testing.T) {
	c := NewChecker()
	c.Register("db", func() ComponentHealth {
		return ComponentHealth{Status: StatusUnhealthy, Message: "down", CheckedAt: time.Now()}
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rw := httptest.NewRecorder()
	c.HTTPHandler()(rw, req)

	if rw.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rw.Code)
	}
}

// TestHTTPHandler_Returns503WhenDegraded verifies that a degraded status also
// returns 503, since the service is not fully operational in that state.
func TestHTTPHandler_Returns503WhenDegraded(t *testing.T) {
	c := NewChecker()
	c.Register("queue", func() ComponentHealth {
		return ComponentHealth{Status: StatusDegraded, Message: "slow", CheckedAt: time.Now()}
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rw := httptest.NewRecorder()
	c.HTTPHandler()(rw, req)

	if rw.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for degraded status, got %d", rw.Code)
	}
}
