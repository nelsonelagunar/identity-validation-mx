package api

import (
	"identity-validation-mx/internal/api/handlers"
	"identity-validation-mx/internal/repository"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, db *repository.Database) {
	healthHandler := handlers.NewHealthHandler(db, nil)
	identityHandler := handlers.NewIdentityHandler(db)
	biometricHandler := handlers.NewBiometricHandler(db)
	signatureHandler := handlers.NewSignatureHandler(db)
	bulkHandler := handlers.NewBulkHandler(db)

	app.Get("/health", healthHandler.HealthCheck)
	app.Get("/ready", healthHandler.ReadinessCheck)

	v1 := app.Group("/api/v1")

	identity := v1.Group("/identity")
	identity.Post("/curp/validate", identityHandler.ValidateCURP)
	identity.Post("/rfc/validate", identityHandler.ValidateRFC)
	identity.Post("/ine/validate", identityHandler.ValidateINE)
	identity.Get("/curp/:id", identityHandler.GetCURPValidation)

	biometric := v1.Group("/biometric")
	biometric.Post("/compare", biometricHandler.CompareFaces)
	biometric.Post("/liveness", biometricHandler.LivenessDetection)

	signature := v1.Group("/signature")
	signature.Post("/sign", signatureHandler.SignDocument)
	signature.Post("/verify", signatureHandler.VerifySignature)

	importGroup := v1.Group("/import")
	importGroup.Post("/bulk", bulkHandler.ImportBulk)
	importGroup.Get("/:id/status", bulkHandler.GetImportStatus)
}