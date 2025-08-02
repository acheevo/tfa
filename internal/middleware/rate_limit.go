package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/acheevo/tfa/internal/auth/domain"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	logger          *slog.Logger
	visitors        map[string]*visitor
	mu              sync.RWMutex
	rate            int           // requests per window
	window          time.Duration // time window
	cleanupInterval time.Duration // cleanup interval
}

type visitor struct {
	count     int
	lastSeen  time.Time
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(logger *slog.Logger, rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		logger:          logger,
		visitors:        make(map[string]*visitor),
		rate:            rate,
		window:          window,
		cleanupInterval: time.Minute * 5, // cleanup every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// AuthRateLimit creates a rate limiter middleware for authentication endpoints
func (rl *RateLimiter) AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.getKey(c)

		if !rl.allow(key) {
			rl.logger.Warn("rate limit exceeded", "ip", c.ClientIP(), "key", key)
			c.JSON(http.StatusTooManyRequests, domain.ErrorResponse{
				Error: "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit creates a specific rate limiter for login attempts
func (rl *RateLimiter) LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply IP-based rate limiting for login attempts
		// We don't parse the JSON here to avoid consuming the request body
		ipKey := fmt.Sprintf("login:%s", c.ClientIP())
		if !rl.allow(ipKey) {
			rl.logger.Warn("login rate limit exceeded by IP", "ip", c.ClientIP())
			c.JSON(http.StatusTooManyRequests, domain.ErrorResponse{
				Error: "too many login attempts, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PasswordResetRateLimit creates a rate limiter for password reset requests
func (rl *RateLimiter) PasswordResetRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply IP-based rate limiting for password reset requests
		// We don't parse the JSON here to avoid consuming the request body
		ipKey := fmt.Sprintf("password_reset:%s", c.ClientIP())
		if !rl.allow(ipKey) {
			rl.logger.Warn("password reset rate limit exceeded by IP", "ip", c.ClientIP())
			c.JSON(http.StatusTooManyRequests, domain.ErrorResponse{
				Error: "too many password reset requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if a request is allowed based on the rate limit
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists {
		rl.visitors[key] = &visitor{
			count:     1,
			lastSeen:  now,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	// Reset count if window has passed
	if now.After(v.resetTime) {
		v.count = 1
		v.resetTime = now.Add(rl.window)
		v.lastSeen = now
		return true
	}

	// Check if rate limit exceeded
	if v.count >= rl.rate {
		v.lastSeen = now
		return false
	}

	// Increment count and allow
	v.count++
	v.lastSeen = now
	return true
}

// getKey generates a key for the rate limiter based on IP
func (rl *RateLimiter) getKey(c *gin.Context) string {
	return fmt.Sprintf("auth:%s", c.ClientIP())
}

// cleanupRoutine removes old entries from the visitors map
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanupExpiredEntries()
	}
}

// cleanupExpiredEntries removes expired entries from the visitors map
func (rl *RateLimiter) cleanupExpiredEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window * 2) // Keep entries for 2 windows

	for key, v := range rl.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(rl.visitors, key)
		}
	}
}

// GetRemainingRequests returns the number of remaining requests for a key
func (rl *RateLimiter) GetRemainingRequests(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	v, exists := rl.visitors[key]
	if !exists {
		return rl.rate
	}

	// If window has passed, return full rate
	if time.Now().After(v.resetTime) {
		return rl.rate
	}

	remaining := rl.rate - v.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetResetTime returns when the rate limit will reset for a key
func (rl *RateLimiter) GetResetTime(key string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	v, exists := rl.visitors[key]
	if !exists {
		return time.Now()
	}

	return v.resetTime
}
