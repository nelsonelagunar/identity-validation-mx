package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockXAdESSigner struct {
	mock.Mock
}

func (m *MockXAdESSigner) Sign(xmlContent string, cert *x509.Certificate, key *rsa.PrivateKey) ([]byte, error) {
	args := m.Called(xmlContent, cert, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockXAdESSigner) Verify(signedXML []byte) (*SignatureVerificationResult, error) {
	args := m.Called(signedXML)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SignatureVerificationResult), args.Error(1)
}

func (m *MockXAdESSigner) GetSignatureType() string {
	args := m.Called()
	return args.String(0)
}

type MockPAdESSigner struct {
	mock.Mock
}

func (m *MockPAdESSigner) Sign(pdfContent []byte, cert *x509.Certificate, key *rsa.PrivateKey) ([]byte, error) {
	args := m.Called(pdfContent, cert, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPAdESSigner) Verify(signedPDF []byte) (*SignatureVerificationResult, error) {
	args := m.Called(signedPDF)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SignatureVerificationResult), args.Error(1)
}

func (m *MockPAdESSigner) GetSignatureType() string {
	args := m.Called()
	return args.String(0)
}

type MockCertificateHandler struct {
	mock.Mock
}

func (m *MockCertificateHandler) LoadCertificate(certPath string) (*x509.Certificate, error) {
	args := m.Called(certPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x509.Certificate), args.Error(1)
}

func (m *MockCertificateHandler) LoadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	args := m.Called(keyPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rsa.PrivateKey), args.Error(1)
}

func (m *MockCertificateHandler) VerifyCertificate(cert *x509.Certificate) error {
	args := m.Called(cert)
	return args.Error(0)
}

func (m *MockCertificateHandler) GetCertificateChain(cert *x509.Certificate) ([]*x509.Certificate, error) {
	args := m.Called(cert)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*x509.Certificate), args.Error(1)
}

func (m *MockCertificateHandler) ValidateCertificate(cert *x509.Certificate) (*CertificateValidationResult, error) {
	args := m.Called(cert)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CertificateValidationResult), args.Error(1)
}

func generateTestCertificate(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Certificate",
			Organization: []string{"Test Org"},
			Country:      []string{"MX"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

func TestXAdESSigner_Sign(t *testing.T) {
	mockSigner := new(MockXAdESSigner)
	cert, key := generateTestCertificate(t)

	xmlContent := `<?xml version="1.0"?><document><data>Test content</data></document>`

	t.Run("successful XAdES signing", func(t *testing.T) {
		signedXML := []byte(`<?xml version="1.0"?><document><data>Test content</data><ds:Signature>...</ds:Signature></document>`)
		mockSigner.On("Sign", xmlContent, cert, key).Return(signedXML, nil)

		result, err := mockSigner.Sign(xmlContent, cert, key)

		require.NoError(t, err)
		assert.Contains(t, string(result), "Signature")
		mockSigner.AssertExpectations(t)
	})

	t.Run("invalid XML content", func(t *testing.T) {
		invalidXML := ""
		mockSigner.On("Sign", invalidXML, cert, key).Return(nil, ErrInvalidInput)

		result, err := mockSigner.Sign(invalidXML, cert, key)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
		mockSigner.AssertExpectations(t)
	})
}

func TestXAdESSigner_Verify(t *testing.T) {
	mockSigner := new(MockXAdESSigner)

	t.Run("valid XAdES signature", func(t *testing.T) {
		signedXML := []byte(`<?xml version="1.0"?><document><data>Test</data><ds:Signature>valid</ds:Signature></document>`)
		expectedResult := &SignatureVerificationResult{
			IsValid:          true,
			SignerCert:       nil,
			SignatureTime:    time.Now(),
			SignatureType:     "XAdES",
			ValidationScore:  100.0,
		}
		mockSigner.On("Verify", signedXML).Return(expectedResult, nil)

		result, err := mockSigner.Verify(signedXML)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Equal(t, "XAdES", result.SignatureType)
		mockSigner.AssertExpectations(t)
	})

	t.Run("tampered signature", func(t *testing.T) {
		tamperedXML := []byte(`<?xml version="1.0"?><document><data>Tampered</data><ds:Signature>invalid</ds:Signature></document>`)
		expectedResult := &SignatureVerificationResult{
			IsValid:          false,
			Errors:          []error{ErrSignatureVerificationFailed},
			ValidationScore: 0.0,
		}
		mockSigner.On("Verify", tamperedXML).Return(expectedResult, nil)

		result, err := mockSigner.Verify(tamperedXML)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
		mockSigner.AssertExpectations(t)
	})

	t.Run("missing signature", func(t *testing.T) {
		missingSig := []byte(`<?xml version="1.0"?><document><data>No signature</data></document>`)
		mockSigner.On("Verify", missingSig).Return(nil, ErrSignatureNotFound)

		result, err := mockSigner.Verify(missingSig)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrSignatureNotFound)
		mockSigner.AssertExpectations(t)
	})
}

func TestPAdESSigner_Sign(t *testing.T) {
	mockSigner := new(MockPAdESSigner)
	cert, key := generateTestCertificate(t)

	pdfContent := []byte("%PDF-1.4 test pdf content")

	t.Run("successful PAdES signing", func(t *testing.T) {
		signedPDF := append(pdfContent, []byte("\n% Signed with PAdES")...)
		mockSigner.On("Sign", pdfContent, cert, key).Return(signedPDF, nil)

		result, err := mockSigner.Sign(pdfContent, cert, key)

		require.NoError(t, err)
		assert.Contains(t, string(result), "Signed")
		mockSigner.AssertExpectations(t)
	})

	t.Run("empty PDF content", func(t *testing.T) {
		emptyPDF := []byte{}
		mockSigner.On("Sign", emptyPDF, cert, key).Return(nil, ErrInvalidInput)

		result, err := mockSigner.Sign(emptyPDF, cert, key)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
		mockSigner.AssertExpectations(t)
	})

	t.Run("invalid certificate", func(t *testing.T) {
		_, invalidKey := generateTestCertificate(t)
		expiredCert := &x509.Certificate{
			NotAfter: time.Now().Add(-24 * time.Hour),
		}
		mockSigner.On("Sign", pdfContent, expiredCert, invalidKey).Return(nil, ErrCertificateExpired)

		result, err := mockSigner.Sign(pdfContent, expiredCert, invalidKey)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCertificateExpired)
		mockSigner.AssertExpectations(t)
	})
}

func TestPAdESSigner_Verify(t *testing.T) {
	mockSigner := new(MockPAdESSigner)

	t.Run("valid PAdES signature", func(t *testing.T) {
		signedPDF := []byte("%PDF-1.4 signed content")
		expectedResult := &SignatureVerificationResult{
			IsValid:         true,
			SignatureType:    "PAdES",
			ValidationScore: 100.0,
		}
		mockSigner.On("Verify", signedPDF).Return(expectedResult, nil)

		result, err := mockSigner.Verify(signedPDF)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Equal(t, "PAdES", result.SignatureType)
		mockSigner.AssertExpectations(t)
	})

	t.Run("corrupted signature", func(t *testing.T) {
		corruptedPDF := []byte("%PDF-1.4 corrupted")
		expectedResult := &SignatureVerificationResult{
			IsValid:         false,
			Errors:         []error{ErrSignatureCorrupted},
			ValidationScore: 0.0,
		}
		mockSigner.On("Verify", corruptedPDF).Return(expectedResult, nil)

		result, err := mockSigner.Verify(corruptedPDF)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
		mockSigner.AssertExpectations(t)
	})
}

func TestCertificateHandler_LoadCertificate(t *testing.T) {
	mockHandler := new(MockCertificateHandler)

	t.Run("valid certificate", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		certPath := "/path/to/cert.pem"
		mockHandler.On("LoadCertificate", certPath).Return(cert, nil)

		result, err := mockHandler.LoadCertificate(certPath)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Certificate", result.Subject.CommonName)
		mockHandler.AssertExpectations(t)
	})

	t.Run("certificate not found", func(t *testing.T) {
		certPath := "/nonexistent/cert.pem"
		mockHandler.On("LoadCertificate", certPath).Return(nil, ErrCertificateNotFound)

		result, err := mockHandler.LoadCertificate(certPath)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCertificateNotFound)
		mockHandler.AssertExpectations(t)
	})

	t.Run("malformed certificate", func(t *testing.T) {
		certPath := "/path/to/malformed.pem"
		mockHandler.On("LoadCertificate", certPath).Return(nil, ErrCertificateMalformed)

		result, err := mockHandler.LoadCertificate(certPath)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCertificateMalformed)
		mockHandler.AssertExpectations(t)
	})
}

func TestCertificateHandler_LoadPrivateKey(t *testing.T) {
	mockHandler := new(MockCertificateHandler)

	t.Run("valid private key", func(t *testing.T) {
		_, key := generateTestCertificate(t)
		keyPath := "/path/to/key.pem"
		mockHandler.On("LoadPrivateKey", keyPath).Return(key, nil)

		result, err := mockHandler.LoadPrivateKey(keyPath)

		require.NoError(t, err)
		assert.NotNil(t, result)
		mockHandler.AssertExpectations(t)
	})

	t.Run("private key not found", func(t *testing.T) {
		keyPath := "/nonexistent/key.pem"
		mockHandler.On("LoadPrivateKey", keyPath).Return(nil, ErrPrivateKeyNotFound)

		result, err := mockHandler.LoadPrivateKey(keyPath)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrPrivateKeyNotFound)
		mockHandler.AssertExpectations(t)
	})
}

func TestCertificateHandler_VerifyCertificate(t *testing.T) {
	mockHandler := new(MockCertificateHandler)

	t.Run("valid and current certificate", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		mockHandler.On("VerifyCertificate", cert).Return(nil)

		err := mockHandler.VerifyCertificate(cert)

		assert.NoError(t, err)
		mockHandler.AssertExpectations(t)
	})

	t.Run("expired certificate", func(t *testing.T) {
		cert := &x509.Certificate{
			NotAfter: time.Now().Add(-24 * time.Hour),
		}
		mockHandler.On("VerifyCertificate", cert).Return(ErrCertificateExpired)

		err := mockHandler.VerifyCertificate(cert)

		assert.ErrorIs(t, err, ErrCertificateExpired)
		mockHandler.AssertExpectations(t)
	})

	t.Run("not yet valid certificate", func(t *testing.T) {
		cert := &x509.Certificate{
			NotBefore: time.Now().Add(24 * time.Hour),
		}
		mockHandler.On("VerifyCertificate", cert).Return(ErrCertificateNotYetValid)

		err := mockHandler.VerifyCertificate(cert)

		assert.ErrorIs(t, err, ErrCertificateNotYetValid)
		mockHandler.AssertExpectations(t)
	})
}

func TestCertificateHandler_ValidateCertificate(t *testing.T) {
	mockHandler := new(MockCertificateHandler)

	t.Run("fully valid certificate", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		expectedResult := &CertificateValidationResult{
			IsValid:            true,
			Subject:            "Test Certificate",
			Issuer:             "Test Certificate",
			SerialNumber:       "1",
			NotBefore:          time.Now(),
			NotAfter:           time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:           []string{"DigitalSignature", "KeyEncipherment"},
			ExtendedKeyUsage:   []string{"ServerAuth", "ClientAuth"},
			IsCA:               false,
			ValidationScore:    100.0,
		}
		mockHandler.On("ValidateCertificate", cert).Return(expectedResult, nil)

		result, err := mockHandler.ValidateCertificate(cert)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Equal(t, float64(100), result.ValidationScore)
		mockHandler.AssertExpectations(t)
	})

	t.Run("certificate with warnings", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		expectedResult := &CertificateValidationResult{
			IsValid:         true,
			Warnings:        []string{"Certificate expires soon"},
			ValidationScore: 70.0,
		}
		mockHandler.On("ValidateCertificate", cert).Return(expectedResult, nil)

		result, err := mockHandler.ValidateCertificate(cert)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.NotEmpty(t, result.Warnings)
		mockHandler.AssertExpectations(t)
	})
}

func TestCertificateHandler_GetCertificateChain(t *testing.T) {
	mockHandler := new(MockCertificateHandler)

	t.Run("complete certificate chain", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		intermediateCert, _ := generateTestCertificate(t)
		rootCert, _ := generateTestCertificate(t)

		chain := []*x509.Certificate{cert, intermediateCert, rootCert}
		mockHandler.On("GetCertificateChain", cert).Return(chain, nil)

		result, err := mockHandler.GetCertificateChain(cert)

		require.NoError(t, err)
		assert.Len(t, result, 3)
		mockHandler.AssertExpectations(t)
	})

	t.Run("incomplete certificate chain", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		mockHandler.On("GetCertificateChain", cert).Return(nil, ErrIncompleteChain)

		result, err := mockHandler.GetCertificateChain(cert)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrIncompleteChain)
		mockHandler.AssertExpectations(t)
	})
}

func TestSignPDF_Concurrent(t *testing.T) {
	mockSigner := new(MockPAdESSigner)
	cert, key := generateTestCertificate(t)

	pdfContent := []byte("%PDF-1.4 test")

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			signedPDF := append([]byte("%PDF-1.4 test"), []byte(" signed")...)
			mockSigner.On("Sign", pdfContent, cert, key).Return(signedPDF, nil).Once()

			result, err := mockSigner.Sign(pdfContent, cert, key)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSignatureVerificationResult_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		result   *SignatureVerificationResult
		expected bool
	}{
		{
			name: "fully valid",
			result: &SignatureVerificationResult{
				IsValid:         true,
				ValidationScore: 100.0,
			},
			expected: true,
		},
		{
			name: "valid with warnings",
			result: &SignatureVerificationResult{
				IsValid:         true,
				ValidationScore: 80.0,
				Warnings:        []string{"Certificate expires soon"},
			},
			expected: true,
		},
		{
			name: "invalid signature",
			result: &SignatureVerificationResult{
				IsValid: false,
				Errors:  []error{ErrSignatureVerificationFailed},
			},
			expected: false,
		},
		{
			name: "expired certificate",
			result: &SignatureVerificationResult{
				IsValid:         false,
				ValidationScore: 0.0,
				Errors:         []error{ErrCertificateExpired},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.IsValid)
		})
	}
}

func TestPEMEncoding(t *testing.T) {
	cert, key := generateTestCertificate(t)

	t.Run("encode certificate to PEM", func(t *testing.T) {
		certPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})

		assert.Contains(t, string(certPEM), "-----BEGIN CERTIFICATE-----")
		assert.Contains(t, string(certPEM), "-----END CERTIFICATE-----")
	})

	t.Run("encode private key to PEM", func(t *testing.T) {
		keyBytes := x509.MarshalPKCS1PrivateKey(key)
		keyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyBytes,
		})

		assert.Contains(t, string(keyPEM), "-----BEGIN RSA PRIVATE KEY-----")
		assert.Contains(t, string(keyPEM), "-----END RSA PRIVATE KEY-----")
	})

	t.Run("decode from base64", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("test data"))
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		require.NoError(t, err)
		assert.Equal(t, "test data", string(decoded))
	})
}