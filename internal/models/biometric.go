package models

import (
	"time"

	"gorm.io/gorm"
)

type FacialComparisonRequest struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	DocumentPhoto string         `gorm:"type:text" json:"document_photo"`
	SelfiePhoto   string         `gorm:"type:text" json:"selfie_photo"`
	Status        string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FacialComparisonRequest) TableName() string {
	return "facial_comparison_requests"
}

type FacialComparisonResponse struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	RequestID        uint           `gorm:"not null;uniqueIndex" json:"request_id"`
	IsMatch          bool           `json:"is_match"`
	SimilarityScore  float64        `json:"similarity_score"`
	ConfidenceLevel  float64        `json:"confidence_level"`
	DetectedAnomalies string        `gorm:"type:text" json:"detected_anomalies"`
	ProcessingTime   int64          `json:"processing_time_ms"`
	ProviderResult   string         `gorm:"type:text" json:"provider_result"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FacialComparisonResponse) TableName() string {
	return "facial_comparison_responses"
}

type LivenessDetectionRequest struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	VideoFile   string         `gorm:"size:500" json:"video_file"`
	ImageFiles  string         `gorm:"type:text" json:"image_files"`
	Status      string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LivenessDetectionRequest) TableName() string {
	return "liveness_detection_requests"
}

type LivenessDetectionResponse struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	RequestID          uint           `gorm:"not null;uniqueIndex" json:"request_id"`
	IsLive             bool           `json:"is_live"`
	LivenessScore      float64        `json:"liveness_score"`
	ConfidenceLevel    float64        `json:"confidence_level"`
	SpoofProbability   float64        `json:"spoof_probability"`
	DetectedAttacks    string         `gorm:"type:text" json:"detected_attacks"`
	ProcessingTime     int64          `json:"processing_time_ms"`
	ProviderResult     string         `gorm:"type:text" json:"provider_result"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LivenessDetectionResponse) TableName() string {
	return "liveness_detection_responses"
}