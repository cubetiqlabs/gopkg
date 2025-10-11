package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/cubetiqlabs/gopkg/metrics"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultMaxBuckets       = 10000             // Prevent memory exhaustion
	bucketCleanupInterval   = 5 * time.Minute   // How often to clean up stale buckets
	bucketInactiveThreshold = 15 * time.Minute  // When to consider a bucket stale
)

// RateLimiter implements a token bucket rate limiter per key.
// It supports:
// - Per-key rate limiting (tenant, API key, IP, etc.)
// - Dynamic burst capacity (half of rate)
// - Automatic bucket cleanup to prevent memory exhaustion
// - Retry-After header for rejected requests
type RateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*bucket
	ratePerMin  int       // Default global rate limit (requests per minute)
	maxBuckets  int       // Max number of buckets to keep in memory
	lastCleanup time.Time // Last time we cleaned up stale buckets
}

// bucket represents a token bucket for a single key.
type bucket struct {
	tokens   float64   // Current token count
	last     time.Time // Last refill time
	accessed time.Time // Last access time (for cleanup)
}

// NewRateLimiter creates a new rate limiter with the specified rate per minute.
//
// Parameters:
//   - ratePerMin: Maximum requests per minute (default: 600 if <= 0)
//
// Example usage:
//
//	limiter := middleware.NewRateLimiter(600) // 600 req/min = 10 req/sec
//	app.Use(middleware.RateLimitMiddleware(limiter, nil))
func NewRateLimiter(ratePerMin int) *RateLimiter {
	if ratePerMin <= 0 {
		ratePerMin = 600
	}
	return &RateLimiter{
		buckets:     make(map[string]*bucket),
		ratePerMin:  ratePerMin,
		maxBuckets:  defaultMaxBuckets,
		lastCleanup: time.Now(),
	}
}

// take attempts to consume one token from the bucket for the given key.
// Returns:
// - allowed: true if request is allowed
// - retryAfter: duration to wait before retrying if rejected
func (rl *RateLimiter) take(key string, rate int) (allowed bool, retryAfter time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Periodic cleanup of inactive buckets
	if now.Sub(rl.lastCleanup) > bucketCleanupInterval {
		rl.cleanupStaleBuckets(now)
		rl.lastCleanup = now
	}

	// Get or create bucket
	b, ok := rl.buckets[key]
	if !ok {
		// Enforce max buckets limit to prevent memory exhaustion DoS
		if len(rl.buckets) >= rl.maxBuckets {
			// Try to evict oldest bucket
			if !rl.evictOldestBucket(now) {
				// Could not evict, reject this request
				return false, time.Minute
			}
		}

		// Create new bucket with initial burst capacity
		dynBurst := rate / 2
		if dynBurst < 1 {
			dynBurst = 1
		}
		b = &bucket{
			tokens:   float64(dynBurst),
			last:     now,
			accessed: now,
		}
		rl.buckets[key] = b
	}

	// Update access time
	b.accessed = now

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.last).Minutes()
	if elapsed > 0 {
		b.tokens += elapsed * float64(rate)
		
		// Cap at burst capacity (half of rate)
		maxTokens := float64(rate / 2)
		if maxTokens < 1 {
			maxTokens = 1
		}
		if b.tokens > maxTokens {
			b.tokens = maxTokens
		}
		b.last = now
	}

	// Try to consume a token
	if b.tokens >= 1 {
		b.tokens -= 1
		return true, 0
	}

	// Not enough tokens - calculate retry time
	deficit := 1 - b.tokens
	minutes := deficit / float64(rate)
	retry := time.Duration(minutes * float64(time.Minute))
	if retry < time.Second {
		retry = time.Second
	}
	
	return false, retry
}

// cleanupStaleBuckets removes buckets that haven't been accessed recently.
// This prevents memory exhaustion from keeping too many buckets.
func (rl *RateLimiter) cleanupStaleBuckets(now time.Time) {
	threshold := now.Add(-bucketInactiveThreshold)
	for key, b := range rl.buckets {
		if b.accessed.Before(threshold) {
			delete(rl.buckets, key)
		}
	}
}

// evictOldestBucket removes the least recently accessed bucket.
// Returns true if eviction succeeded, false if no buckets could be evicted.
func (rl *RateLimiter) evictOldestBucket(now time.Time) bool {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, b := range rl.buckets {
		if first || b.accessed.Before(oldestTime) {
			oldestKey = key
			oldestTime = b.accessed
			first = false
		}
	}

	if oldestKey != "" {
		delete(rl.buckets, oldestKey)
		return true
	}
	return false
}

// RateLimitConfig defines configuration for rate limit middleware.
type RateLimitConfig struct {
	// KeyGenerator generates a unique key for rate limiting
	// Default: uses IP address
	KeyGenerator func(c *fiber.Ctx) string
	
	// RateGetter returns the rate limit for a specific request
	// Default: uses the limiter's default rate
	RateGetter func(c *fiber.Ctx) int
}

// RateLimitMiddleware returns a Fiber middleware that enforces rate limits.
//
// Parameters:
//   - limiter: The rate limiter instance
//   - reg: Optional metrics registry for tracking allowed/rejected requests
//
// Example usage:
//
//	limiter := middleware.NewRateLimiter(600)
//	reg := metrics.NewRegistry()
//	app.Use(middleware.RateLimitMiddleware(limiter, reg))
func RateLimitMiddleware(limiter *RateLimiter, reg *metrics.Registry) fiber.Handler {
	return RateLimitMiddlewareWithConfig(limiter, reg, RateLimitConfig{})
}

// RateLimitMiddlewareWithConfig returns a rate limit middleware with custom configuration.
//
// Example usage:
//
//	limiter := middleware.NewRateLimiter(600)
//	app.Use(middleware.RateLimitMiddlewareWithConfig(limiter, nil, middleware.RateLimitConfig{
//	    KeyGenerator: func(c *fiber.Ctx) string {
//	        return c.Get("X-API-Key") // Rate limit by API key
//	    },
//	}))
func RateLimitMiddlewareWithConfig(limiter *RateLimiter, reg *metrics.Registry, cfg RateLimitConfig) fiber.Handler {
	// Set defaults
	if cfg.KeyGenerator == nil {
		cfg.KeyGenerator = func(c *fiber.Ctx) string {
			return c.IP() // Default: rate limit by IP
		}
	}
	if cfg.RateGetter == nil {
		cfg.RateGetter = func(c *fiber.Ctx) int {
			return limiter.ratePerMin
		}
	}

	return func(c *fiber.Ctx) error {
		// Generate rate limit key
		key := cfg.KeyGenerator(c)
		if key == "" {
			key = "anonymous"
		}

		// Get rate for this request
		rate := cfg.RateGetter(c)

		// Check rate limit
		allowed, retryAfter := limiter.take(key, rate)
		
		if !allowed {
			// Record rejection metric
			if reg != nil {
				reg.RateRejected.Inc()
			}
			
			// Set Retry-After header
			c.Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			
			// Return 429 Too Many Requests
			return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
		}

		// Record allowed metric
		if reg != nil {
			reg.RateAllowed.Inc()
		}

		return c.Next()
	}
}
