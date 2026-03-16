package handler

import (
	"net/http"
	"sync"
	"time"
)

// ── Security Headers Middleware ───────────────────────────────────────────────

// SecureHeaders adds hardening HTTP response headers to every response.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		// Prevent MIME-type sniffing
		h.Set("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		h.Set("X-Frame-Options", "DENY")
		// XSS filter (legacy browsers)
		h.Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Don't cache API responses
		if len(r.URL.Path) > 4 && r.URL.Path[:4] == "/api" {
			h.Set("Cache-Control", "no-store")
		}
		// Content Security Policy
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://fonts.googleapis.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://fonts.gstatic.com; "+
				"font-src https://fonts.gstatic.com; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self';",
		)
		next.ServeHTTP(w, r)
	})
}

// ── Rate Limiter ──────────────────────────────────────────────────────────────

// RateLimiter is a simple per-IP token-bucket rate limiter.
// Each IP gets `capacity` requests per `window` duration.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity int
	window   time.Duration
}

type bucket struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter creates a limiter allowing `capacity` requests per `window`.
func NewRateLimiter(capacity int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: capacity,
		window:   window,
	}
	// Clean up stale buckets periodically
	go func() {
		for range time.Tick(window * 2) {
			rl.purge()
		}
	}()
	return rl
}

// Middleware wraps a handler with rate limiting by client IP.
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Retry-After", "60")
			jsonError(w, "too many requests — please slow down", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[ip]
	now := time.Now()
	if !ok || now.After(b.resetAt) {
		rl.buckets[ip] = &bucket{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if b.count >= rl.capacity {
		return false
	}
	b.count++
	return true
}

func (rl *RateLimiter) purge() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, b := range rl.buckets {
		if now.After(b.resetAt) {
			delete(rl.buckets, ip)
		}
	}
}

// clientIP extracts the real client IP, respecting X-Forwarded-For for proxies.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first (leftmost) IP — the original client
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
