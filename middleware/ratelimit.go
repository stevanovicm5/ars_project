package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.Mutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		// Clean old requests
		now := time.Now()
		var validRequests []time.Time
		for _, t := range rl.requests[clientIP] {
			if now.Sub(t) <= rl.window {
				validRequests = append(validRequests, t)
			}
		}

		// Check if over limit
		if len(validRequests) >= rl.limit {
			resetTime := now.Add(rl.window)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", resetTime.Format(time.RFC1123))
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window/time.Second)))

			http.Error(w, "Rate limit exceeded. Too many requests.", http.StatusTooManyRequests)
			return
		}

		// Add current request
		validRequests = append(validRequests, now)
		rl.requests[clientIP] = validRequests

		// Set rate limit headers
		remaining := rl.limit - len(validRequests)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", now.Add(rl.window).Format(time.RFC1123))

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	// Check for forwarded IP (behind proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check for real IP
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to remote address
	return r.RemoteAddr
}
