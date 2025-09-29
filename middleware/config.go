package middleware

import "time"

type RateLimitConfig struct {
	Limit  int
	Window time.Duration
}

var (
	// Different rate limits for different endpoints
	DefaultRateLimit = RateLimitConfig{
		Limit:  100,
		Window: time.Minute,
	}

	WriteRateLimit = RateLimitConfig{
		Limit:  50,
		Window: time.Minute,
	}

	ReadRateLimit = RateLimitConfig{
		Limit:  200,
		Window: time.Minute,
	}
)
