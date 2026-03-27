package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type APIKey struct {
	Key        string
	Tier       string
	RateLimit  int
	IsValid    bool
}

type AuthMiddleware struct {
	apiKeys map[string]*APIKey
}

func NewAuthMiddleware(apiKeys map[string]*APIKey) *AuthMiddleware {
	return &AuthMiddleware{
		apiKeys: apiKeys,
	}
}

func (a *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing API key",
			})
		}

		key, exists := a.apiKeys[apiKey]
		if !exists || !key.IsValid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid API key",
			})
		}

		c.Locals("apiKey", key)
		return c.Next()
	}
}

func (a *AuthMiddleware) LoadKeysFromDatabase(keys []APIKey) {
	for _, key := range keys {
		a.apiKeys[key.Key] = &APIKey{
			Key:       key.Key,
			Tier:      key.Tier,
			RateLimit: key.RateLimit,
			IsValid:   key.IsValid,
		}
	}
}

func (a *AuthMiddleware) GetAPIKey(c *fiber.Ctx) *APIKey {
	key, ok := c.Locals("apiKey").(*APIKey)
	if !ok {
		return nil
	}
	return key
}

func (a *AuthMiddleware) RequireTier(tiers ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := a.GetAPIKey(c)
		if key == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		for _, tier := range tiers {
			if strings.EqualFold(key.Tier, tier) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions",
		})
	}
}