package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type WebhookService interface {
	Subscribe(ctx context.Context, userID, url, secret string, events []string) (*models.Webhook, error)
	Unsubscribe(ctx context.Context, webhookID string) error
	Trigger(ctx context.Context, event models.WebhookEventType, data interface{}) error
	List(ctx context.Context, userID string) ([]models.Webhook, error)
}

type webhookService struct {
	store   *WebhookStore
	client  *WebhookClient
	maxRetries int
	retryDelay time.Duration
}

func NewWebhookService(store *WebhookStore, client *WebhookClient) WebhookService {
	return &webhookService{
		store:      store,
		client:     client,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
}

func (s *webhookService) Subscribe(ctx context.Context, userID, url, secret string, events []string) (*models.Webhook, error) {
	if err := s.validateURL(url); err != nil {
		return nil, err
	}

	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	webhook := &models.Webhook{
		ID:        uuid.New().String(),
		UserID:    userID,
		URL:       url,
		Secret:    secret,
		Events:    string(eventsJSON),
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.store.Create(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return webhook, nil
}

func (s *webhookService) Unsubscribe(ctx context.Context, webhookID string) error {
	return s.store.Deactivate(ctx, webhookID)
}

func (s *webhookService) Trigger(ctx context.Context, event models.WebhookEventType, data interface{}) error {
	webhooks, err := s.store.GetActiveByEvent(ctx, string(event))
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	for _, webhook := range webhooks {
		go s.deliverWithRetry(webhook, event, data)
	}

	return nil
}

func (s *webhookService) deliverWithRetry(webhook *models.Webhook, event models.WebhookEventType, data interface{}) {
	payload := models.WebhookPayload{
		Event:     event,
		Timestamp: time.Now(),
		Data:      data,
	}

	payload.Signature = s.generateSignature(payload, webhook.Secret)

	for attempt := 0; attempt < s.maxRetries; attempt++ {
		err := s.client.Deliver(webhook.URL, payload)
		if err == nil {
			s.recordDelivery(webhook.ID, string(event), "success", 200, "")
			return
		}

		if attempt < s.maxRetries-1 {
			time.Sleep(s.retryDelay * time.Duration(attempt+1))
		}

		s.recordDelivery(webhook.ID, string(event), "failed", 0, err.Error())
	}
}

func (s *webhookService) generateSignature(payload models.WebhookPayload, secret string) string {
	data, _ := json.Marshal(payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *webhookService) validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL is required")
	}
	if len(url) < 10 {
		return fmt.Errorf("invalid URL format")
	}
	return nil
}

func (s *webhookService) recordDelivery(webhookID, event, status string, statusCode int, response string) {
	delivery := &models.WebhookDelivery{
		ID:          uuid.New().String(),
		WebhookID:   webhookID,
		Event:       event,
		Status:      status,
		StatusCode:  statusCode,
		Response:    response,
		Attempts:    1,
		LastAttempt: time.Now(),
		CreatedAt:   time.Now(),
	}

	s.store.RecordDelivery(context.Background(), delivery)
}

func (s *webhookService) List(ctx context.Context, userID string) ([]models.Webhook, error) {
	return s.store.ListByUser(ctx, userID)
}
