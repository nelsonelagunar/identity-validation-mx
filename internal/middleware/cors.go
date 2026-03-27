package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

func NewCORSMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key,X-Request-ID",
		AllowCredentials: false,
		MaxAge:           86400,
	})
}

func NewCORSMiddlewareWithConfig(config *CORSConfig) fiber.Handler {
	origins := "*"
	if len(config.AllowOrigins) > 0 {
		origins = ""
		for i, origin := range config.AllowOrigins {
			if i > 0 {
				origins += ","
			}
			origins += origin
		}
	}

	methods := "GET,POST,PUT,DELETE,OPTIONS,PATCH"
	if len(config.AllowMethods) > 0 {
		methods = ""
		for i, method := range config.AllowMethods {
			if i > 0 {
				methods += ","
			}
			methods += method
		}
	}

	headers := "Origin,Content-Type,Accept,Authorization,X-API-Key,X-Request-ID"
	if len(config.AllowHeaders) > 0 {
		headers = ""
		for i, header := range config.AllowHeaders {
			if i > 0 {
				headers += ","
			}
			headers += header
		}
	}

	exposeHeaders := ""
	if len(config.ExposeHeaders) > 0 {
		for i, header := range config.ExposeHeaders {
			if i > 0 {
				exposeHeaders += ","
			}
			exposeHeaders += header
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     methods,
		AllowHeaders:     headers,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

func DefaultProductionCORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "https://example.com",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key",
		AllowCredentials: true,
		MaxAge:           3600,
	})
}