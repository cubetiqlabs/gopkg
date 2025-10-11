package metrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter_Inc(t *testing.T) {
	c := &Counter{}
	
	assert.Equal(t, uint64(0), c.Get())
	
	c.Inc()
	assert.Equal(t, uint64(1), c.Get())
	
	c.Inc()
	assert.Equal(t, uint64(2), c.Get())
}

func TestCounter_Add(t *testing.T) {
	c := &Counter{}
	
	c.Add(5)
	assert.Equal(t, uint64(5), c.Get())
	
	c.Add(10)
	assert.Equal(t, uint64(15), c.Get())
}

func TestCounter_ConcurrentInc(t *testing.T) {
	c := &Counter{}
	iterations := 1000
	
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				c.Inc()
			}
			done <- true
		}()
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
	
	assert.Equal(t, uint64(10*iterations), c.Get())
}

func TestHistogram_Observe(t *testing.T) {
	h := &Histogram{}
	
	h.Observe(10)
	h.Observe(20)
	h.Observe(30)
	
	avg := h.Avg()
	assert.Equal(t, 20.0, avg)
}

func TestHistogram_Avg(t *testing.T) {
	tests := []struct {
		name     string
		values   []int64
		expected float64
	}{
		{"single value", []int64{42}, 42.0},
		{"multiple values", []int64{10, 20, 30}, 20.0},
		{"empty", []int64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Histogram{}
			for _, v := range tt.values {
				h.Observe(v)
			}
			
			avg := h.Avg()
			assert.InDelta(t, tt.expected, avg, 0.0001)
		})
	}
}

func TestHistogram_ConcurrentObserve(t *testing.T) {
	h := &Histogram{}
	iterations := 1000
	
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				h.Observe(10)
			}
			done <- true
		}()
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
	
	avg := h.Avg()
	assert.Equal(t, 10.0, avg)
}

func TestRegistry_IncLabeled(t *testing.T) {
	r := NewRegistry()
	
	r.IncLabeled("test_metric", map[string]string{"status": "200", "method": "GET"})
	r.IncLabeled("test_metric", map[string]string{"status": "200", "method": "GET"})
	r.IncLabeled("test_metric", map[string]string{"status": "404", "method": "GET"})
	r.IncLabeled("test_metric", map[string]string{"status": "200", "method": "POST"})
	
	// Verify via Prometheus output
	output := r.RenderPrometheus()
	assert.Contains(t, output, `test_metric{method="GET",status="200"} 2`)
	assert.Contains(t, output, `test_metric{method="GET",status="404"} 1`)
	assert.Contains(t, output, `test_metric{method="POST",status="200"} 1`)
}

func TestRegistry_AddLabeled(t *testing.T) {
	r := NewRegistry()
	
	r.AddLabeled("test_metric", map[string]string{"type": "user"}, 5)
	r.AddLabeled("test_metric", map[string]string{"type": "user"}, 10)
	r.AddLabeled("test_metric", map[string]string{"type": "admin"}, 3)
	
	output := r.RenderPrometheus()
	assert.Contains(t, output, `test_metric{type="user"} 15`)
	assert.Contains(t, output, `test_metric{type="admin"} 3`)
}

func TestRenderPrometheus(t *testing.T) {
	r := NewRegistry()
	
	r.RequestsTotal.Add(42)
	r.RequestDuration.Observe(100)
	r.RequestDuration.Observe(200)
	r.RateAllowed.Add(10)
	r.RateRejected.Add(2)
	
	output := r.RenderPrometheus()
	
	// Check base metrics
	assert.Contains(t, output, "http_requests_total 42")
	assert.Contains(t, output, "http_request_duration_ms_avg 150.00")
	assert.Contains(t, output, "rate_allowed_total 10")
	assert.Contains(t, output, "rate_rejected_total 2")
	assert.Contains(t, output, "uptime_seconds")
}

func TestRenderPrometheus_Format(t *testing.T) {
	r := NewRegistry()
	
	r.IncLabeled("test_metric", map[string]string{"method": "GET", "status": "200"})
	
	output := r.RenderPrometheus()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Should have the labeled metric line
	found := false
	for _, line := range lines {
		if strings.Contains(line, `test_metric{method="GET",status="200"} 1`) {
			found = true
			break
		}
	}
	
	assert.True(t, found, "Should contain properly formatted labeled metric")
}

func TestRenderPrometheus_Sorting(t *testing.T) {
	r := NewRegistry()
	
	r.IncLabeled("test_metric", map[string]string{"b": "2", "a": "1"})
	
	output := r.RenderPrometheus()
	
	// Labels should be sorted alphabetically
	assert.Contains(t, output, `test_metric{a="1",b="2"} 1`)
}

func TestRenderPrometheus_EmptyLabels(t *testing.T) {
	r := NewRegistry()
	
	r.IncLabeled("test_metric", map[string]string{})
	
	output := r.RenderPrometheus()
	
	// Metric without labels should not have braces
	assert.Contains(t, output, "test_metric 1")
}

func TestRegistry_Reset(t *testing.T) {
	r := NewRegistry()
	
	r.RequestsTotal.Add(100)
	r.IncLabeled("test_metric", map[string]string{"label": "value"})
	
	r.Reset()
	
	assert.Equal(t, uint64(0), r.RequestsTotal.Get())
	
	output := r.RenderPrometheus()
	assert.NotContains(t, output, "test_metric")
}
