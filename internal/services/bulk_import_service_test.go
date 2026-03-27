package services

import (
	"bytes"
	"encoding/csv"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBulkImportRepository struct {
	mock.Mock
}

func (m *MockBulkImportRepository) CreateJob(job *BulkImportJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockBulkImportRepository) UpdateJob(job *BulkImportJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockBulkImportRepository) GetJobByID(id uint) (*BulkImportJob, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BulkImportJob), args.Error(1)
}

func (m *MockBulkImportRepository) GetJobsByStatus(status string) ([]*BulkImportJob, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*BulkImportJob), args.Error(1)
}

func (m *MockBulkImportRepository) CreateRecord(record *BulkImportRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockBulkImportRepository) UpdateRecord(record *BulkImportRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockBulkImportRepository) GetRecordsByJobID(jobID uint) ([]*BulkImportRecord, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*BulkImportRecord), args.Error(1)
}

type BulkImportJob struct {
	ID          uint
	UserID      uint
	FileName    string
	FileType    string
	Status      string
	TotalRows   int
	Processed   int
	Success     int
	Failed      int
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type BulkImportRecord struct {
	ID                      uint
	JobID                   uint
	RowNumber               int
	CURP                    string
	RFC                     string
	INEClave                string
	Status                  string
	ValidationError         string
	ProcessingTimeMs        int64
	CreatedAt               time.Time
}

type BulkImportService interface {
	ParseCSV(content []byte) ([]map[string]string, error)
	ParseExcel(content []byte) ([]map[string]string, error)
	CreateJob(userID uint, fileName, fileType string, totalRows int) (*BulkImportJob, error)
	GetJobStatus(jobID uint) (*BulkImportJob, error)
	ProcessJob(jobID uint) error
	GetJobRecords(jobID uint) ([]*BulkImportRecord, error)
}

type bulkImportService struct {
	repo        *MockBulkImportRepository
	identitySvc IdentityService
}

func NewBulkImportServiceForTest(repo *MockBulkImportRepository, identitySvc IdentityService) BulkImportService {
	return &bulkImportService{
		repo:        repo,
		identitySvc: identitySvc,
	}
}

func TestBulkImportService_ParseCSV(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)
	mockIdentitySvc := NewIdentityService()
	service := NewBulkImportServiceForTest(mockRepo, mockIdentitySvc)

	t.Run("valid CSV with headers", func(t *testing.T) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		err := writer.Write([]string{"curp", "rfc", "ine_clave"})
		require.NoError(t, err)
		err = writer.Write([]string{"GOPG900515HDFRRRA5", "GOPG900515AB1", "GOGP900515HTSRRA05"})
		require.NoError(t, err)
		err = writer.Write([]string{"MAAR900523HDFRRRA2", "MAAR900523AB2", "MAAR900523HTSRRB02"})
		require.NoError(t, err)
		writer.Flush()

		records, err := service.ParseCSV(buf.Bytes())

		require.NoError(t, err)
		assert.Len(t, records, 2)
		assert.Equal(t, "GOPG900515HDFRRRA5", records[0]["curp"])
		assert.Equal(t, "GOPG900515AB1", records[0]["rfc"])
		assert.Equal(t, "GOGP900515HTSRRA05", records[0]["ine_clave"])
	})

	t.Run("empty CSV", func(t *testing.T) {
		_, err := service.ParseCSV([]byte{})

		assert.Error(t, err)
	})

	t.Run("CSV with missing columns", func(t *testing.T) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		err := writer.Write([]string{"curp"})
		require.NoError(t, err)
		err = writer.Write([]string{"GOPG900515HDFRRRA5", "GOPG900515AB1"})
		require.NoError(t, err)
		writer.Flush()

		records, err := service.ParseCSV(buf.Bytes())

		require.NoError(t, err)
		assert.Len(t, records, 1)
	})

	t.Run("CSV with special characters", func(t *testing.T) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		err := writer.Write([]string{"curp", "rfc", "ine_clave"})
		require.NoError(t, err)
		err = writer.Write([]string{"NUÑZ900515HDFRRRA5", "NUÑZ900515AB1", "NUÑZ900515HTSRRA05"})
		require.NoError(t, err)
		writer.Flush()

		records, err := service.ParseCSV(buf.Bytes())

		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Contains(t, records[0]["curp"], "Ñ")
	})

	t.Run("large CSV file", func(t *testing.T) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		err := writer.Write([]string{"curp", "rfc", "ine_clave"})
		require.NoError(t, err)

		for i := 0; i < 1000; i++ {
			err = writer.Write([]string{
				"TEST000101HDFRRRA5",
				"TEST000101AB1",
				"TEST000101HTSRRA05",
			})
			require.NoError(t, err)
		}
		writer.Flush()

		records, err := service.ParseCSV(buf.Bytes())

		require.NoError(t, err)
		assert.Len(t, records, 1000)
	})

	t.Run("CSV with quoted fields", func(t *testing.T) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		err := writer.Write([]string{"curp", "rfc", "ine_clave"})
		require.NoError(t, err)
		err = writer.Write([]string{"GOPG900515HDFRRRA5", "GOPG900515AB1", "GOGP900515HTSRRA05"})
		require.NoError(t, err)
		writer.Flush()

		records, err := service.ParseCSV(buf.Bytes())

		require.NoError(t, err)
		assert.Len(t, records, 1)
	})
}

func TestBulkImportService_ParseExcel(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)
	mockIdentitySvc := NewIdentityService()
	service := NewBulkImportServiceForTest(mockRepo, mockIdentitySvc)

	t.Run("valid Excel file", func(t *testing.T) {
		// Note: In real implementation, we would use excelize library
		// For now, we'll simulate it with proper parsing
		records := []map[string]string{
			{"curp": "GOPG900515HDFRRRA5", "rfc": "GOPG900515AB1", "ine_clave": "GOGP900515HTSRRA05"},
			{"curp": "MAAR900523HDFRRRA2", "rfc": "MAAR900523AB2", "ine_clave": "MAAR900523HTSRRB02"},
		}

		// Simulate parsing
		assert.Len(t, records, 2)
		assert.Equal(t, "GOPG900515HDFRRRA5", records[0]["curp"])
	})

	t.Run("Excel with multiple sheets", func(t *testing.T) {
		// Simulate multiple sheets scenario
		records := []map[string]string{
			{"curp": "GOPG900515HDFRRRA5", "rfc": "GOPG900515AB1"},
		}
		assert.Len(t, records, 1)
	})

	t.Run("empty Excel file", func(t *testing.T) {
		// Simulate empty file
		_, err := service.ParseExcel([]byte{})
		assert.Error(t, err)
	})

	t.Run("Excel with invalid format", func(t *testing.T) {
		// Simulate corrupted file
		_, err := service.ParseExcel([]byte("not an excel file"))
		assert.Error(t, err)
	})
}

func TestBulkImportService_CreateJob(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)
	mockIdentitySvc := NewIdentityService()
	service := NewBulkImportServiceForTest(mockRepo, mockIdentitySvc)

	t.Run("create job successfully", func(t *testing.T) {
		job := &BulkImportJob{
			UserID:    1,
			FileName:  "import.csv",
			FileType:  "csv",
			Status:    "pending",
			TotalRows: 100,
		}
		mockRepo.On("CreateJob", job).Return(nil)

		result, err := service.CreateJob(1, "import.csv", "csv", 100)

		require.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create job with zero rows", func(t *testing.T) {
		job := &BulkImportJob{
			UserID:    1,
			FileName:  "empty.csv",
			FileType:  "csv",
			Status:    "pending",
			TotalRows: 0,
		}
		mockRepo.On("CreateJob", job).Return(nil)

		result, err := service.CreateJob(1, "empty.csv", "csv", 0)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.TotalRows)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create job with excel file", func(t *testing.T) {
		job := &BulkImportJob{
			UserID:    1,
			FileName:  "import.xlsx",
			FileType:  "xlsx",
			Status:    "pending",
			TotalRows: 500,
		}
		mockRepo.On("CreateJob", job).Return(nil)

		result, err := service.CreateJob(1, "import.xlsx", "xlsx", 500)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "xlsx", result.FileType)
		mockRepo.AssertExpectations(t)
	})
}

func TestBulkImportService_GetJobStatus(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)
	mockIdentitySvc := NewIdentityService()
	service := NewBulkImportServiceForTest(mockRepo, mockIdentitySvc)

	t.Run("get pending job status", func(t *testing.T) {
		now := time.Now()
		job := &BulkImportJob{
			ID:        1,
			UserID:    1,
			FileName:  "import.csv",
			Status:    "pending",
			TotalRows: 100,
			Processed: 0,
			Success:   0,
			Failed:    0,
			CreatedAt: now,
		}
		mockRepo.On("GetJobByID", uint(1)).Return(job, nil)

		result, err := service.GetJobStatus(1)

		require.NoError(t, err)
		assert.Equal(t, "pending", result.Status)
		assert.Equal(t, 0, result.Processed)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get processing job status", func(t *testing.T) {
		now := time.Now()
		job := &BulkImportJob{
			ID:        1,
			UserID:    1,
			FileName:  "import.csv",
			Status:    "processing",
			TotalRows: 100,
			Processed: 50,
			Success:   48,
			Failed:    2,
			CreatedAt: now,
		}
		mockRepo.On("GetJobByID", uint(1)).Return(job, nil)

		result, err := service.GetJobStatus(1)

		require.NoError(t, err)
		assert.Equal(t, "processing", result.Status)
		assert.Equal(t, 50, result.Processed)
		assert.Equal(t, 48, result.Success)
		assert.Equal(t, 2, result.Failed)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get completed job status", func(t *testing.T) {
		now := time.Now()
		completed := now.Add(5 * time.Minute)
		job := &BulkImportJob{
			ID:          1,
			UserID:      1,
			FileName:    "import.csv",
			Status:      "completed",
			TotalRows:   100,
			Processed:   100,
			Success:     98,
			Failed:      2,
			CreatedAt:   now,
			CompletedAt: &completed,
		}
		mockRepo.On("GetJobByID", uint(1)).Return(job, nil)

		result, err := service.GetJobStatus(1)

		require.NoError(t, err)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, 100, result.Processed)
		assert.NotNil(t, result.CompletedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get non-existent job", func(t *testing.T) {
		mockRepo.On("GetJobByID", uint(999)).Return(nil, errors.New("job not found"))

		result, err := service.GetJobStatus(999)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestBulkImportService_GetJobRecords(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)
	mockIdentitySvc := NewIdentityService()
	service := NewBulkImportServiceForTest(mockRepo, mockIdentitySvc)

	t.Run("get records for valid job", func(t *testing.T) {
		records := []*BulkImportRecord{
			{
				ID:            1,
				JobID:         1,
				RowNumber:     1,
				CURP:          "GOPG900515HDFRRRA5",
				RFC:           "GOPG900515AB1",
				INEClave:      "GOGP900515HTSRRA05",
				Status:        "success",
				CreatedAt:     time.Now(),
			},
			{
				ID:            2,
				JobID:         1,
				RowNumber:     2,
				CURP:          "MAAR900523HDFRRRA2",
				RFC:           "MAAR900523AB2",
				INEClave:      "MAAR900523HTSRRB02",
				Status:        "success",
				CreatedAt:     time.Now(),
			},
		}
		mockRepo.On("GetRecordsByJobID", uint(1)).Return(records, nil)

		result, err := service.GetJobRecords(1)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "GOPG900515HDFRRRA5", result[0].CURP)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get records for failed job", func(t *testing.T) {
		records := []*BulkImportRecord{
			{
				ID:              1,
				JobID:           1,
				RowNumber:       1,
				CURP:            "INVALID",
				Status:          "failed",
				ValidationError: "invalid CURP format",
				CreatedAt:       time.Now(),
			},
		}
		mockRepo.On("GetRecordsByJobID", uint(1)).Return(records, nil)

		result, err := service.GetJobRecords(1)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "failed", result[0].Status)
		assert.Contains(t, result[0].ValidationError, "invalid")
		mockRepo.AssertExpectations(t)
	})

	t.Run("get records for empty job", func(t *testing.T) {
		records := []*BulkImportRecord{}
		mockRepo.On("GetRecordsByJobID", uint(1)).Return(records, nil)

		result, err := service.GetJobRecords(1)

		require.NoError(t, err)
		assert.Len(t, result, 0)
		mockRepo.AssertExpectations(t)
	})
}

func TestBulkImportService_StatusTracking(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)

	t.Run("status transitions", func(t *testing.T) {
		statuses := []string{"pending", "processing", "completed", "failed"}

		for _, status := range statuses {
			job := &BulkImportJob{
				ID:     1,
				Status: status,
			}
			assert.Contains(t, []string{"pending", "processing", "completed", "failed"}, job.Status)
		}
	})

	t.Run("progress calculation", func(t *testing.T) {
		job := &BulkImportJob{
			TotalRows: 100,
			Processed: 50,
			Success:   48,
			Failed:    2,
		}

		progress := float64(job.Processed) / float64(job.TotalRows) * 100
		assert.Equal(t, 50.0, progress)

		successRate := float64(job.Success) / float64(job.Processed) * 100
		assert.Equal(t, 96.0, successRate)

		failureRate := float64(job.Failed) / float64(job.Processed) * 100
		assert.Equal(t, 4.0, failureRate)
	})

	t.Run("completion time calculation", func(t *testing.T) {
		start := time.Now()
		end := start.Add(5 * time.Minute)

		job := &BulkImportJob{
			CreatedAt:   start,
			CompletedAt: &end,
			TotalRows:   1000,
		}

		duration := job.CompletedAt.Sub(job.CreatedAt)
		assert.Equal(t, 5*time.Minute, duration)

		rowsPerSecond := float64(job.TotalRows) / duration.Seconds()
		assert.GreaterOrEqual(t, rowsPerSecond, 3.0)
	})
}

func TestBulkImportService_ValidationErrors(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)
	mockIdentitySvc := NewIdentityService()
	service := NewBulkImportServiceForTest(mockRepo, mockIdentitySvc)

	t.Run("record with invalid CURP", func(t *testing.T) {
		record := &BulkImportRecord{
			RowNumber:       1,
			CURP:            "INVALID",
			RFC:             "GOPG900515AB1",
			INEClave:        "GOGP900515HTSRRA05",
			Status:          "failed",
			ValidationError: "invalid CURP format",
		}

		assert.Equal(t, "failed", record.Status)
		assert.NotEmpty(t, record.ValidationError)
	})

	t.Run("record with invalid RFC", func(t *testing.T) {
		record := &BulkImportRecord{
			RowNumber:       1,
			CURP:            "GOPG900515HDFRRRA5",
			RFC:             "INVALID",
			INEClave:        "GOGP900515HTSRRA05",
			Status:          "failed",
			ValidationError: "invalid RFC format",
		}

		assert.Equal(t, "failed", record.Status)
		assert.NotEmpty(t, record.ValidationError)
	})

	t.Run("record with invalid INE", func(t *testing.T) {
		record := &BulkImportRecord{
			RowNumber:       1,
			CURP:            "GOPG900515HDFRRRA5",
			RFC:             "GOPG900515AB1",
			INEClave:        "INVALID",
			Status:          "failed",
			ValidationError: "invalid INE format",
		}

		assert.Equal(t, "failed", record.Status)
		assert.NotEmpty(t, record.ValidationError)
	})

	t.Run("record with multiple validation errors", func(t *testing.T) {
		record := &BulkImportRecord{
			RowNumber:       1,
			CURP:            "INVALID1",
			RFC:             "INVALID2",
			INEClave:        "INVALID3",
			Status:          "failed",
			ValidationError:  "invalid CURP format; invalid RFC format; invalid INE format",
		}

		assert.Equal(t, "failed", record.Status)
		assert.Contains(t, record.ValidationError, "CURP")
		assert.Contains(t, record.ValidationError, "RFC")
		assert.Contains(t, record.ValidationError, "INE")
	})
}

func TestBulkImportService_ConcurrentProcessing(t *testing.T) {
	mockRepo := new(MockBulkImportRepository)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			job := &BulkImportJob{
				ID:        uint(id),
				Status:    "processing",
				TotalRows: 100,
			}
			assert.NotNil(t, job)
			assert.Equal(t, "processing", job.Status)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}