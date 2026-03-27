package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type JobQueue struct {
	client    *redis.Client
	ctx       context.Context
	queueName string
	workers   int
	stopChan  chan struct{}
	wg        sync.WaitGroup
	jobStore  map[string]*models.BulkImportJob
	mu        sync.RWMutex
}

func NewJobQueue(client *redis.Client, workers int) *JobQueue {
	return &JobQueue{
		client:    client,
		ctx:       context.Background(),
		queueName: "bulk_import_queue",
		workers:   workers,
		jobStore:  make(map[string]*models.BulkImportJob),
	}
}

func (q *JobQueue) Enqueue(jobID string, records []interface{}, jobType string) error {
	jobData := struct {
		JobID   string        `json:"job_id"`
		Type    string        `json:"type"`
		Records []interface{} `json:"records"`
	}{
		JobID:   jobID,
		Type:    jobType,
		Records: records,
	}

	data, err := json.Marshal(jobData)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	job := &models.BulkImportJob{
		ID:        jobID,
		Status:    models.JobStatusQueued,
		Total:     len(records),
		UpdatedAt: time.Now(),
	}

	q.mu.Lock()
	q.jobStore[jobID] = job
	q.mu.Unlock()

	if err := q.client.LPush(q.ctx, q.queueName, data).Err(); err != nil {
		return fmt.Errorf("failed to push job to queue: %w", err)
	}

	return nil
}

func (q *JobQueue) Start(processor func(records []interface{}, jobType string) error) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(processor)
	}
}

func (q *JobQueue) worker(processor func([]interface{}, string) error) {
	defer q.wg.Done()

	for {
		select {
		case <-q.stopChan:
			return
		default:
			result, err := q.client.BRPop(q.ctx, 5*time.Second, q.queueName).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				continue
			}

			var jobData struct {
				JobID   string        `json:"job_id"`
				Type    string        `json:"type"`
				Records []interface{} `json:"records"`
			}

			if err := json.Unmarshal([]byte(result[1]), &jobData); err != nil {
				continue
			}

			q.updateJobStatus(jobData.JobID, models.JobStatusProcessing)

			if err := processor(jobData.Records, jobData.Type); err != nil {
				q.updateJobError(jobData.JobID, err.Error())
			} else {
				q.updateJobStatus(jobData.JobID, models.JobStatusCompleted)
			}
		}
	}
}

func (q *JobQueue) Stop() {
	close(q.stopChan)
	q.wg.Wait()
}

func (q *JobQueue) GetStatus(jobID string) (*models.BulkImportJob, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, exists := q.jobStore[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

func (q *JobQueue) ListByUser(userID string, limit, offset int) ([]models.BulkImportJob, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var jobs []models.BulkImportJob
	for _, job := range q.jobStore {
		if job.UserID == userID {
			jobs = append(jobs, *job)
		}
	}

	return jobs, nil
}

func (q *JobQueue) Cancel(jobID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, exists := q.jobStore[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status == models.JobStatusCompleted {
		return fmt.Errorf("cannot cancel completed job")
	}

	job.Status = models.JobStatusCancelled
	job.UpdatedAt = time.Now()

	return nil
}

func (q *JobQueue) updateJobStatus(jobID string, status models.JobStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if job, exists := q.jobStore[jobID]; exists {
		job.Status = status
		job.UpdatedAt = time.Now()
	}
}

func (q *JobQueue) updateJobError(jobID string, errorMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if job, exists := q.jobStore[jobID]; exists {
		job.Status = models.JobStatusFailed
		job.Error = errorMsg
		job.UpdatedAt = time.Now()
	}
}
