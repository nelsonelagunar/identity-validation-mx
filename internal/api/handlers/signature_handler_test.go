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

type MockSignatureService struct {
	mock.Mock
}

func (m *MockSignatureService) SignXML(xmlContent string, certPath, keyPath string) ([]byte, error) {
	args := m.Called(xmlContent, certPath, keyPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSignatureService) SignPDF(pdfContent []byte, certPath, keyPath string) ([]byte, error) {
	args := m.Called(pdfContent, certPath, keyPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSignatureService) VerifySignature(signedContent []byte, signatureType string) (*services.SignatureVerificationResult, error) {
	args := m.Called(signedContent, signatureType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.SignatureVerificationResult), args.Error(1)
}

type SignXMLRequest struct {
	XMLContent string `json:"xml_content"`
	CertPath   string `json:"cert_path"`
	KeyPath    string `json:"key_path"`
	UserID     uint   `json:"user_id"`
}

type SignPDFRequest struct {
	PDFContent string `json:"pdf_content"`
	CertPath   string `json:"cert_path"`
	KeyPath    string `json:"key_path"`
	UserID     uint   `json:"user_id"`
}

type VerifySignatureRequest struct {
	SignedContent string `json:"signed_content"`
	SignatureType string `json:"signature_type"`
	UserID        uint   `json:"user_id"`
}

func TestSignatureHandler_SignXML(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSignatureService)

	app.Post("/api/v1/signature/xades", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"signed_content":    "base64encodedsignedxml",
			"signature_type":     "XAdES",
			"processing_time_ms": 150,
		})
	})

	t.Run("successful XAdES signing", func(t *testing.T) {
		xmlContent := `<?xml version="1.0"?><document><data>Test</data></document>`
		signedXML := []byte(`<?xml version="1.0"?><document><data>Test</data><ds:Signature>...</ds:Signature></document>`)

		mockService.On("SignXML", xmlContent, "/path/to/cert.pem", "/path/to/key.pem").Return(signedXML, nil)

		body := SignXMLRequest{
			XMLContent: xmlContent,
			CertPath:   "/path/to/cert.pem",
			KeyPath:    "/path/to/key.pem",
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("missing certificate path", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/signature/xades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "cert_path is required",
			})
		})

		body := SignXMLRequest{
			XMLContent: "<xml>test</xml>",
			KeyPath:    "/path/to/key.pem",
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("missing key path", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/signature/xades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "key_path is required",
			})
		})

		body := SignXMLRequest{
			XMLContent: "<xml>test</xml>",
			CertPath:   "/path/to/cert.pem",
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("empty XML content", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/signature/xades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "xml_content is required",
			})
		})

		body := SignXMLRequest{
			XMLContent: "",
			CertPath:   "/path/to/cert.pem",
			KeyPath:    "/path/to/key.pem",
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestSignatureHandler_SignPDF(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSignatureService)

	app.Post("/api/v1/signature/pades", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"signed_content":    "base64encodedsignedpdf",
			"signature_type":    "PAdES",
			"processing_time_ms": 200,
		})
	})

	t.Run("successful PAdES signing", func(t *testing.T) {
		pdfContent := []byte("%PDF-1.4 test pdf content")
		signedPDF := append(pdfContent, []byte("\n% Signed with PAdES")...)

		mockService.On("SignPDF", pdfContent, "/path/to/cert.pem", "/path/to/key.pem").Return(signedPDF, nil)

		body := SignPDFRequest{
			PDFContent: "base64encodedpdf",
			CertPath:    "/path/to/cert.pem",
			KeyPath:     "/path/to/key.pem",
			UserID:      1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/pades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("missing PDF content", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/signature/pades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "pdf_content is required",
			})
		})

		body := SignPDFRequest{
			CertPath: "/path/to/cert.pem",
			KeyPath:  "/path/to/key.pem",
			UserID:   1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/pades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestSignatureHandler_VerifySignature(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSignatureService)

	app.Post("/api/v1/signature/verify", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"is_valid":          true,
			"signature_type":    "XAdES",
			"validation_score":  100.0,
			"processing_time_ms": 100,
		})
	})

	t.Run("verify XAdES signature - valid", func(t *testing.T) {
		signedContent := []byte(`<?xml version="1.0"?><document><ds:Signature>...</ds:Signature></document>`)
		expectedResult := &services.SignatureVerificationResult{
			IsValid:         true,
			SignatureType:    "XAdES",
			ValidationScore: 100.0,
		}

		mockService.On("VerifySignature", signedContent, "xades").Return(expectedResult, nil)

		body := VerifySignatureRequest{
			SignedContent: "base64encodedcontent",
			SignatureType: "xades",
			UserID:        1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/verify", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("verify PAdES signature - valid", func(t *testing.T) {
		signedContent := []byte("%PDF-1.4 signed content")
		expectedResult := &services.SignatureVerificationResult{
			IsValid:         true,
			SignatureType:    "PAdES",
			ValidationScore: 100.0,
		}

		mockService.On("VerifySignature", signedContent, "pades").Return(expectedResult, nil)

		body := VerifySignatureRequest{
			SignedContent: "base64encodedcontent",
			SignatureType: "pades",
			UserID:        1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/verify", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("verify signature - tampered", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/signature/verify", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"is_valid":          false,
				"signature_type":    "XAdES",
				"validation_score":  0.0,
				"errors":            []string{"signature_tampered"},
			})
		})

		body := VerifySignatureRequest{
			SignedContent: "base64encodedcontent",
			SignatureType: "xades",
			UserID:        1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/verify", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("verify signature - expired certificate", func(t *testing.T) {
		app2 := fiber.New()
		app2.Post("/api/v1/signature/verify", func(c *fiber.Ctx) error {
			return c.Status(200).JSON(fiber.Map{
				"is_valid":          false,
				"signature_type":    "XAdES",
				"validation_score":  0.0,
				"errors":            []string{"certificate_expired"},
			})
		})

		body := VerifySignatureRequest{
			SignedContent: "base64encodedcontent",
			SignatureType: "xades",
			UserID:        1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/verify", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app2.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestSignatureHandler_CertificateHandling(t *testing.T) {
	app := fiber.New()

	t.Run("certificate not found", func(t *testing.T) {
		app.Post("/api/v1/signature/xades", func(c *fiber.Ctx) error {
			return c.Status(404).JSON(fiber.Map{
				"error": "certificate file not found",
			})
		})

		body := SignXMLRequest{
			XMLContent: "<xml>test</xml>",
			CertPath:   "/nonexistent/cert.pem",
			KeyPath:    "/path/to/key.pem",
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("private key not found", func(t *testing.T) {
		app.Post("/api/v1/signature/pades", func(c *fiber.Ctx) error {
			return c.Status(404).JSON(fiber.Map{
				"error": "private key file not found",
			})
		})

		body := SignPDFRequest{
			PDFContent: "base64encodedpdf",
			CertPath:    "/path/to/cert.pem",
			KeyPath:     "/nonexistent/key.pem",
			UserID:      1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/pades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("expired certificate", func(t *testing.T) {
		app.Post("/api/v1/signature/xades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "certificate has expired",
			})
		})

		body := SignXMLRequest{
			XMLContent: "<xml>test</xml>",
			CertPath:   "/path/to/expired/cert.pem",
			KeyPath:    "/path/to/key.pem",
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("certificate not yet valid", func(t *testing.T) {
		app.Post("/api/v1/signature/pades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "certificate is not yet valid",
			})
		})

		body := SignPDFRequest{
			PDFContent: "base64encodedpdf",
			CertPath:    "/path/to/future/cert.pem",
			KeyPath:     "/path/to/key.pem",
			UserID:      1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/pades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestSignatureHandler_ErrorCases(t *testing.T) {
	app := fiber.New()

	t.Run("malformed XML", func(t *testing.T) {
		app.Post("/api/v1/signature/xades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "malformed XML content",
			})
		})

		body := SignXMLRequest{
			XMLContent: "<invalid>xml",
			CertPath:   "/path/to/cert.pem",
			KeyPath:    "/path/to/key.pem",
			UserID:     1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/xades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("malformed PDF", func(t *testing.T) {
		app.Post("/api/v1/signature/pades", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "malformed PDF content",
			})
		})

		body := SignPDFRequest{
			PDFContent: "not a valid pdf",
			CertPath:    "/path/to/cert.pem",
			KeyPath:     "/path/to/key.pem",
			UserID:      1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/pades", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("unsupported signature type", func(t *testing.T) {
		app.Post("/api/v1/signature/verify", func(c *fiber.Ctx) error {
			return c.Status(400).JSON(fiber.Map{
				"error": "unsupported signature type",
			})
		})

		body := VerifySignatureRequest{
			SignedContent: "base64encodedcontent",
			SignatureType: "unsupported",
			UserID:        1,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/signature/verify", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}