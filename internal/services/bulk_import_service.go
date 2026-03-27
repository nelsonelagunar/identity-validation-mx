package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type BulkImportService interface {
	CreateImportJob(ctx context.Context, userID string, file *multipart.FileHeader, fileType string) (*models.BulkImportJob, error)
	GetImportStatus(ctx context.Context, jobID string) (*models.BulkImportJob, error)
	ListJobs(ctx context.Context, userID string, limit, offset int) ([]models.BulkImportJob, error)
	CancelJob(ctx context.Context, jobID string) error
}

type bulkImportService struct {
	jobQueue  *JobQueue
	csvProc   *CSVProcessor
	excelProc *ExcelProcessor
}

func NewBulkImportService(jobQueue *JobQueue) BulkImportService {
	return &bulkImportService{
		jobQueue:  jobQueue,
		csvProc:   NewCSVProcessor(),
		excelProc: NewExcelProcessor(),
	}
}

func (s *bulkImportService) CreateImportJob(ctx context.Context, userID string, file *multipart.FileHeader, fileType string) (*models.BulkImportJob, error) {
	job := &models.BulkImportJob{
		ID:        uuid.New().String(),
		UserID:    userID,
		FileName:  file.Filename,
		Status:    models.JobStatusPending,
		Total:     0,
		Processed: 0,
		Failed:    0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fileExt := strings.ToLower(file.Filename[strings.LastIndex(file.Filename, ".")+1:])

	f, err := file.Open()
	if err != nil {
		job.Status = models.JobStatusFailed
		job.Error = err.Error()
		return job, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var records []interface{}

	switch fileExt {
	case "csv":
		records, err = s.csvProc.Process(f)
	case "xlsx", "xls":
		records, err = s.excelProc.Process(f)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", fileExt)
	}

	if err != nil {
		job.Status = models.JobStatusFailed
		job.Error = err.Error()
		return job, fmt.Errorf("failed to process file: %w", err)
	}

	job.Total = len(records)

	if err := s.jobQueue.Enqueue(job.ID, records, fileType); err != nil {
		job.Status = models.JobStatusFailed
		job.Error = err.Error()
		return job, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return job, nil
}

func (s *bulkImportService) GetImportStatus(ctx context.Context, jobID string) (*models.BulkImportJob, error) {
	return s.jobQueue.GetStatus(jobID)
}

func (s *bulkImportService) ListJobs(ctx context.Context, userID string, limit, offset int) ([]models.BulkImportJob, error) {
	return s.jobQueue.ListByUser(userID, limit, offset)
}

func (s *bulkImportService) CancelJob(ctx context.Context, jobID string) error {
	return s.jobQueue.Cancel(jobID)
}
