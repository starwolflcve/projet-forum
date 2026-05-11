package middleware

import (
	"net/http"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	clients = make(map[string][]time.Time)
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		ip := r.RemoteAddr
		now := time.Now()

		var validRequests []time.Time
		for _, t := range clients[ip] {
			if now.Sub(t) < time.Minute {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) >= 50 {
			http.Error(w, "Trop de requêtes", http.StatusTooManyRequests)
			return
		}

		clients[ip] = append(validRequests, now)
		next.ServeHTTP(w, r)
	})
}
