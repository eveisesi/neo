package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/vektah/gqlparser/gqlerror"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mtx      sync.Mutex
	visitors map[string]*visitor
)

func (s *Server) RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		limiter := getVisitor(ip)
		if !limiter.Allow() {

			b, _ := json.Marshal(gqlerror.Error{
				Message: "Too Many Requests",
			})

			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write(b)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func addVisitor(ip string) *rate.Limiter {
	limiter := rate.NewLimiter(5, 10)
	mtx.Lock()
	visitors[ip] = &visitor{limiter, time.Now()}
	mtx.Unlock()
	return limiter
}

func getVisitor(ip string) *rate.Limiter {
	mtx.Lock()
	v, exists := visitors[ip]
	if !exists {
		mtx.Unlock()
		return addVisitor(ip)
	}

	v.lastSeen = time.Now()
	mtx.Unlock()
	return v.limiter
}

func cleanUpVisitors() {
	for {
		time.Sleep(time.Minute)
		mtx.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mtx.Unlock()
	}
}
