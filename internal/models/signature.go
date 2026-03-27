package models

import (
	"time"

	"gorm.io/gorm"
)

type DigitalSignatureRequest struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	DocumentHash   string         `gorm:"size:128;not null" json:"document_hash"`
	SignerName     string         `gorm:"size:200;not null" json:"signer_name"`
	SignerRFCCURP  string         `gorm:"size:18" json:"signer_rfc_curp"`
	SignatureType  string         `gorm:"size:20;default:'basic'" json:"signature_type" validate:"omitempty,oneof=basic advanced qualified"`
	ExpiresAt      *time.Time     `json:"expires_at"`
	Status         string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed cancelled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DigitalSignatureRequest) TableName() string {
	return "digital_signature_requests"
}

type DigitalSignatureResponse struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	RequestID        uint           `gorm:"not null;uniqueIndex" json:"request_id"`
	Signature        string         `gorm:"type:text" json:"signature"`
	SerialNumber     string         `gorm:"size:64" json:"serial_number"`
	IssuerDN         string         `gorm:"size:500" json:"issuer_dn"`
	SubjectDN        string         `gorm:"size:500" json:"subject_dn"`
	ValidFrom        *time.Time     `json:"valid_from"`
	ValidTo          *time.Time     `json:"valid_to"`
	SignatureBase64   string         `gorm:"type:text" json:"signature_base64"`
	Certificate      string         `gorm:"type:text" json:"certificate"`
	ProviderResult   string         `gorm:"type:text" json:"provider_result"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DigitalSignatureResponse) TableName() string {
	return "digital_signature_responses"
}

type SignatureVerificationRequest struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	SignatureID    uint           `gorm:"not null;index" json:"signature_id"`
	DocumentHash   string         `gorm:"size:128;not null" json:"document_hash"`
	Signature      string         `gorm:"type:text;not null" json:"signature"`
	Status         string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SignatureVerificationRequest) TableName() string {
	return "signature_verification_requests"
}

type SignatureVerificationResponse struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	RequestID         uint           `gorm:"not null;uniqueIndex" json:"request_id"`
	IsValid           bool           `json:"is_valid"`
	SignerVerified    bool           `json:"signer_verified"`
	DocumentIntegrity bool           `json:"document_integrity"`
	TimestampValid    bool           `json:"timestamp_valid"`
	ErrorCode         string         `gorm:"size:50" json:"error_code"`
	ErrorMessage      string         `gorm:"type:text" json:"error_message"`
	VerificationDetails string        `gorm:"type:text" json:"verification_details"`
	ProcessingTime    int64          `json:"processing_time_ms"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SignatureVerificationResponse) TableName() string {
	return "signature_verification_responses"
}