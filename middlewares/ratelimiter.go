package middlewares

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientBucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu        sync.Mutex
	clients   map[string]*clientBucket
	r         rate.Limit
	burst     int
	ttl       time.Duration
	allowlist map[string]struct{}
}

func NewRateLimiterFromEnv() *RateLimiter {
	// default: 5 req/sec, burst 10, ttl 5 menit
	rps := getenvFloat("RATE_LIMIT_RPS", 5)
	burst := getenvInt("RATE_LIMIT_BURST", 10)
	ttlSec := getenvInt("RATE_LIMIT_TTL_SECONDS", 300)

	allow := parseAllowlist(os.Getenv("RATE_LIMIT_ALLOWLIST"))

	rl := &RateLimiter{
		clients:   make(map[string]*clientBucket),
		r:         rate.Limit(rps),
		burst:     burst,
		ttl:       time.Duration(ttlSec) * time.Second,
		allowlist: allow,
	}

	// cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	enabled := strings.ToLower(strings.TrimSpace(os.Getenv("RATE_LIMIT_ENABLED")))
	if enabled == "false" || enabled == "0" || enabled == "no" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Jangan limit preflight CORS
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// (opsional) lewatin static
		if strings.HasPrefix(r.URL.Path, "/public/") || strings.HasPrefix(r.URL.Path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}

		ip := getClientIP(r)
		if ip == "" {
			ip = "unknown"
		}

		if _, ok := rl.allowlist[ip]; ok {
			next.ServeHTTP(w, r)
			return
		}

		lim := rl.getLimiter(ip)
		if !lim.Allow() {
			retryAfter := "1" // sederhana; bisa kamu tweak kalau mau lebih presisi
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", retryAfter)
			w.WriteHeader(http.StatusTooManyRequests)

			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  429,
				"error":   true,
				"message": "Too many requests, please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if c, ok := rl.clients[ip]; ok {
		c.lastSeen = time.Now()
		return c.limiter
	}

	lim := rate.NewLimiter(rl.r, rl.burst)
	rl.clients[ip] = &clientBucket{
		limiter:  lim,
		lastSeen: time.Now(),
	}
	return lim
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, c := range rl.clients {
		if now.Sub(c.lastSeen) > rl.ttl {
			delete(rl.clients, ip)
		}
	}
}

func parseAllowlist(raw string) map[string]struct{} {
	m := make(map[string]struct{})
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return m
	}
	for _, s := range strings.Split(raw, ",") {
		ip := strings.TrimSpace(s)
		if ip == "" {
			continue
		}
		m[ip] = struct{}{}
	}
	return m
}

func getClientIP(r *http.Request) string {
	// X-Forwarded-For: "client, proxy1, proxy2"
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if parsed := net.ParseIP(ip); parsed != nil {
			return parsed.String()
		}
	}

	xrip := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if xrip != "" {
		if parsed := net.ParseIP(xrip); parsed != nil {
			return parsed.String()
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		if parsed := net.ParseIP(host); parsed != nil {
			return parsed.String()
		}
		return host
	}

	// fallback (jarang)
	if parsed := net.ParseIP(r.RemoteAddr); parsed != nil {
		return parsed.String()
	}

	return ""
}

func getenvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvFloat(key string, def float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}
