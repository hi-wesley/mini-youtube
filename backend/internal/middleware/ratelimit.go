package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

var (
	limiter *redis_rate.Limiter
	rdb     *redis.Client
)

// InitRateLimiter initializes the Redis-based rate limiter
func InitRateLimiter(redisURL string, redisDB int) error {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse redis URL: %w", err)
	}
	opt.DB = redisDB

	rdb = redis.NewClient(opt)
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	limiter = redis_rate.NewLimiter(rdb)
	return nil
}

// RateLimitByIP creates a rate limiting middleware based on client IP
func RateLimitByIP(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("ip:%s:%s", clientIP, c.Request.URL.Path)

		ctx := c.Request.Context()
		res, err := limiter.Allow(ctx, key, redis_rate.Limit{
			Rate:   limit,
			Period: window,
			Burst:  limit,
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiting error",
			})
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(res.ResetAfter).Unix(), 10))

		if res.Allowed == 0 {
			retryAfter := res.RetryAfter.Seconds()
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
				"retry_after": int(retryAfter),
			})
			return
		}

		c.Next()
	}
}

// RateLimitByUser creates a rate limiting middleware based on authenticated user
func RateLimitByUser(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		uid, exists := c.Get("uid")
		if !exists {
			// Fall back to IP-based limiting if user is not authenticated
			RateLimitByIP(limit, window)(c)
			return
		}

		key := fmt.Sprintf("user:%s:%s", uid, c.Request.URL.Path)

		ctx := c.Request.Context()
		res, err := limiter.Allow(ctx, key, redis_rate.Limit{
			Rate:   limit,
			Period: window,
			Burst:  limit,
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiting error",
			})
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(res.ResetAfter).Unix(), 10))

		if res.Allowed == 0 {
			retryAfter := res.RetryAfter.Seconds()
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
				"retry_after": int(retryAfter),
			})
			return
		}

		c.Next()
	}
}

// RateLimitHybrid combines IP and user-based rate limiting
func RateLimitHybrid(ipLimit int, userLimit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		// Check IP limit first
		clientIP := c.ClientIP()
		ipKey := fmt.Sprintf("ip:%s:%s", clientIP, c.Request.URL.Path)

		ctx := c.Request.Context()
		ipRes, err := limiter.Allow(ctx, ipKey, redis_rate.Limit{
			Rate:   ipLimit,
			Period: window,
			Burst:  ipLimit,
		})

		if err != nil || ipRes.Allowed == 0 {
			handleRateLimitExceeded(c, ipRes)
			return
		}

		// If authenticated, also check user limit
		if uid, exists := c.Get("uid"); exists {
			userKey := fmt.Sprintf("user:%s:%s", uid, c.Request.URL.Path)
			userRes, err := limiter.Allow(ctx, userKey, redis_rate.Limit{
				Rate:   userLimit,
				Period: window,
				Burst:  userLimit,
			})

			if err != nil || userRes.Allowed == 0 {
				handleRateLimitExceeded(c, userRes)
				return
			}

			// Use the more restrictive limit for headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(userLimit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(userRes.Remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(userRes.ResetAfter).Unix(), 10))
		} else {
			// Not authenticated, use IP limits for headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(ipLimit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(ipRes.Remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ipRes.ResetAfter).Unix(), 10))
		}

		c.Next()
	}
}

func handleRateLimitExceeded(c *gin.Context, res *redis_rate.Result) {
	if res != nil {
		retryAfter := res.RetryAfter.Seconds()
		c.Header("Retry-After", strconv.Itoa(int(retryAfter)))
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": "Too many requests. Please try again later.",
			"retry_after": int(retryAfter),
		})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Rate limiting error",
		})
	}
}

// WebSocket connection limiting functions

// CheckWebSocketLimit checks if a user can establish a WebSocket connection for a video
func CheckWebSocketLimit(userID, videoID string) (bool, error) {
	if rdb == nil {
		return true, nil // Rate limiting disabled
	}

	ctx := context.Background()
	key := fmt.Sprintf("ws:user:%s:video:%s", userID, videoID)

	// Check if connection already exists
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if exists > 0 {
		return false, nil // Connection already exists
	}

	// Allow connection and set key with 1 hour expiry
	err = rdb.Set(ctx, key, "1", time.Hour).Err()
	return err == nil, err
}

// ReleaseWebSocketLimit releases a WebSocket connection limit
func ReleaseWebSocketLimit(userID, videoID string) error {
	if rdb == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("ws:user:%s:video:%s", userID, videoID)
	return rdb.Del(ctx, key).Err()
}

// RefreshWebSocketLimit refreshes the TTL for an active WebSocket connection
func RefreshWebSocketLimit(userID, videoID string) error {
	if rdb == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("ws:user:%s:video:%s", userID, videoID)
	return rdb.Expire(ctx, key, time.Hour).Err()
}