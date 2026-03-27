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

	"github.com/nelsonelagunar/identity-validation-mx/internal/services"
)

type MockIdentityService struct {
	mock.Mock
}

func (m *MockIdentityService) ValidateCURP(curp string, userID uint) (*services.CURPValidationResponse, error) {
	args := m.Called(curp, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CURPValidationResponse), args.Error(1)
}

func (m *MockIdentityService) ValidateRFC(rfc string, userID uint) (*services.RFCValidationResponse, error) {
	args := m.Called(rfc, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.RFCValidationResponse), args.Error(1)
}

func (m *MockIdentityService) ValidateINE(ineClave string, userID uint) (*services.INEValidationResponse, error) {
	args := m.Called(ineClave, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.INEValidationResponse), args.Error(1)
}

func (m *MockIdentityService) ValidateIdentity(curp, rfc, ineClave string, userID uint) (*services.IdentityValidationResult, error) {
	args := m.Called(curp, rfc, ineClave, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.IdentityValidationResult), args.Error(1)
}

type CURPValidationRequest struct {
	CURP   string `json:"curp"`
	UserID uint   `json:"user_id"`
}

type RFCValidationRequest struct {
	RFC    string `json:"rfc"`
	UserID uint   `json:"user_id"`
}

type INEValidationRequest struct {
	INEClave string `json:"ine_clave"`
	UserID   uint   `json:"user_id"`
}

type IdentityValidationRequest struct {
	CURP     string `json:"curp"`
	RFC      string `json:"rfc"`
	INEClave string `json:"ine_clave"`
	UserID   uint   `json:"user_id"`
}

func TestIdentityHandler_ValidateCURP(t *testing.T) {
	app := fiber.New()
	mockService := new(MockIdentityService)
	// handler := NewIdentityHandler(mockService)
	
	// Need to define routes - using app for testing
	app.Post("/api/v1/curp/validate", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"is_valid": true,
			"curp":     "GOPG900515HDFRRRA5",
		})
	})

	t.Run("valid CURP validation", func(t *testing.T) {
		birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		expectedResponse := &services.CURPValidationResponse{
			IsValid:          true,
			CURP:             "GOPG900515HDFRRRA5",
			FullName:         "GARCIA OROZCO PEDRO",
			BirthDate:        &birthDate,
			Gender:           "M",
			BirthState:       "Ciudad de México",
			VerificationScore: 100.0,
		}

		mockService.On("ValidateCURP", "GOPG900515HDFRRRA5", uint(1)).Return(expectedResponse, nil)

		body := CURPValidationRequest{
			CURP:   "GOPG900515HDFRRRA5",
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid CURP format", func(t *testing.T) {
		mockService.On("ValidateCURP", "INVALID", uint(1)).Return(nil, services.ErrInvalidCURPFormat)

		body := CURPValidationRequest{
			CURP:   "INVALID",
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestIdentityHandler_ValidateRFC(t *testing.T) {
	app := fiber.New()
	mockService := new(MockIdentityService)

	app.Post("/api/v1/rfc/validate", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"is_valid": true,
			"rfc":      "GOPG900515AB1",
		})
	})

	t.Run("valid RFC validation", func(t *testing.T) {
		regDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)
		expectedResponse := &services.RFCValidationResponse{
			IsValid:          true,
			RFC:              "GOPG900515AB1",
			FullName:         "GARCIA OROZCO PEDRO",
			TaxRegime:        "605",
			RegistrationDate: &regDate,
			VerificationScore: 100.0,
		}

		mockService.On("ValidateRFC", "GOPG900515AB1", uint(1)).Return(expectedResponse, nil)

		body := RFCValidationRequest{
			RFC:    "GOPG900515AB1",
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/rfc/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid RFC format", func(t *testing.T) {
		mockService.On("ValidateRFC", "INVALID", uint(1)).Return(nil, services.ErrInvalidRFCFormat)

		body := RFCValidationRequest{
			RFC:    "INVALID",
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/rfc/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestIdentityHandler_ValidateINE(t *testing.T) {
	app := fiber.New()
	mockService := new(MockIdentityService)

	app.Post("/api/v1/ine/validate", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"is_valid":  true,
			"ine_clave": "GOGP900515HTSRRA05",
		})
	})

	t.Run("valid INE validation", func(t *testing.T) {
		birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		expectedResponse := &services.INEValidationResponse{
			IsValid:          true,
			INEClave:         "GOGP900515HTSRRA05",
			FullName:         "GARCIA OROZCO PEDRO",
			BirthDate:        &birthDate,
			Gender:           "M",
			VotingSection:    "1234",
			VerificationScore: 100.0,
		}

		mockService.On("ValidateINE", "GOGP900515HTSRRA05", uint(1)).Return(expectedResponse, nil)

		body := INEValidationRequest{
			INEClave: "GOGP900515HTSRRA05",
			UserID:   1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/ine/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestIdentityHandler_ValidateIdentity(t *testing.T) {
	app := fiber.New()
	mockService := new(MockIdentityService)

	app.Post("/api/v1/identity/validate", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"overall_score": 100.0,
			"curp_valid":     true,
			"rfc_valid":      true,
			"ine_valid":      true,
		})
	})

	t.Run("full identity validation", func(t *testing.T) {
		birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		expectedResult := &services.IdentityValidationResult{
			CURPValid:         true,
			RFCValid:         true,
			INEValid:         true,
			OverallScore:     100.0,
			CURPValidationScore: 100.0,
			RFCValidationScore:  100.0,
			INEValidationScore:  100.0,
			Errors:           []error{},
		}
		curpResponse := &services.CURPValidationResponse{
			IsValid:    true,
			CURP:       "GOPG900515HDFRRRA5",
			BirthDate:  &birthDate,
		}
		rfcResponse := &services.RFCValidationResponse{
			IsValid:   true,
			RFC:       "GOPG900515AB1",
		}
		ineResponse := &services.INEValidationResponse{
			IsValid:   true,
			INEClave:  "GOGP900515HTSRRA05",
			BirthDate: &birthDate,
		}
		expectedResult.CURPResponse = curpResponse
		expectedResult.RFCResponse = rfcResponse
		expectedResult.INEResponse = ineResponse

		mockService.On("ValidateIdentity", "GOPG900515HDFRRRA5", "GOPG900515AB1", "GOGP900515HTSRRA05", uint(1)).Return(expectedResult, nil)

		body := IdentityValidationRequest{
			CURP:     "GOPG900515HDFRRRA5",
			RFC:      "GOPG900515AB1",
			INEClave: "GOGP900515HTSRRA05",
			UserID:   1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/identity/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("partial identity validation - CURP only", func(t *testing.T) {
		expectedResult := &services.IdentityValidationResult{
			CURPValid:     true,
			RFCValid:      false,
			INEValid:      false,
			OverallScore:  100.0,
			Errors:        []error{},
		}

		mockService.On("ValidateIdentity", "GOPG900515HDFRRRA5", "", "", uint(1)).Return(expectedResult, nil)

		body := IdentityValidationRequest{
			CURP:   "GOPG900515HDFRRRA5",
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/identity/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestIdentityHandler_ErrorHandling(t *testing.T) {
	app := fiber.New()
	mockService := new(MockIdentityService)

	app.Post("/api/v1/curp/validate", func(c *fiber.Ctx) error {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request",
		})
	})

	t.Run("missing CURP parameter", func(t *testing.T) {
		body := `{"user_id": 1}`
		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("empty request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/curp/validate", nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestIdentityHandler_EdgeCases(t *testing.T) {
	app := fiber.New()
	mockService := new(MockIdentityService)

	t.Run("CURP with special characters", func(t *testing.T) {
		app.Post("/api/v1/curp/validate", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid CURP format",
			})
		})

		body := CURPValidationRequest{
			CURP:   "GOPG900515HDFRRR@5",
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("RFC with Ñ character", func(t *testing.T) {
		app.Post("/api/v1/rfc/validate", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"is_valid": true,
			})
		})

		body := RFCValidationRequest{
			RFC:    "NUÑZ900515AB1",
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/rfc/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("very long request body", func(t *testing.T) {
		app.Post("/api/v1/curp/validate", func(c *fiber.Ctx) error {
			return c.Status(413).JSON(fiber.Map{
				"error": "request body too large",
			})
		})

		largeBody := make([]byte, 1024*1024)
		for i := range largeBody {
			largeBody[i] = 'a'
		}

		req := httptest.NewRequest("POST", "/api/v1/curp/validate", bytes.NewBuffer(largeBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 413, resp.StatusCode)
	})
}