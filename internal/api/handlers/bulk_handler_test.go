package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBulkImportService struct {
	mock.Mock
}

func (m *MockBulkImportService) ParseCSV(content []byte) ([]map[string]string, error) {
	args := m.Called(content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]string), args.Error(1)
}

func (m *MockBulkImportService) ParseExcel(content []byte) ([]map[string]string, error) {
	args := m.Called(content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]string), args.Error(1)
}

func (m *MockBulkImportService) CreateJob(userID uint, fileName, fileType string, totalRows int) (*BulkImportJob, error) {
	args := m.Called(userID, fileName, fileType, totalRows)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BulkImportJob), args.Error(1)
}

func (m *MockBulkImportService) GetJobStatus(jobID uint) (*BulkImportJob, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BulkImportJob), args.Error(1)
}

func (m *MockBulkImportService) ProcessJob(jobID uint) error {
	args := m.Called(jobID)
	return args.Error(0)
}

func (m *MockBulkImportService) GetJobRecords(jobID uint) ([]*BulkImportRecord, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*BulkImportRecord), args.Error(1)
}

type CreateJobRequest struct {
	UserID   uint   `json:"user_id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

type UploadCSVRequest struct {
	UserID    uint   `json:"user_id"`
	Content   string `json:"content"`
	FileName  string `json:"file_name"`
}

type UploadExcelRequest struct {
	UserID    uint   `json:"user_id"`
	Content   string `json:"content"`
	FileName  string `json:"file_name"`
}

func TestBulkHandler_CreateJob(t *testing.T) {
	app := fiber.New()
	mockService := new(MockBulkImportService)

	app.Post("/api/v1/bulk/jobs", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"id":         1,
			"user_id":    1,
			"file_name":  "import.csv",
			"file_type":  "csv",
			"status":     "pending",
			"total_rows": 100,
		})
	})

	t.Run("create job successfully", func(t *testing.T) {
		now := time.Now()
		expectedJob := &BulkImportJob{
			ID:        1,
			UserID:    1,
			FileName:  "import.csv",
			FileType:  "csv",
			Status:    "pending",
			TotalRows: 100,
			CreatedAt: now,
		}

		mockService.On("CreateJob", uint(1), "import.csv", "csv", 100).Return(expectedJob, nil)

		body := CreateJobRequest{
			UserID:   1,
			FileName: "import.csv",
			FileType: "csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/jobs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(1), result["id"])
		assert.Equal(t, "import.csv", result["file_name"])
		mockService.AssertExpectations(t)
	})

	t.Run("create job for Excel file", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/bulk/jobs", func(c *fiber.Ctx) error {
			return c.Status(201).JSON(fiber.Map{
				"id":         2,
				"file_type":  "xlsx",
				"total_rows": 500,
			})
		})

		body := CreateJobRequest{
			UserID:   1,
			FileName: "import.xlsx",
			FileType: "xlsx",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/jobs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "xlsx", result["file_type"])
	})

	t.Run("missing file name", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/bulk/jobs", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "file_name is required",
			})
		})

		body := CreateJobRequest{
			UserID:   1,
			FileType: "csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/jobs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("unsupported file type", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/bulk/jobs", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "unsupported file type",
			})
		})

		body := CreateJobRequest{
			UserID:   1,
			FileName: "import.txt",
			FileType: "txt",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/jobs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestBulkHandler_GetJobStatus(t *testing.T) {
	app := fiber.New()
	mockService := new(MockBulkImportService)

	app.Get("/api/v1/bulk/jobs/:id", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"id":         1,
			"status":     "processing",
			"total_rows": 100,
			"processed":  50,
			"success":    48,
			"failed":     2,
		})
	})

	t.Run("get job status - pending", func(t *testing.T) {
		now := time.Now()
		expectedJob := &BulkImportJob{
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

		mockService.On("GetJobStatus", uint(1)).Return(expectedJob, nil)

		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/1", nil)
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("get job status - processing", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/api/v1/bulk/jobs/:id", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"status":     "processing",
				"progress":   50.0,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/1", nil)
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "processing", result["status"])
	})

	t.Run("get job status - completed", func(t *testing.T) {
		app2 := fiber.New()
		now := time.Now()
		completed := now.Add(5 * time.Minute)
		app2.Get("/api/v1/bulk/jobs/:id", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"status":       "completed",
				"total_rows":   100,
				"processed":   100,
				"success":      98,
				"failed":       2,
				"completed_at": completed,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/1", nil)
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("get job status - not found", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/api/v1/bulk/jobs/:id", func(c *fiber.Ctx) error {
			return c.Status(404).JSON(fiber.Map{
				"error": "job not found",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/999", nil)
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestBulkHandler_UploadCSV(t *testing.T) {
	app := fiber.New()
	mockService := new(MockBulkImportService)

	app.Post("/api/v1/bulk/upload/csv", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"job_id":     1,
			"total_rows": 2,
			"status":     "pending",
		})
	})

	t.Run("upload valid CSV", func(t *testing.T) {
		csvContent := "curp,rfc,ine_clave\nGOPG900515HDFRRRA5,GOPG900515AB1,GOGP900515HTSRRA05\nMAAR900523HDFRRRA2,MAAR900523AB2,MAAR900523HTSRRB02"
		records := []map[string]string{
			{"curp": "GOPG900515HDFRRRA5", "rfc": "GOPG900515AB1", "ine_clave": "GOGP900515HTSRRA05"},
			{"curp": "MAAR900523HDFRRRA2", "rfc": "MAAR900523AB2", "ine_clave": "MAAR900523HTSRRB02"},
		}

		mockService.On("ParseCSV", []byte(csvContent)).Return(records, nil)

		body := UploadCSVRequest{
			UserID:   1,
			Content:  csvContent,
			FileName: "import.csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/upload/csv", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("upload empty CSV", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/bulk/upload/csv", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "CSV content is empty",
			})
		})

		body := UploadCSVRequest{
			UserID:   1,
			Content:  "",
			FileName: "empty.csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/upload/csv", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("upload malformed CSV", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/bulk/upload/csv", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "malformed CSV content",
			})
		})

		body := UploadCSVRequest{
			UserID:   1,
			Content:  "curp,rfc\nGOPG900515HDFRRRA5", // Missing column
			FileName: "malformed.csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/upload/csv", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("upload large CSV", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/bulk/upload/csv", func(c *fiber.Ctx) error {
			return c.Status(201).JSON(fiber.Map{
				"job_id":     1,
				"total_rows": 10000,
			})
		})

		csvContent := "curp\n" + repeatString("TEST000101HDFRRRA5\n", 10000)

		body := UploadCSVRequest{
			UserID:   1,
			Content:  csvContent,
			FileName: "large.csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/upload/csv", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	})
}

func TestBulkHandler_UploadExcel(t *testing.T) {
	app := fiber.New()
	mockService := new(MockBulkImportService)

	app.Post("/api/v1/bulk/upload/excel", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"job_id":     1,
			"total_rows": 2,
			"status":     "pending",
		})
	})

	t.Run("upload valid Excel", func(t *testing.T) {
		records := []map[string]string{
			{"curp": "GOPG900515HDFRRRA5", "rfc": "GOPG900515AB1"},
			{"curp": "MAAR900523HDFRRRA2", "rfc": "MAAR900523AB2"},
		}

		mockService.On("ParseExcel", mock.Anything).Return(records, nil)

		body := UploadExcelRequest{
			UserID:   1,
			Content:  "base64encodedexcel",
			FileName: "import.xlsx",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/upload/excel", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	})

	t.Run("upload empty Excel", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/bulk/upload/excel", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "Excel content is empty",
			})
		})

		body := UploadExcelRequest{
			UserID:   1,
			Content:  "",
			FileName: "empty.xlsx",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/upload/excel", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestBulkHandler_GetJobRecords(t *testing.T) {
	app := fiber.New()

	app.Get("/api/v1/bulk/jobs/:id/records", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"records": []fiber.Map{
				{"row_number": 1, "curp": "GOPG900515HDFRRRA5", "status": "success"},
				{"row_number": 2, "curp": "MAAR900523HDFRRRA2", "status": "success"},
			},
		})
	})

	t.Run("get records for successful job", func(t *testing.T) {
		records := []*BulkImportRecord{
			{
				ID:        1,
				JobID:     1,
				RowNumber: 1,
				CURP:      "GOPG900515HDFRRRA5",
				RFC:       "GOPG900515AB1",
				Status:    "success",
			},
			{
				ID:        2,
				JobID:     1,
				RowNumber: 2,
				CURP:      "MAAR900523HDFRRRA2",
				RFC:       "MAAR900523AB2",
				Status:    "success",
			},
		}

		mockService := new(MockBulkImportService)
		mockService.On("GetJobRecords", uint(1)).Return(records, nil)

		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/1/records", nil)
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		recordsArray := result["records"].([]interface{})
		assert.Len(t, recordsArray, 2)
	})

	t.Run("get records with validation errors", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/api/v1/bulk/jobs/:id/records", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"records": []fiber.Map{
					{"row_number": 1, "curp": "INVALID", "status": "failed", "validation_error": "invalid CURP format"},
				},
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/1/records", nil)
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		recordsArray := result["records"].([]interface{})
		record := recordsArray[0].(map[string]interface{})
		assert.Equal(t, "failed", record["status"])
		assert.NotEmpty(t, record["validation_error"])
	})

	t.Run("get records non-existent job", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/api/v1/bulk/jobs/:id/records", func(c *fiber.Ctx) error {
			return c.Status(404).JSON(fiber.Map{
				"error": "job not found",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/999/records", nil)
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestBulkHandler_JobStatusTracking(t *testing.T) {
	app := fiber.New()

	statuses := []string{"pending", "processing", "completed", "failed"}

	for _, status := range statuses {
		t.Run("status_"+status, func(t *testing.T) {
			app2 := fiber.New()
			app2.Get("/api/v1/bulk/jobs/:id", func(c *fiber.Ctx) error {
				return c.Status(200).JSON(fiber.Map{
					"status": status,
				})
			})

			req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/1", nil)
			resp, err := app2.Test(req, -1)

			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			assert.Equal(t, status, result["status"])
		})
	}
}

func repeatString(s string, count int) string {
	var result string
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}