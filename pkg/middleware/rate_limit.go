package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter IP başına rate limiter ve son görülme zamanını tutar
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter IP bazlı rate limiting yöneticisi
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit
	burst    int
}

// NewRateLimiter belirtilen rate ve burst değerleriyle yeni bir RateLimiter oluşturur.
// r: saniyedeki token üretim hızı (örn. 5 istek/dakika = 5.0/60.0)
// burst: aynı anda izin verilen maksimum istek sayısı
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		r:        r,
		burst:    burst,
	}
	// Uzun süredir görülmeyen IP'lerin limiter'larını temizle
	go rl.cleanupLoop()
	return rl
}

// getLimiter verilen IP için mevcut limiter'ı döndürür, yoksa yeni oluşturur
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ipl, exists := rl.limiters[ip]
	if !exists {
		ipl = &ipLimiter{
			limiter:  rate.NewLimiter(rl.r, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = ipl
	}

	ipl.lastSeen = time.Now()
	return ipl.limiter
}

// cleanupLoop 10 dakikadan uzun süre görülmeyen IP limiter'larını periyodik olarak siler
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, ipl := range rl.limiters {
			if time.Since(ipl.lastSeen) > 10*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware Gin middleware olarak rate limiting uygular.
// Limit aşıldığında 429 Too Many Requests döndürür ve isteği durdurur.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Çok fazla istek gönderildi. Lütfen bir süre bekleyip tekrar deneyin.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Önceden tanımlanmış rate limiter örnekleri:
// LoginRateLimiter    → dakikada 5 istek, burst 5  (brute force koruması)
// RegisterRateLimiter → dakikada 3 istek, burst 3
var (
	LoginRateLimiter    = NewRateLimiter(rate.Limit(5.0/60.0), 5)
	RegisterRateLimiter = NewRateLimiter(rate.Limit(3.0/60.0), 3)
)
