package models

import "time"

type Webhook struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index"`
	URL       string    `json:"url" gorm:"not null"`
	Secret    string    `json:"secret"`
	Events    string    `json:"events"` // JSON array of event types
	Active    bool      `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebhookEventType string

const (
	WebhookEventValidationComplete WebhookEventType = "validation.complete"
	WebhookEventImportComplete     WebhookEventType = "import.complete"
	WebhookEventSignatureComplete  WebhookEventType = "signature.complete"
	WebhookEventBiometricComplete  WebhookEventType = "biometric.complete"
	WebhookEventError              WebhookEventType = "error"
)

type WebhookPayload struct {
	Event     WebhookEventType `json:"event"`
	Timestamp time.Time       `json:"timestamp"`
	Data      interface{}     `json:"data"`
	Signature string          `json:"signature,omitempty"`
}

type WebhookDelivery struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	WebhookID   string    `json:"webhook_id" gorm:"index"`
	Event       string    `json:"event"`
	Status      string    `json:"status"` // success, failed, retry
	StatusCode  int       `json:"status_code"`
	Response    string    `json:"response"`
	Attempts    int       `json:"attempts"`
	LastAttempt time.Time `json:"last_attempt"`
	CreatedAt   time.Time `json:"created_at"`
}
