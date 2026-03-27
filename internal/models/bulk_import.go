package models

import (
	"time"

	"gorm.io/gorm"
)

type BulkImportJob struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	JobID        string         `gorm:"size:64;uniqueIndex;not null" json:"job_id" validate:"required"`
	UserID       uint           `gorm:"not null;index" json:"user_id" validate:"required"`
	FileName     string         `gorm:"size:255;not null" json:"file_name" validate:"required"`
	FileType     string         `gorm:"size:20;not null" json:"file_type" validate:"required,oneof=csv xlsx json"`
	FileHash     string         `gorm:"size:128" json:"file_hash"`
	TotalRecords int            `json:"total_records"`
	ProcessedRecords int        `json:"processed_records"`
	SuccessRecords int           `json:"success_records"`
	FailedRecords int           `json:"failed_records"`
	Status       string         `gorm:"size:20;default:'pending'" json:"status" validate:"omitempty,oneof=pending processing completed failed cancelled"`
	ErrorMessage string         `gorm:"type:text" json:"error_message"`
	StartedAt    *time.Time     `json:"started_at"`
	CompletedAt  *time.Time     `json:"completed_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BulkImportJob) TableName() string {
	return "bulk_import_jobs"
}

type ImportStatusTracking struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	JobID         uint           `gorm:"not null;index" json:"job_id" validate:"required"`
	RecordNumber  int            `gorm:"not null" json:"record_number" validate:"required"`
	RecordData    string         `gorm:"type:text" json:"record_data"`
	ValidationType string        `gorm:"size:20;not null" json:"validation_type" validate:"required,oneof=CURP RFC INE BIOMETRIC SIGNATURE"`
	Status        string         `gorm:"size:20;not null" json:"status" validate:"required,oneof=pending processing success failed"`
	ErrorMessage  string         `gorm:"type:text" json:"error_message"`
	ProcessingTime int64         `json:"processing_time_ms"`
	RetryCount    int            `gorm:"default:0" json:"retry_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ImportStatusTracking) TableName() string {
	return "import_status_tracking"
}

type BulkImportStats struct {
	JobID           uint  `gorm:"primaryKey" json:"job_id"`
	TotalTime       int64 `json:"total_time_ms"`
	AverageTime     int64 `json:"average_time_ms"`
	MinTime         int64 `json:"min_time_ms"`
	MaxTime         int64 `json:"max_time_ms"`
	SuccessRate     float64 `json:"success_rate"`
	FailureRate     float64 `json:"failure_rate"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (BulkImportStats) TableName() string {
	return "bulk_import_stats"
}