package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type CURPValidationRequest struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CURP      string         `gorm:"size:18;uniqueIndex;not null" json:"curp" validate:"required,len=18,curp_format"`
	UserID    uint           `gorm:"not null;index" json:"user_id" validate:"required"`
	Status    string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CURPValidationRequest) TableName() string {
	return "curp_validation_requests"
}

type CURPValidationResponse struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	RequestID           uint           `gorm:"not null;uniqueIndex" json:"request_id" validate:"required"`
	IsValid             bool           `json:"is_valid"`
	FullName            string         `gorm:"size:200" json:"full_name" validate:"omitempty"`
	BirthDate           *time.Time     `json:"birth_date"`
	Gender              string         `gorm:"size:1" json:"gender" validate:"omitempty,oneof=M F"`
	BirthState          string         `gorm:"size:50" json:"birth_state"`
	ValidationError     string         `gorm:"size:500" json:"validation_error"`
	RenapoVerified      bool           `json:"renapo_verified"`
	VerificationScore   float64        `json:"verification_score"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CURPValidationResponse) TableName() string {
	return "curp_validation_responses"
}

type RFCValidationRequest struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	RFC       string         `gorm:"size:13;uniqueIndex;not null" json:"rfc" validate:"required,len=12|len=13,rfc_format"`
	UserID    uint           `gorm:"not null;index" json:"user_id" validate:"required"`
	Status    string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RFCValidationRequest) TableName() string {
	return "rfc_validation_requests"
}

type RFCValidationResponse struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	RequestID         uint           `gorm:"not null;uniqueIndex" json:"request_id" validate:"required"`
	IsValid           bool           `json:"is_valid"`
	FullName          string         `gorm:"size:200" json:"full_name"`
	TaxRegime         string         `gorm:"size:50" json:"tax_regime"`
	RegistrationDate  *time.Time     `json:"registration_date"`
	StatusSAT         string         `gorm:"size:50" json:"status_sat"`
	ValidationError   string         `gorm:"size:500" json:"validation_error"`
	SatVerified       bool           `json:"sat_verified"`
	VerificationScore float64        `json:"verification_score"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RFCValidationResponse) TableName() string {
	return "rfc_validation_responses"
}

type INEValidationRequest struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	INEClave       string         `gorm:"size:18;uniqueIndex;not null" json:"ine_clave" validate:"required,len=18"`
	UserID         uint           `gorm:"not null;index" json:"user_id" validate:"required"`
	OCRNumber      string         `gorm:"size:13" json:"ocr_number" validate:"omitempty,len=13"`
	ElectionKey    string         `gorm:"size:18" json:"election_key" validate:"omitempty,len=18"`
	Status         string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (INEValidationRequest) TableName() string {
	return "ine_validation_requests"
}

type INEValidationResponse struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	RequestID           uint           `gorm:"not null;uniqueIndex" json:"request_id" validate:"required"`
	IsValid             bool           `json:"is_valid"`
	FullName            string         `gorm:"size:200" json:"full_name"`
	BirthDate           *time.Time     `json:"birth_date"`
	Gender              string         `gorm:"size:1" json:"gender" validate:"omitempty,oneof=M F"`
	Address             string         `gorm:"size:500" json:"address"`
	VotingSection       string         `gorm:"size:10" json:"voting_section"`
	ValidationError     string         `gorm:"size:500" json:"validation_error"`
	INEVerified         bool           `json:"ine_verified"`
	VerificationScore   float64        `json:"verification_score"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (INEValidationResponse) TableName() string {
	return "ine_validation_responses"
}

func SetupIdentityValidator(v *validator.Validate) {
	v.RegisterValidation("curp_format", validateCURPFormat)
	v.RegisterValidation("rfc_format", validateRFCFormat)
}

func validateCURPFormat(fl validator.FieldLevel) bool {
	curp := fl.Field().String()
	if len(curp) != 18 {
		return false
	}
	return true
}

func validateRFCFormat(fl validator.FieldLevel) bool {
	rfc := fl.Field().String()
	l := len(rfc)
	return l == 12 || l == 13
}