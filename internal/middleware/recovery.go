package middleware

import (
	"fmt"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
)

type RecoveryConfig struct {
	Logger *zerolog.Logger
}

func NewRecoveryMiddleware() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			fmt.Printf("Panic recovered: %v\nStack trace:\n%s\n", e, stackTrace)
		},
	})
}

func NewRecoveryMiddlewareWithConfig(config *RecoveryConfig) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			if config.Logger != nil {
				config.Logger.Error().
					Str("path", c.Path()).
					Str("method", c.Method()).
					Interface("error", e).
					Str("stack", stackTrace).
					Msg("panic recovered")
			}
		},
	})
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   message,
		"code":    code,
		"success": false,
	})
}