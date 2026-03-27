package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type WebhookStore struct {
	webhooks   map[string]*models.Webhook
	deliveries map[string][]models.WebhookDelivery
	mu         sync.RWMutex
}

func NewWebhookStore() *WebhookStore {
	return &WebhookStore{
		webhooks:   make(map[string]*models.Webhook),
		deliveries: make(map[string][]models.WebhookDelivery),
	}
}

func (s *WebhookStore) Create(ctx context.Context, webhook *models.Webhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.webhooks[webhook.ID] = webhook
	return nil
}

func (s *WebhookStore) GetByID(ctx context.Context, id string) (*models.Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	webhook, exists := s.webhooks[id]
	if !exists {
		return nil, fmt.Errorf("webhook not found: %s", id)
	}

	return webhook, nil
}

func (s *WebhookStore) ListByUser(ctx context.Context, userID string) ([]models.Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var webhooks []models.Webhook
	for _, webhook := range s.webhooks {
		if webhook.UserID == userID {
			webhooks = append(webhooks, *webhook)
		}
	}

	return webhooks, nil
}

func (s *WebhookStore) GetActiveByEvent(ctx context.Context, event string) ([]*models.Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var webhooks []*models.Webhook
	for _, webhook := range s.webhooks {
		if !webhook.Active {
			continue
		}

		var events []string
		if err := json.Unmarshal([]byte(webhook.Events), &events); err != nil {
			continue
		}

		for _, e := range events {
			if e == event || e == "*" {
				webhooks = append(webhooks, webhook)
				break
			}
		}
	}

	return webhooks, nil
}

func (s *WebhookStore) Deactivate(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	webhook, exists := s.webhooks[id]
	if !exists {
		return fmt.Errorf("webhook not found: %s", id)
	}

	webhook.Active = false
	return nil
}

func (s *WebhookStore) Update(ctx context.Context, webhook *models.Webhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.webhooks[webhook.ID] = webhook
	return nil
}

func (s *WebhookStore) RecordDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deliveries[delivery.WebhookID] = append(s.deliveries[delivery.WebhookID], *delivery)
	return nil
}

func (s *WebhookStore) GetDeliveryHistory(ctx context.Context, webhookID string, limit int) ([]models.WebhookDelivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	deliveries, exists := s.deliveries[webhookID]
	if !exists {
		return nil, fmt.Errorf("no deliveries found for webhook: %s", webhookID)
	}

	if limit > 0 && len(deliveries) > limit {
		return deliveries[:limit], nil
	}

	return deliveries, nil
}
