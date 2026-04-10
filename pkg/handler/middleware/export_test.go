package middleware

import "time"

// NewRateLimiterForTest exposes newRateLimiter for use in external test packages.
func NewRateLimiterForTest(limit int, window time.Duration) *rateLimiter {
	return newRateLimiter(limit, window)
}
