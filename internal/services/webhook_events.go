package services

import (
	"context"

	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type WebhookEventBus struct {
	service WebhookService
}

func NewWebhookEventBus(service WebhookService) *WebhookEventBus {
	return &WebhookEventBus{
		service: service,
	}
}

func (b *WebhookEventBus) PublishValidationComplete(ctx context.Context, validationID string, result interface{}) error {
	return b.service.Trigger(ctx, models.WebhookEventValidationComplete, map[string]interface{}{
		"validation_id": validationID,
		"result":        result,
	})
}

func (b *WebhookEventBus) PublishImportComplete(ctx context.Context, jobID string, summary map[string]interface{}) error {
	return b.service.Trigger(ctx, models.WebhookEventImportComplete, map[string]interface{}{
		"job_id":  jobID,
		"summary": summary,
	})
}

func (b *WebhookEventBus) PublishSignatureComplete(ctx context.Context, signatureID string, documentURL string) error {
	return b.service.Trigger(ctx, models.WebhookEventSignatureComplete, map[string]interface{}{
		"signature_id": signatureID,
		"document_url": documentURL,
	})
}

func (b *WebhookEventBus) PublishBiometricComplete(ctx context.Context, biometricID string, score float64) error {
	return b.service.Trigger(ctx, models.WebhookEventBiometricComplete, map[string]interface{}{
		"biometric_id": biometricID,
		"score":         score,
	})
}

func (b *WebhookEventBus) PublishError(ctx context.Context, errorType string, message string, details interface{}) error {
	return b.service.Trigger(ctx, models.WebhookEventError, map[string]interface{}{
		"error_type": errorType,
		"message":    message,
		"details":    details,
	})
}
