package tests

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nelsonelagunar/identity-validation-mx/internal/api"
	"github.com/nelsonelagunar/identity-validation-mx/internal/config"
	"github.com/nelsonelagunar/identity-validation-mx/internal/services"
)

type TestServer struct {
	App       *fiber.App
	IdentitySvc services.IdentityService
	BiometricSvc services.BiometricService
}

func SetupTestServer() *TestServer {
	cfg := &config.Config{
		Port:         "3000",
		DatabaseURL:  "test://localhost:5432/testdb",
		RedisURL:     "localhost:6379",
		JWTSecret:    "test-secret",
		LogLevel:     "debug",
	}

	identitySvc := services.NewIdentityService()
	biometricSvc := services.NewBiometricService()

	app := fiber.New(fiber.Config{
		AppName:      "Identity Validation MX Test",
		ServerHeader: "Identity Validation MX",
	})

	api.SetupRoutes(app, identitySvc, biometricSvc)

	return &TestServer{
		App:         app,
		IdentitySvc: identitySvc,
		BiometricSvc: biometricSvc,
	}
}

func (s *TestServer) Close() error {
	return s.App.Shutdown()
}

func TestE2E_HealthCheck(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("GET /health returns OK", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "OK", result["status"])
	})

	t.Run("GET /api/v1/health returns API status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestE2E_CURPValidation(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/curp/validate with valid CURP", func(t *testing.T) {
		body := map[string]interface{}{
			"curp":    "GOPG900515HDFRRRA5",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["is_valid"].(bool))
		assert.NotEmpty(t, result["curp"])
	})

	t.Run("POST /api/v1/curp/validate with invalid CURP", func(t *testing.T) {
		body := map[string]interface{}{
			"curp":    "INVALID",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("POST /api/v1/curp/validate with empty CURP", func(t *testing.T) {
		body := map[string]interface{}{
			"curp":    "",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestE2E_RFCValidation(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/rfc/validate with valid RFC", func(t *testing.T) {
		body := map[string]interface{}{
			"rfc":     "GOPG900515AB1",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/rfc/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["is_valid"].(bool))
	})

	t.Run("POST /api/v1/rfc/validate with moral RFC", func(t *testing.T) {
		body := map[string]interface{}{
			"rfc":     "ABC900620ABC",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/rfc/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestE2E_INEValidation(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/ine/validate with valid INE", func(t *testing.T) {
		body := map[string]interface{}{
			"ine_clave": "GOGP900515HTSRRA05",
			"user_id":   1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/ine/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["is_valid"].(bool))
	})
}

func TestE2E_FullIdentityValidation(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/identity/validate with all fields", func(t *testing.T) {
		body := map[string]interface{}{
			"curp":      "GOPG900515HDFRRRA5",
			"rfc":       "GOPG900515AB1",
			"ine_clave": "GOGP900515HTSRRA05",
			"user_id":   1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/identity/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["curp_valid"].(bool))
		assert.True(t, result["rfc_valid"].(bool))
		assert.True(t, result["ine_valid"].(bool))
		assert.Equal(t, 100.0, result["overall_score"].(float64))
	})

	t.Run("POST /api/v1/identity/validate with partial fields", func(t *testing.T) {
		body := map[string]interface{}{
			"curp":    "GOPG900515HDFRRRA5",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/identity/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["curp_valid"].(bool))
	})
}

func TestE2E_BiometricComparison(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/biometric/compare with matching faces", func(t *testing.T) {
		body := map[string]interface{}{
			"source_image":      "base64encodedsourcedata",
			"source_image_type": "base64",
			"target_image":      "base64encodedtargetdata",
			"target_image_type": "base64",
			"user_id":           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("POST /api/v1/biometric/compare missing source image", func(t *testing.T) {
		body := map[string]interface{}{
			"target_image":      "base64encodedtargetdata",
			"target_image_type": "base64",
			"user_id":           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestE2E_LivenessDetection(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/biometric/liveness with images", func(t *testing.T) {
		body := map[string]interface{}{
			"images":       []string{"image1", "image2", "image3"},
			"image_types":  []string{"base64", "base64", "base64"},
			"user_id":      1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/liveness", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("POST /api/v1/biometric/liveness without images", func(t *testing.T) {
		body := map[string]interface{}{
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/liveness", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestE2E_SignatureOperations(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/signature/xades with valid XML", func(t *testing.T) {
		body := map[string]interface{}{
			"xml_content": "<?xml version=\"1.0\"?><document><data>Test</data></document>",
			"cert_path":    "/path/to/cert.pem",
			"key_path":     "/path/to/key.pem",
			"user_id":      1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("POST /api/v1/signature/pades with valid PDF", func(t *testing.T) {
		body := map[string]interface{}{
			"pdf_content": "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZS9DYXRhbG9n",
			"cert_path":    "/path/to/cert.pem",
			"key_path":     "/path/to/key.pem",
			"user_id":      1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/pades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("POST /api/v1/signature/verify with XAdES signature", func(t *testing.T) {
		body := map[string]interface{}{
			"signed_content": "base64encodedsignedxml",
			"signature_type": "xades",
			"user_id":        1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/verify", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestE2E_BulkImportOperations(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("POST /api/v1/bulk/jobs creates new job", func(t *testing.T) {
		body := map[string]interface{}{
			"user_id":   1,
			"file_name": "import.csv",
			"file_type": "csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/jobs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	})

	t.Run("GET /api/v1/bulk/jobs/:id returns job status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/bulk/jobs/1", nil)
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("POST /api/v1/bulk/upload/csv uploads CSV", func(t *testing.T) {
		csvContent := "curp,rfc,ine_clave\nGOPG900515HDFRRRA5,GOPG900515AB1,GOGP900515HTSRRA05"
		body := map[string]interface{}{
			"user_id":   1,
			"content":   csvContent,
			"file_name": "import.csv",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/bulk/upload/csv", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	})
}

func TestE2E_ConcurrentRequests(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("concurrent CURP validations", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				body := map[string]interface{}{
					"curp":    "GOPG900515HDFRRRA5",
					"user_id": 1,
				}
				jsonBody, _ := json.Marshal(body)

				req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				resp, err := server.App.Test(req, -1)

				assert.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent mixed validations", func(t *testing.T) {
		done := make(chan bool, 30)

		for i := 0; i < 10; i++ {
			go func() {
				body := map[string]interface{}{
					"curp":    "GOPG900515HDFRRRA5",
					"user_id": 1,
				}
				jsonBody, _ := json.Marshal(body)
				req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				resp, _ := server.App.Test(req, -1)
				assert.Equal(t, 200, resp.StatusCode)
				done <- true
			}()

			go func() {
				body := map[string]interface{}{
					"rfc":     "GOPG900515AB1",
					"user_id": 1,
				}
				jsonBody, _ := json.Marshal(body)
				req := httptest.NewRequest("POST", "/api/v1/rfc/validate", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				resp, _ := server.App.Test(req, -1)
				assert.Equal(t, 200, resp.StatusCode)
				done <- true
			}()

			go func() {
				body := map[string]interface{}{
					"ine_clave": "GOGP900515HTSRRA05",
					"user_id":   1,
				}
				jsonBody, _ := json.Marshal(body)
				req := httptest.NewRequest("POST", "/api/v1/ine/validate", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				resp, _ := server.App.Test(req, -1)
				assert.Equal(t, 200, resp.StatusCode)
				done <- true
			}()
		}

		for i := 0; i < 30; i++ {
			<-done
		}
	})
}

func TestE2E_RequestTimeout(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("request within timeout", func(t *testing.T) {
		start := time.Now()

		body := map[string]interface{}{
			"curp":    "GOPG900515HDFRRRA5",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, 5000) // 5 second timeout

		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Less(t, elapsed.Milliseconds(), int64(5000))
	})
}

func TestE2E_ResponseFormat(t *testing.T) {
	server := SetupTestServer()
	defer server.Close()

	t.Run("response has correct content type", func(t *testing.T) {
		body := map[string]interface{}{
			"curp":    "GOPG900515HDFRRRA5",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("response contains required fields", func(t *testing.T) {
		body := map[string]interface{}{
			"curp":    "GOPG900515HDFRRRA5",
			"user_id": 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req, -1)

		require.NoError(t, err)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Contains(t, result, "is_valid")
		assert.Contains(t, result, "curp")
	})
}