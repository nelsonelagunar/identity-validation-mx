package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
	"gorm.io/gorm"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type JobStore interface {
	SaveJob(ctx context.Context, job *models.BulkImportJob) error
	GetJobByID(ctx context.Context, id uint) (*models.BulkImportJob, error)
	GetJobByJobID(ctx context.Context, jobID string) (*models.BulkImportJob, error)
	UpdateJobStatus(ctx context.Context, jobID string, status string, errorMsg string) error
	UpdateJobProgress(ctx context.Context, jobID string, processed, success, failed int) error
	ListJobs(ctx context.Context, userID uint, limit, offset int) ([]models.BulkImportJob, int64, error)
	DeleteJob(ctx context.Context, jobID string) error

	SaveRecords(ctx context.Context, jobID uint, records []ValidationRequest) error
	GetRecords(ctx context.Context, jobID uint) ([]models.ImportStatusTracking, error)
	UpdateRecordStatus(ctx context.Context, recordID uint, status string, errorMsg string, processingTime int64) error
	GetRecordByRowNumber(ctx context.Context, jobID uint, rowNumber int) (*models.ImportStatusTracking, error)

	SaveStats(ctx context.Context, stats *models.BulkImportStats) error
	GetStats(ctx context.Context, jobID uint) (*models.BulkImportStats, error)
}

type GormJobStore struct {
	db *gorm.DB
}

func NewGormJobStore(db *gorm.DB) JobStore {
	return &GormJobStore{db: db}
}

func (s *GormJobStore) SaveJob(ctx context.Context, job *models.BulkImportJob) error {
	return s.db.WithContext(ctx).Save(job).Error
}

func (s *GormJobStore) GetJobByID(ctx context.Context, id uint) (*models.BulkImportJob, error) {
	var job models.BulkImportJob
	if err := s.db.WithContext(ctx).First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &job, nil
}

func (s *GormJobStore) GetJobByJobID(ctx context.Context, jobID string) (*models.BulkImportJob, error) {
	var job models.BulkImportJob
	if err := s.db.WithContext(ctx).Where("job_id = ?", jobID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &job, nil
}

func (s *GormJobStore) UpdateJobStatus(ctx context.Context, jobID string, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == "processing" {
		now := time.Now()
		updates["started_at"] = &now
	} else if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return s.db.WithContext(ctx).
		Model(&models.BulkImportJob{}).
		Where("job_id = ?", jobID).
		Updates(updates).Error
}

func (s *GormJobStore) UpdateJobProgress(ctx context.Context, jobID string, processed, success, failed int) error {
	return s.db.WithContext(ctx).
		Model(&models.BulkImportJob{}).
		Where("job_id = ?", jobID).
		Updates(map[string]interface{}{
			"processed_records": processed,
			"success_records":   success,
			"failed_records":    failed,
			"updated_at":        time.Now(),
		}).Error
}

func (s *GormJobStore) ListJobs(ctx context.Context, userID uint, limit, offset int) ([]models.BulkImportJob, int64, error) {
	var jobs []models.BulkImportJob
	var total int64

	query := s.db.WithContext(ctx).Model(&models.BulkImportJob{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

func (s *GormJobStore) DeleteJob(ctx context.Context, jobID string) error {
	return s.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Delete(&models.BulkImportJob{}).Error
}

func (s *GormJobStore) SaveRecords(ctx context.Context, jobID uint, records []ValidationRequest) error {
	if len(records) == 0 {
		return nil
	}

	trackings := make([]models.ImportStatusTracking, len(records))
	for i, record := range records {
		recordData, _ := json.Marshal(record)
		trackings[i] = models.ImportStatusTracking{
			JobID:          jobID,
			RecordNumber:   record.RowNumber,
			RecordData:     string(recordData),
			ValidationType: record.ValidationType,
			Status:         "pending",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	}

	return s.db.WithContext(ctx).CreateInBatches(&trackings, 100).Error
}

func (s *GormJobStore) GetRecords(ctx context.Context, jobID uint) ([]models.ImportStatusTracking, error) {
	var records []models.ImportStatusTracking
	if err := s.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("record_number ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (s *GormJobStore) UpdateRecordStatus(ctx context.Context, recordID uint, status string, errorMsg string, processingTime int64) error {
	updates := map[string]interface{}{
		"status":          status,
		"processing_time": processingTime,
		"updated_at":      time.Now(),
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return s.db.WithContext(ctx).
		Model(&models.ImportStatusTracking{}).
		Where("id = ?", recordID).
		Updates(updates).Error
}

func (s *GormJobStore) GetRecordByRowNumber(ctx context.Context, jobID uint, rowNumber int) (*models.ImportStatusTracking, error) {
	var record models.ImportStatusTracking
	if err := s.db.WithContext(ctx).
		Where("job_id = ? AND record_number = ?", jobID, rowNumber).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (s *GormJobStore) SaveStats(ctx context.Context, stats *models.BulkImportStats) error {
	return s.db.WithContext(ctx).Save(stats).Error
}

func (s *GormJobStore) GetStats(ctx context.Context, jobID uint) (*models.BulkImportStats, error) {
	var stats models.BulkImportStats
	if err := s.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		First(&stats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &stats, nil
}

type SQLJobStore struct {
	db *sql.DB
}

func NewSQLJobStore(db *sql.DB) JobStore {
	return &SQLJobStore{db: db}
}

func (s *SQLJobStore) SaveJob(ctx context.Context, job *models.BulkImportJob) error {
	query := `
		INSERT INTO bulk_import_jobs (job_id, user_id, file_name, file_type, file_hash, 
			total_records, processed_records, success_records, failed_records, status, 
			error_message, started_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (job_id) DO UPDATE SET
			processed_records = EXCLUDED.processed_records,
			success_records = EXCLUDED.success_records,
			failed_records = EXCLUDED.failed_records,
			status = EXCLUDED.status,
			error_message = EXCLUDED.error_message,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	now := time.Now()
	err := s.db.QueryRowContext(ctx, query,
		job.JobID, job.UserID, job.FileName, job.FileType, job.FileHash,
		job.TotalRecords, job.ProcessedRecords, job.SuccessRecords, job.FailedRecords,
		job.Status, job.ErrorMessage, job.StartedAt, job.CompletedAt, job.CreatedAt, job.UpdatedAt,
	).Scan(&job.ID)

	return err
}

func (s *SQLJobStore) GetJobByID(ctx context.Context, id uint) (*models.BulkImportJob, error) {
	query := `
		SELECT id, job_id, user_id, file_name, file_type, file_hash,
			total_records, processed_records, success_records, failed_records,
			status, error_message, started_at, completed_at, created_at, updated_at
		FROM bulk_import_jobs WHERE id = $1
	`

	job := &models.BulkImportJob{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.JobID, &job.UserID, &job.FileName, &job.FileType, &job.FileHash,
		&job.TotalRecords, &job.ProcessedRecords, &job.SuccessRecords, &job.FailedRecords,
		&job.Status, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	return job, err
}

func (s *SQLJobStore) GetJobByJobID(ctx context.Context, jobID string) (*models.BulkImportJob, error) {
	query := `
		SELECT id, job_id, user_id, file_name, file_type, file_hash,
			total_records, processed_records, success_records, failed_records,
			status, error_message, started_at, completed_at, created_at, updated_at
		FROM bulk_import_jobs WHERE job_id = $1
	`

	job := &models.BulkImportJob{}
	err := s.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.JobID, &job.UserID, &job.FileName, &job.FileType, &job.FileHash,
		&job.TotalRecords, &job.ProcessedRecords, &job.SuccessRecords, &job.FailedRecords,
		&job.Status, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	return job, err
}

func (s *SQLJobStore) UpdateJobStatus(ctx context.Context, jobID string, status string, errorMsg string) error {
	var startedAt, completedAt interface{}

	if status == "processing" {
		now := time.Now()
		startedAt = &now
	} else if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE bulk_import_jobs 
		SET status = $1, error_message = $2, started_at = COALESCE($3, started_at),
			completed_at = COALESCE($4, completed_at), updated_at = $5
		WHERE job_id = $6
	`

	_, err := s.db.ExecContext(ctx, query, status, errorMsg, startedAt, completedAt, time.Now(), jobID)
	return err
}

func (s *SQLJobStore) UpdateJobProgress(ctx context.Context, jobID string, processed, success, failed int) error {
	query := `
		UPDATE bulk_import_jobs 
		SET processed_records = $1, success_records = $2, failed_records = $3, updated_at = $4
		WHERE job_id = $5
	`

	_, err := s.db.ExecContext(ctx, query, processed, success, failed, time.Now(), jobID)
	return err
}

func (s *SQLJobStore) ListJobs(ctx context.Context, userID uint, limit, offset int) ([]models.BulkImportJob, int64, error) {
	var countQuery string
	var query string
	var args []interface{}
	var countArgs []interface{}

	if userID > 0 {
		countQuery = "SELECT COUNT(*) FROM bulk_import_jobs WHERE user_id = $1"
		query = `
			SELECT id, job_id, user_id, file_name, file_type, file_hash,
				total_records, processed_records, success_records, failed_records,
				status, error_message, started_at, completed_at, created_at, updated_at
			FROM bulk_import_jobs WHERE user_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`
		countArgs = []interface{}{userID}
		args = []interface{}{userID, limit, offset}
	} else {
		countQuery = "SELECT COUNT(*) FROM bulk_import_jobs"
		query = `
			SELECT id, job_id, user_id, file_name, file_type, file_hash,
				total_records, processed_records, success_records, failed_records,
				status, error_message, started_at, completed_at, created_at, updated_at
			FROM bulk_import_jobs
			ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []models.BulkImportJob
	for rows.Next() {
		var job models.BulkImportJob
		if err := rows.Scan(
			&job.ID, &job.JobID, &job.UserID, &job.FileName, &job.FileType, &job.FileHash,
			&job.TotalRecords, &job.ProcessedRecords, &job.SuccessRecords, &job.FailedRecords,
			&job.Status, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, job)
	}

	return jobs, total, nil
}

func (s *SQLJobStore) DeleteJob(ctx context.Context, jobID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM bulk_import_jobs WHERE job_id = $1", jobID)
	return err
}

func (s *SQLJobStore) SaveRecords(ctx context.Context, jobID uint, records []ValidationRequest) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO import_status_tracking (job_id, record_number, record_data, validation_type, 
			status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, record := range records {
		recordData, _ := json.Marshal(record)
		_, err := stmt.ExecContext(ctx, jobID, record.RowNumber, string(recordData),
			record.ValidationType, "pending", now, now)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLJobStore) GetRecords(ctx context.Context, jobID uint) ([]models.ImportStatusTracking, error) {
	query := `
		SELECT id, job_id, record_number, record_data, validation_type,
			status, error_message, processing_time, retry_count, created_at, updated_at
		FROM import_status_tracking WHERE job_id = $1 ORDER BY record_number ASC
	`

	rows, err := s.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.ImportStatusTracking
	for rows.Next() {
		var record models.ImportStatusTracking
		if err := rows.Scan(
			&record.ID, &record.JobID, &record.RecordNumber, &record.RecordData,
			&record.ValidationType, &record.Status, &record.ErrorMessage,
			&record.ProcessingTime, &record.RetryCount, &record.CreatedAt, &record.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (s *SQLJobStore) UpdateRecordStatus(ctx context.Context, recordID uint, status string, errorMsg string, processingTime int64) error {
	query := `
		UPDATE import_status_tracking 
		SET status = $1, error_message = $2, processing_time = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := s.db.ExecContext(ctx, query, status, errorMsg, processingTime, time.Now(), recordID)
	return err
}

func (s *SQLJobStore) GetRecordByRowNumber(ctx context.Context, jobID uint, rowNumber int) (*models.ImportStatusTracking, error) {
	query := `
		SELECT id, job_id, record_number, record_data, validation_type,
			status, error_message, processing_time, retry_count, created_at, updated_at
		FROM import_status_tracking WHERE job_id = $1 AND record_number = $2
	`

	record := &models.ImportStatusTracking{}
	err := s.db.QueryRowContext(ctx, query, jobID, rowNumber).Scan(
		&record.ID, &record.JobID, &record.RecordNumber, &record.RecordData,
		&record.ValidationType, &record.Status, &record.ErrorMessage,
		&record.ProcessingTime, &record.RetryCount, &record.CreatedAt, &record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	return record, err
}

func (s *SQLJobStore) SaveStats(ctx context.Context, stats *models.BulkImportStats) error {
	query := `
		INSERT INTO bulk_import_stats (job_id, total_time, average_time, min_time, max_time,
			success_rate, failure_rate, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (job_id) DO UPDATE SET
			total_time = EXCLUDED.total_time,
			average_time = EXCLUDED.average_time,
			min_time = EXCLUDED.min_time,
			max_time = EXCLUDED.max_time,
			success_rate = EXCLUDED.success_rate,
			failure_rate = EXCLUDED.failure_rate,
			updated_at = EXCLUDED.updated_at
	`

	_, err := s.db.ExecContext(ctx, query,
		stats.JobID, stats.TotalTime, stats.AverageTime, stats.MinTime, stats.MaxTime,
		stats.SuccessRate, stats.FailureRate, stats.CreatedAt, stats.UpdatedAt)

	return err
}

func (s *SQLJobStore) GetStats(ctx context.Context, jobID uint) (*models.BulkImportStats, error) {
	query := `
		SELECT job_id, total_time, average_time, min_time, max_time,
			success_rate, failure_rate, created_at, updated_at
		FROM bulk_import_stats WHERE job_id = $1
	`

	stats := &models.BulkImportStats{}
	err := s.db.QueryRowContext(ctx, query, jobID).Scan(
		&stats.JobID, &stats.TotalTime, &stats.AverageTime, &stats.MinTime, &stats.MaxTime,
		&stats.SuccessRate, &stats.FailureRate, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	return stats, err
}

type JobProcessorService struct {
	jobStore    JobStore
	jobQueue    JobQueue
	identitySvc IdentityService
}

func NewJobProcessorService(jobStore JobStore, identitySvc IdentityService) *JobProcessorService {
	return &JobProcessorService{
		jobStore:    jobStore,
		identitySvc: identitySvc,
	}
}

func (p *JobProcessorService) Process(ctx context.Context, payload *JobPayload) error {
	job, err := p.jobStore.GetJobByID(ctx, payload.InternalJobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job.Status == "cancelled" {
		return ErrJobCancelled
	}

	if err := p.jobStore.UpdateJobStatus(ctx, payload.JobID, "processing", ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	records, err := p.jobStore.GetRecords(ctx, payload.InternalJobID)
	if err != nil {
		p.jobStore.UpdateJobStatus(ctx, payload.JobID, "failed", err.Error())
		return fmt.Errorf("failed to get records: %w", err)
	}

	processedCount := 0
	successCount := 0
	failedCount := 0

	for _, record := range records {
		select {
		case <-ctx.Done():
			p.jobStore.UpdateJobStatus(ctx, payload.JobID, "failed", "context cancelled")
			return ctx.Err()
		default:
		}

		var req ValidationRequest
		if err := json.Unmarshal([]byte(record.RecordData), &req); err != nil {
			p.jobStore.UpdateRecordStatus(ctx, record.ID, "failed", err.Error(), 0)
			processedCount++
			failedCount++
			p.jobStore.UpdateJobProgress(ctx, payload.JobID, processedCount, successCount, failedCount)
			continue
		}

		start := time.Now()
		var processingErr error

		switch record.ValidationType {
		case "CURP":
			if req.CURP != "" {
				_, processingErr = p.identitySvc.ValidateCURP(req.CURP, payload.UserID)
			} else {
				processingErr = ErrEmptyInput
			}
		case "RFC":
			if req.RFC != "" {
				_, processingErr = p.identitySvc.ValidateRFC(req.RFC, payload.UserID)
			} else {
				processingErr = ErrEmptyInput
			}
		case "INE":
			if req.INEClave != "" {
				_, processingErr = p.identitySvc.ValidateINE(req.INEClave, payload.UserID)
			} else {
				processingErr = ErrEmptyInput
			}
		default:
			_, processingErr = p.identitySvc.ValidateIdentity(req.CURP, req.RFC, req.INEClave, payload.UserID)
		}

		processingTime := time.Since(start).Milliseconds()

		if processingErr != nil {
			p.jobStore.UpdateRecordStatus(ctx, record.ID, "failed", processingErr.Error(), processingTime)
			failedCount++
		} else {
			p.jobStore.UpdateRecordStatus(ctx, record.ID, "success", "", processingTime)
			successCount++
		}

		processedCount++
		p.jobStore.UpdateJobProgress(ctx, payload.JobID, processedCount, successCount, failedCount)
	}

	status := "completed"
	var errorMsg string
	if failedCount > 0 && successCount == 0 {
		status = "failed"
		errorMsg = "All records failed to process"
	} else if failedCount > 0 {
		status = "completed"
		errorMsg = fmt.Sprintf("Processed with %d failed records", failedCount)
	}

	return p.jobStore.UpdateJobStatus(ctx, payload.JobID, status, errorMsg)
}
