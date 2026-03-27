package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nelsonelagunar/identity-validation-mx/internal/services"
)

type MockBiometricService struct {
	mock.Mock
}

func (m *MockBiometricService) CompareFaces(ctx context.Context, input services.CompareFacesInput) (*services.CompareFacesOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CompareFacesOutput), args.Error(1)
}

func (m *MockBiometricService) DetectLiveness(ctx context.Context, input services.DetectLivenessInput) (*services.DetectLivenessOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.DetectLivenessOutput), args.Error(1)
}

func (m *MockBiometricService) SetProvider(provider services.BiometricProvider) {
	m.Called(provider)
}

func (m *MockBiometricService) GetProvider() services.BiometricProvider {
	args := m.Called()
	return args.Get(0).(services.BiometricProvider)
}

type CompareFacesRequest struct {
	SourceImage     string `json:"source_image"`
	SourceImageType string `json:"source_image_type"`
	TargetImage     string `json:"target_image"`
	TargetImageType string `json:"target_image_type"`
	UserID          uint   `json:"user_id"`
}

type DetectLivenessRequest struct {
	Images      []string `json:"images"`
	ImageTypes  []string `json:"image_types"`
	VideoFile   string   `json:"video_file,omitempty"`
	UserID      uint     `json:"user_id"`
}

func TestBiometricHandler_CompareFaces(t *testing.T) {
	app := fiber.New()
	mockService := new(MockBiometricService)

	app.Post("/api/v1/biometric/compare", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"is_match":          true,
			"similarity_score":   0.98,
			"confidence_level":   0.95,
			"processing_time_ms": 150,
		})
	})

	t.Run("successful face comparison", func(t *testing.T) {
		expectedOutput := &services.CompareFacesOutput{
			IsMatch:          true,
			SimilarityScore:  0.98,
			ConfidenceLevel:  0.95,
			ProcessingTimeMs: 150,
		}

		input := services.CompareFacesInput{
			SourceImage:      "base64encodedsource",
			SourceImageType:  services.ImageTypeBase64,
			TargetImage:      "base64encodedtarget",
			TargetImageType:  services.ImageTypeBase64,
			UserID:           1,
		}

		mockService.On("CompareFaces", mock.Anything, input).Return(expectedOutput, nil)

		body := CompareFacesRequest{
			SourceImage:      "base64encodedsource",
			SourceImageType:  "base64",
			TargetImage:      "base64encodedtarget",
			TargetImageType:  "base64",
			UserID:           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("face comparison - no match", func(t *testing.T) {
		expectedOutput := &services.CompareFacesOutput{
			IsMatch:           false,
			SimilarityScore:   0.35,
			ConfidenceLevel:   0.85,
			DetectedAnomalies: []string{"different_person"},
			ProcessingTimeMs:  120,
		}

		input := services.CompareFacesInput{
			SourceImage:      "differentface1",
			TargetImage:      "differentface2",
			SourceImageType:  services.ImageTypeBase64,
			TargetImageType:  services.ImageTypeBase64,
			UserID:           1,
		}

		mockService.On("CompareFaces", mock.Anything, input).Return(expectedOutput, nil)

		body := CompareFacesRequest{
			SourceImage:      "differentface1",
			TargetImage:      "differentface2",
			SourceImageType:  "base64",
			TargetImageType:  "base64",
			UserID:           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("missing source image", func(t *testing.T) {
		app.Post("/api/v1/biometric/compare", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "source_image is required",
			})
		})

		body := CompareFacesRequest{
			TargetImage:     "base64encodedtarget",
			TargetImageType: "base64",
			UserID:          1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("invalid image type", func(t *testing.T) {
		app.Post("/api/v1/biometric/compare", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid image type",
			})
		})

		body := CompareFacesRequest{
			SourceImage:      "testdata",
			SourceImageType:  "invalid",
			TargetImage:      "testdata",
			TargetImageType:  "base64",
			UserID:           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestBiometricHandler_DetectLiveness(t *testing.T) {
	app := fiber.New()
	mockService := new(MockBiometricService)

	app.Post("/api/v1/biometric/liveness", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"is_live":            true,
			"liveness_score":      0.97,
			"confidence_level":    0.95,
			"spoof_probability":   0.02,
			"processing_time_ms":  250,
		})
	})

	t.Run("successful liveness detection", func(t *testing.T) {
		expectedOutput := &services.DetectLivenessOutput{
			IsLive:           true,
			LivenessScore:    0.97,
			ConfidenceLevel:  0.95,
			SpoofProbability: 0.02,
			ProcessingTimeMs: 250,
		}

		input := services.DetectLivenessInput{
			Images:     []string{"image1", "image2", "image3"},
			ImageTypes: []services.ImageType{services.ImageTypeBase64, services.ImageTypeBase64, services.ImageTypeBase64},
			UserID:     1,
		}

		mockService.On("DetectLiveness", mock.Anything, input).Return(expectedOutput, nil)

		body := DetectLivenessRequest{
			Images:     []string{"image1", "image2", "image3"},
			ImageTypes: []string{"base64", "base64", "base64"},
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/liveness", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("liveness detection - spoof detected", func(t *testing.T) {
		expectedOutput := &services.DetectLivenessOutput{
			IsLive:           false,
			LivenessScore:    0.25,
			ConfidenceLevel:  0.90,
			SpoofProbability: 0.85,
			DetectedAttacks:  []string{"photo_replay", "screen_replay"},
			ProcessingTimeMs: 180,
		}

		input := services.DetectLivenessInput{
			Images:     []string{"spoofimage"},
			ImageTypes: []services.ImageType{services.ImageTypeBase64},
			UserID:     1,
		}

		mockService.On("DetectLiveness", mock.Anything, input).Return(expectedOutput, nil)

		body := DetectLivenessRequest{
			Images:     []string{"spoofimage"},
			ImageTypes: []string{"base64"},
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/liveness", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("missing images", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/biometric/liveness", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "images are required",
			})
		})

		body := DetectLivenessRequest{
			UserID: 1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/liveness", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("empty images array", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/biometric/liveness", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "at least one image is required",
			})
		})

		body := DetectLivenessRequest{
			Images:     []string{},
			ImageTypes: []string{},
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/liveness", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestBiometricHandler_Providers(t *testing.T) {
	app := fiber.New()

	t.Run("AWS provider configuration", func(t *testing.T) {
		app.Get("/api/v1/biometric/provider", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"provider": "aws",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/biometric/provider", nil)
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "aws", result["provider"])
	})

	t.Run("Azure provider configuration", func(t *testing.T) {
		app.Get("/api/v1/biometric/provider", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"provider": "azure",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/biometric/provider", nil)
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("local provider configuration", func(t *testing.T) {
		app.Get("/api/v1/biometric/provider", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"provider": "local",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/biometric/provider", nil)
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestBiometricHandler_ErrorCases(t *testing.T) {
	app := fiber.New()

	t.Run("no face detected", func(t *testing.T) {
		app.Post("/api/v1/biometric/compare", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "no face detected in source image",
			})
		})

		body := CompareFacesRequest{
			SourceImage:      "nofaceimage",
			SourceImageType:  "base64",
			TargetImage:      "validface",
			TargetImageType:  "base64",
			UserID:           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("multiple faces detected", func(t *testing.T) {
		app.Post("/api/v1/biometric/compare", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "multiple faces detected in source image",
			})
		})

		body := CompareFacesRequest{
			SourceImage:      "multiplefaces",
			SourceImageType:  "base64",
			TargetImage:      "singleface",
			TargetImageType:  "base64",
			UserID:           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("low quality image", func(t *testing.T) {
		app.Post("/api/v1/biometric/liveness", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "image quality too low for liveness detection",
			})
		})

		body := DetectLivenessRequest{
			Images:     []string{"lowquality"},
			ImageTypes: []string{"base64"},
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/liveness", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("image too large", func(t *testing.T) {
		app.Post("/api/v1/biometric/compare", func(c *fiber.Ctx) error {
			return c.Status(413).JSON(fiber.Map{
				"error": "image size exceeds limit",
			})
		})

		largeContent := make([]byte, 10*1024*1024)
		for i := range largeContent {
			largeContent[i] = 'a'
		}

		body := CompareFacesRequest{
			SourceImage:      string(largeContent[:1000]),
			SourceImageType:  "base64",
			TargetImage:      string(largeContent[:1000]),
			TargetImageType:  "base64",
			UserID:           1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/biometric/compare", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 413, resp.StatusCode)
	})
}

// Add missing import
import "context"