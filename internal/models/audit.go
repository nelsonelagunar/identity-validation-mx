package models

import (
	"time"

	"gorm.io/gorm"
)

type AuditTrail struct {
	AuditID    uint           `gorm:"primaryKey" json:"audit_id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	Action     string         `gorm:"size:100;not null;index" json:"action" validate:"required"`
	Timestamp  time.Time      `gorm:"not null;index" json:"timestamp"`
	Request    string         `gorm:"type:text" json:"request"`
	Response   string         `gorm:"type:text" json:"response"`
	Status     string         `gorm:"size:20;not null" json:"status" validate:"required,oneof=success failed pending"`
	IPAddress  string         `gorm:"size:45" json:"ip_address"`
	UserAgent  string         `gorm:"size:500" json:"user_agent"`
	Module     string         `gorm:"size:50;index" json:"module"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuditTrail) TableName() string {
	return "audit_trail"
}

func NewAuditTrail(userID uint, action, status string, request, response string) *AuditTrail {
	return &AuditTrail{
		UserID:    userID,
		Action:    action,
		Status:    status,
		Request:   request,
		Response:  response,
		Timestamp: time.Now(),
	}
}

func (a *AuditTrail) WithIP(ip string) *AuditTrail {
	a.IPAddress = ip
	return a
}

func (a *AuditTrail) WithUserAgent(ua string) *AuditTrail {
	a.UserAgent = ua
	return a
}

func (a *AuditTrail) WithModule(module string) *AuditTrail {
	a.Module = module
	return a
}