package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
	KeyExtractor      func(c *fiber.Ctx) string
}

type RateLimiter struct {
	redis      *redis.Client
	configs    map[string]*RateLimitConfig
	defaultTTL time.Duration
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis:      redisClient,
		configs:    make(map[string]*RateLimitConfig),
		defaultTTL: time.Minute,
	}
}

func (r *RateLimiter) AddConfig(tier string, config *RateLimitConfig) {
	r.configs[tier] = config
}

func (r *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := GetAPIKeyFromContext(c)
		if apiKey == nil {
			key := c.IP()
			return r.checkRateLimit(c, key, "default")
		}

		tierConfig, exists := r.configs[apiKey.Tier]
		if !exists {
			tierConfig = r.configs["default"]
		}

		return r.checkRateLimit(c, apiKey.Key, apiKey.Tier)
	}
}

func (r *RateLimiter) checkRateLimit(c *fiber.Ctx, key string, tier string) error {
	ctx := context.Background()
	config, exists := r.configs[tier]
	if !exists {
		config = &RateLimitConfig{
			RequestsPerMinute: 60,
			Burst:             10,
		}
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	now := time.Now()
	windowStart := now.Truncate(time.Minute)

	count, err := r.redis.Get(ctx, redisKey).Int()
	if err != nil && err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "rate limit check failed",
		})
	}

	if count >= config.RequestsPerMinute {
		retryAfter := time.Minute - now.Sub(windowStart)
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
		c.Set("X-RateLimit-Remaining", "0")
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(retryAfter).Unix()))
		c.Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))

		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "rate limit exceeded",
		})
	}

	pipe := r.redis.Pipeline()
	pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, r.defaultTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "rate limit check failed",
		})
	}

	remaining := config.RequestsPerMinute - count - 1
	c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
	c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

	return c.Next()
}

func (r *RateLimiter) MiddlewareWithConfig(config *RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := config.KeyExtractor(c)
		return r.checkRateLimit(c, key, "custom")
	}
}

func GetAPIKeyFromContext(c *fiber.Ctx) *APIKey {
	key, ok := c.Locals("apiKey").(*APIKey)
	if !ok {
		return nil
	}
	return key
}

func DefaultKeyExtractor(c *fiber.Ctx) string {
	apiKey := GetAPIKeyFromContext(c)
	if apiKey != nil {
		return apiKey.Key
	}
	return c.IP()
}