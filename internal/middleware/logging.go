package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type LoggingConfig struct {
	Logger   *zerolog.Logger
	SkipPaths []string
}

func NewLoggingMiddleware(logger *zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		event := logger.Info()
		if status >= 400 {
			event = logger.Warn()
		}
		if status >= 500 {
			event = logger.Error()
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("duration", duration).
			Str("ip", c.IP()).
			Str("user-agent", c.Get("User-Agent")).
			Int64("bytes_out", int64(len(c.Response().Body()))).
			Msg("request")

		return err
	}
}

func NewLoggingMiddlewareWithConfig(config *LoggingConfig) fiber.Handler {
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()
		if skipMap[path] {
			return c.Next()
		}

		start := time.Now()
		method := c.Method()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		apiKey := GetAPIKeyFromContext(c)
		var keyStr string
		if apiKey != nil {
			keyStr = apiKey.Key
		}

		event := config.Logger.Info()
		if status >= 400 {
			event = config.Logger.Warn()
		}
		if status >= 500 {
			event = config.Logger.Error()
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("duration", duration).
			Str("ip", c.IP()).
			Str("api_key", keyStr).
			Str("request_id", c.Get("X-Request-ID")).
			Msg("request")

		return err
	}
}

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("requestId", requestID)
		return c.Next()
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}