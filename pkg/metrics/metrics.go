package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Counter is a simple monotonically increasing counter.
type Counter struct {
	mu    sync.Mutex
	value int64
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Add increments the counter by n. Negative values are ignored.
func (c *Counter) Add(n int64) {
	if n <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += n
}

// Value returns the current counter value.
func (c *Counter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Gauge holds a float64 value that can go up or down.
type Gauge struct {
	mu    sync.Mutex
	value float64
}

// Set sets the gauge to the given value.
func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = v
}

// Value returns the current gauge value.
func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Registry holds named counters and gauges.
type Registry struct {
	mu       sync.RWMutex
	counters map[string]*Counter
	gauges   map[string]*Gauge
	start    time.Time
}

// NewRegistry creates a new metrics Registry.
func NewRegistry() *Registry {
	return &Registry{
		counters: make(map[string]*Counter),
		gauges:   make(map[string]*Gauge),
		start:    time.Now(),
	}
}

// Counter returns (or creates) a named counter.
func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{}
	r.counters[name] = c
	return c
}

// Gauge returns (or creates) a named gauge.
func (r *Registry) Gauge(name string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.gauges[name]; ok {
		return g
	}
	g := &Gauge{}
	r.gauges[name] = g
	return g
}

// UptimeSeconds returns seconds since the registry was created.
func (r *Registry) UptimeSeconds() float64 {
	return time.Since(r.start).Seconds()
}

// HTTPHandler returns an http.HandlerFunc that exposes metrics as plain text.
func (r *Registry) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		r.mu.RLock()
		defer r.mu.RUnlock()
		for name, c := range r.counters {
			fmt.Fprintf(w, "counter_%s %d\n", name, c.Value())
		}
		for name, g := range r.gauges {
			fmt.Fprintf(w, "gauge_%s %g\n", name, g.Value())
		}
		fmt.Fprintf(w, "uptime_seconds %g\n", r.UptimeSeconds())
	}
}
