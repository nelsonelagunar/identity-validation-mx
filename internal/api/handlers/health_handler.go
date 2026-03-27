package handlers

import (
	"time"

	"identity-validation-mx/internal/api/dto"
	"identity-validation-mx/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	db    *repository.Database
	redis any
}

func NewHealthHandler(db *repository.Database, redis any) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(dto.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	})
}

func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	checks := make(map[string]dto.Check)

	dbStatus := "ok"
	dbLatency := ""
	dbError := ""
	if h.db != nil {
		start := time.Now()
		if err := h.db.HealthCheck(); err != nil {
			dbStatus = "error"
			dbError = err.Error()
		} else {
			dbLatency = time.Since(start).String()
		}
	} else {
		dbStatus = "not_configured"
	}
	checks["database"] = dto.Check{
		Status:  dbStatus,
		Latency: dbLatency,
		Error:   dbError,
	}

	redisStatus := "ok"
	redisLatency := ""
	redisError := ""
	if h.redis != nil {
		start := time.Now()
		if err := checkRedisHealth(h.redis); err != nil {
			redisStatus = "error"
			redisError = err.Error()
		} else {
			redisLatency = time.Since(start).String()
		}
	} else {
		redisStatus = "not_configured"
	}
	checks["redis"] = dto.Check{
		Status:  redisStatus,
		Latency: redisLatency,
		Error:   redisError,
	}

	allHealthy := true
	for _, check := range checks {
		if check.Status == "error" {
			allHealthy = false
			break
		}
	}

	status := "ready"
	if !allHealthy {
		status = "not_ready"
	}

	return c.Status(fiber.StatusOK).JSON(dto.ReadinessResponse{
		Status: status,
		Checks: checks,
	})
}

func checkRedisHealth(redis any) error {
	return nil
}