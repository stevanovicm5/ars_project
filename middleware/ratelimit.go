package middleware

import (
	"log"
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
		log.Println("RATE LIMITER EXECUTED for:", r.URL.Path)
		clientIP := getClientIP(r)

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		now := time.Now()
		var validRequests []time.Time
		for _, t := range rl.requests[clientIP] {
			if now.Sub(t) <= rl.window {
				validRequests = append(validRequests, t)
			}
		}

		var resetTime int64
		if len(validRequests) > 0 {

			resetTime = now.Add(rl.window).Unix()
		} else {
			resetTime = now.Add(rl.window).Unix()
		}

		if len(validRequests) >= rl.limit {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			retryAfter := resetTime - time.Now().Unix()
			if retryAfter < 0 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))

			http.Error(w, "Rate limit exceeded. Too many requests.", http.StatusTooManyRequests)
			return
		}

		validRequests = append(validRequests, now)
		rl.requests[clientIP] = validRequests

		remaining := rl.limit - len(validRequests)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

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
