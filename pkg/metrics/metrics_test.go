package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCounter_IncAndValue(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("requests")
	if c.Value() != 0 {
		t.Fatalf("expected 0, got %d", c.Value())
	}
	c.Inc()
	c.Inc()
	if c.Value() != 2 {
		t.Fatalf("expected 2, got %d", c.Value())
	}
}

func TestCounter_SameNameReturnsSameInstance(t *testing.T) {
	r := NewRegistry()
	c1 := r.Counter("hits")
	c2 := r.Counter("hits")
	c1.Inc()
	if c2.Value() != 1 {
		t.Fatal("expected same counter instance for same name")
	}
}

func TestGauge_SetAndValue(t *testing.T) {
	r := NewRegistry()
	g := r.Gauge("cpu")
	g.Set(42.5)
	if g.Value() != 42.5 {
		t.Fatalf("expected 42.5, got %g", g.Value())
	}
	g.Set(0)
	if g.Value() != 0 {
		t.Fatalf("expected 0, got %g", g.Value())
	}
}

func TestGauge_SameNameReturnsSameInstance(t *testing.T) {
	r := NewRegistry()
	g1 := r.Gauge("mem")
	g2 := r.Gauge("mem")
	g1.Set(99.9)
	if g2.Value() != 99.9 {
		t.Fatal("expected same gauge instance for same name")
	}
}

func TestUptimeSeconds_NonNegative(t *testing.T) {
	r := NewRegistry()
	if r.UptimeSeconds() < 0 {
		t.Fatal("uptime should be non-negative")
	}
}

// TestHTTPHandler_ContainsMetrics verifies that the /metrics endpoint returns
// all registered counters, gauges, and the uptime field with a 200 status.
func TestHTTPHandler_ContainsMetrics(t *testing.T) {
	r := NewRegistry()
	r.Counter("requests").Inc()
	r.Counter("requests").Inc()
	r.Gauge("cpu").Set(55.0)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	r.HTTPHandler()(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "counter_requests 2") {
		t.Errorf("expected counter_requests 2 in body, got:\n%s", body)
	}
	if !strings.Contains(body, "gauge_cpu 55") {
		t.Errorf("expected gauge_cpu 55 in body, got:\n%s", body)
	}
	if !strings.Contains(body, "uptime_seconds") {
		t.Errorf("expected uptime_seconds in body, got:\n%s", body)
	}
	if rr.Code != 200 {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
