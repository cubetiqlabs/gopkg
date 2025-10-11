package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Counter is an atomic counter for metrics.
type Counter struct {
	v uint64
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	atomic.AddUint64(&c.v, 1)
}

// Add increments the counter by delta.
func (c *Counter) Add(delta uint64) {
	atomic.AddUint64(&c.v, delta)
}

// Get returns the current counter value.
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.v)
}

// Histogram tracks a distribution of values (simple sum + count for average).
// Can be extended with buckets for percentiles if needed.
type Histogram struct {
	sum   uint64
	count uint64
}

// Observe records a value in milliseconds.
func (h *Histogram) Observe(ms int64) {
	atomic.AddUint64(&h.sum, uint64(ms))
	atomic.AddUint64(&h.count, 1)
}

// Avg returns the average value.
func (h *Histogram) Avg() float64 {
	c := atomic.LoadUint64(&h.count)
	if c == 0 {
		return 0
	}
	s := atomic.LoadUint64(&h.sum)
	return float64(s) / float64(c)
}

// Count returns the number of observations.
func (h *Histogram) Count() uint64 {
	return atomic.LoadUint64(&h.count)
}

// Sum returns the sum of all observations.
func (h *Histogram) Sum() uint64 {
	return atomic.LoadUint64(&h.sum)
}

// Registry holds metrics for an application.
// It provides common metrics out of the box and supports custom labeled metrics.
type Registry struct {
	// HTTP metrics
	RequestsTotal   *Counter   // Total HTTP requests
	RequestDuration *Histogram // HTTP request duration in milliseconds
	
	// Rate limiting metrics
	RateAllowed  *Counter // Requests allowed by rate limiter
	RateRejected *Counter // Requests rejected by rate limiter
	
	// gRPC metrics
	GrpcRequests *Counter   // Total gRPC requests
	GrpcDuration *Histogram // gRPC request duration in milliseconds
	
	// System metrics
	Started time.Time // When the registry was created
	
	// Custom labeled metrics
	mu      sync.RWMutex
	labeled map[string]*Counter // key: metric|labelString
}

// NewRegistry creates a new metrics registry with initialized counters and histograms.
func NewRegistry() *Registry {
	return &Registry{
		RequestsTotal:   &Counter{},
		RequestDuration: &Histogram{},
		RateAllowed:     &Counter{},
		RateRejected:    &Counter{},
		GrpcRequests:    &Counter{},
		GrpcDuration:    &Histogram{},
		Started:         time.Now().UTC(),
		labeled:         make(map[string]*Counter),
	}
}

// IncLabeled increments a labeled counter for the given metric name and label map.
// Labels are automatically sorted for consistent key generation.
//
// Example:
//
//	reg.IncLabeled("http_requests", map[string]string{
//	    "method": "GET",
//	    "path":   "/api/users",
//	    "status": "200",
//	})
func (r *Registry) IncLabeled(metric string, labels map[string]string) {
	// Generate stable key from sorted labels
	key := buildLabelKey(metric, labels)
	
	// Fast path: read lock first
	r.mu.RLock()
	c, ok := r.labeled[key]
	r.mu.RUnlock()
	
	if !ok {
		// Slow path: write lock to create counter
		r.mu.Lock()
		// Double-check after acquiring write lock
		if c, ok = r.labeled[key]; !ok {
			c = &Counter{}
			r.labeled[key] = c
		}
		r.mu.Unlock()
	}
	
	c.Inc()
}

// AddLabeled adds delta to a labeled counter.
func (r *Registry) AddLabeled(metric string, labels map[string]string, delta uint64) {
	key := buildLabelKey(metric, labels)
	
	r.mu.RLock()
	c, ok := r.labeled[key]
	r.mu.RUnlock()
	
	if !ok {
		r.mu.Lock()
		if c, ok = r.labeled[key]; !ok {
			c = &Counter{}
			r.labeled[key] = c
		}
		r.mu.Unlock()
	}
	
	c.Add(delta)
}

// buildLabelKey generates a consistent key for labeled metrics.
// Format: metric|key1=value1,key2=value2 (sorted by key)
func buildLabelKey(metric string, labels map[string]string) string {
	if len(labels) == 0 {
		return metric
	}
	
	// Sort keys for consistency
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Build label string
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+labels[k])
	}
	
	return metric + "|" + strings.Join(parts, ",")
}

// RenderPrometheus outputs metrics in Prometheus text format.
// This can be exposed on a /metrics endpoint for scraping.
//
// Example output:
//
//	http_requests_total 12345
//	http_request_duration_ms_avg 45.67
//	uptime_seconds 3600
//	custom_metric{label1="value1",label2="value2"} 42
func (r *Registry) RenderPrometheus() string {
	uptime := time.Since(r.Started).Seconds()
	
	sb := &strings.Builder{}
	
	// Base metrics
	fmt.Fprintf(sb, "http_requests_total %d\n", r.RequestsTotal.Get())
	fmt.Fprintf(sb, "http_request_duration_ms_avg %.2f\n", r.RequestDuration.Avg())
	fmt.Fprintf(sb, "http_request_duration_ms_sum %d\n", r.RequestDuration.Sum())
	fmt.Fprintf(sb, "http_request_duration_ms_count %d\n", r.RequestDuration.Count())
	fmt.Fprintf(sb, "rate_allowed_total %d\n", r.RateAllowed.Get())
	fmt.Fprintf(sb, "rate_rejected_total %d\n", r.RateRejected.Get())
	fmt.Fprintf(sb, "uptime_seconds %.0f\n", uptime)
	fmt.Fprintf(sb, "grpc_requests_total %d\n", r.GrpcRequests.Get())
	fmt.Fprintf(sb, "grpc_request_duration_ms_avg %.2f\n", r.GrpcDuration.Avg())
	
	// Labeled metrics
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for key, counter := range r.labeled {
		// Parse key: metric|label1=value1,label2=value2
		parts := strings.SplitN(key, "|", 2)
		metric := parts[0]
		lbls := ""
		
		if len(parts) == 2 && parts[1] != "" {
			// Convert label string to Prometheus format: {label1="value1",label2="value2"}
			lblPairs := strings.Split(parts[1], ",")
			for i, p := range lblPairs {
				lblPairs[i] = strings.ReplaceAll(p, "=", "=\"") + "\""
			}
			lbls = "{" + strings.Join(lblPairs, ",") + "}"
		}
		
		fmt.Fprintf(sb, "%s%s %d\n", metric, lbls, counter.Get())
	}
	
	return sb.String()
}

// Reset resets all metrics to zero. Useful for testing.
func (r *Registry) Reset() {
	r.RequestsTotal = &Counter{}
	r.RequestDuration = &Histogram{}
	r.RateAllowed = &Counter{}
	r.RateRejected = &Counter{}
	r.GrpcRequests = &Counter{}
	r.GrpcDuration = &Histogram{}
	
	r.mu.Lock()
	r.labeled = make(map[string]*Counter)
	r.mu.Unlock()
}
