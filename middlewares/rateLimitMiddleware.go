package middlewares

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/luponetn/hng-stage-1/utils"
)

type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (rl *rateLimiter) isAllowed(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-window)

	// Filter out old requests
	var validRequests []time.Time
	for _, t := range rl.requests[key] {
		if t.After(threshold) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= limit {
		rl.requests[key] = validRequests
		return false
	}

	rl.requests[key] = append(validRequests, now)
	return true
}

var authLimiter = newRateLimiter()
var apiLimiter = newRateLimiter()

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		// Determine key (User ID if authenticated, else IP)
		key := r.RemoteAddr
		claims, ok := GetUserClaims(r.Context())
		if ok {
			key = claims.UserID
		}

		if strings.HasPrefix(path, "/auth/") {
			if !authLimiter.isAllowed(key, 10, time.Minute) {
				utils.JSONResponse(w, http.StatusTooManyRequests, map[string]string{
					"status":  "error",
					"message": "Too many requests. Auth limit is 10 per minute.",
				})
				return
			}
		} else {
			if !apiLimiter.isAllowed(key, 60, time.Minute) {
				utils.JSONResponse(w, http.StatusTooManyRequests, map[string]string{
					"status":  "error",
					"message": "Too many requests. API limit is 60 per minute.",
				})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
