package repository

import (
	"context"

	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type JobRepository interface {
	Create(ctx context.Context, job *models.BulkImportJob) error
	GetByID(ctx context.Context, id string) (*models.BulkImportJob, error)
	Update(ctx context.Context, job *models.BulkImportJob) error
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]models.BulkImportJob, error)
	ListByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.BulkImportJob, error)
}

type jobRepository struct {
	db Database
}

func NewJobRepository(db Database) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Create(ctx context.Context, job *models.BulkImportJob) error {
	return r.db.Create(job)
}

func (r *jobRepository) GetByID(ctx context.Context, id string) (*models.BulkImportJob, error) {
	var job models.BulkImportJob
	if err := r.db.First(&job, "id = ?", id); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) Update(ctx context.Context, job *models.BulkImportJob) error {
	return r.db.Save(job)
}

func (r *jobRepository) Delete(ctx context.Context, id string) error {
	return r.db.Delete(&models.BulkImportJob{}, "id = ?", id)
}

func (r *jobRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]models.BulkImportJob, error) {
	var jobs []models.BulkImportJob
	if err := r.db.Find(&jobs, "user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", userID, limit, offset); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *jobRepository) ListByStatus(ctx context.Context, status models.JobStatus, limit int) ([]models.BulkImportJob, error) {
	var jobs []models.BulkImportJob
	if err := r.db.Find(&jobs, "status = ? ORDER BY created_at DESC LIMIT ?", status, limit); err != nil {
		return nil, err
	}
	return jobs, nil
}
