package server

import (
	"log/slog"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

var rateLimitMap = make(map[string]*rate.Limiter)
var mutexRateLimitMap = sync.Mutex{}

func isRateLimitAllowed(ip string) bool {
	mutexRateLimitMap.Lock()
	defer mutexRateLimitMap.Unlock()

	rl, exists := rateLimitMap[ip]

	if !exists {
		rl = rate.NewLimiter(DefaultConfig.Rate.RequestsPerSecond, DefaultConfig.Rate.Burst)
		rateLimitMap[ip] = rl
	}

	return rl.Allow()
}

func RateLimiterMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, ok := GetRequestIPAddress(r)
			if !ok {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !isRateLimitAllowed(ip) {
				logger.Warn("Rate limit exceeded", "ip", ip)
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
