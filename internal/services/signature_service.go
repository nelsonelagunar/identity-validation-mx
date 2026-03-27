package services

import (
	"context"
	"time"

	"github.com/nlaguna/identity-validation-mx/internal/models"
)

type SignatureType string

const (
	SignatureTypeXAdES SignatureType = "xades"
	SignatureTypePAdES SignatureType = "pades"
)

type XAdESLevel string

const (
	XAdESLevelBES XAdESLevel = "xades-bes"
	XAdESLevelT   XAdESLevel = "xades-t"
)

type PAdESLevel string

const (
	PAdESLevelBES PAdESLevel = "pades-bes"
)

type SignOptions struct {
	SignatureType    SignatureType
	XAdESLevel      XAdESLevel
	PAdESLevel      PAdESLevel
	AddTimestamp    bool
	VisibleSignature bool
	SignerName      string
	SignerRFCCURP   string
	Reason          string
	Location        string
}

type SignResult struct {
	Signature        string    `json:"signature"`
	SignatureBase64  string    `json:"signature_base64"`
	Certificate      string    `json:"certificate"`
	SerialNumber     string    `json:"serial_number"`
	IssuerDN        string    `json:"issuer_dn"`
	SubjectDN       string    `json:"subject_dn"`
	ValidFrom       time.Time `json:"valid_from"`
	ValidTo         time.Time `json:"valid_to"`
	DocumentHash    string    `json:"document_hash"`
	Timestamp       time.Time `json:"timestamp"`
	TimestampToken   string    `json:"timestamp_token"`
}

type VerifyOptions struct {
	DocumentHash      string
	Signature         string
	CheckTimestamp    bool
	CheckCertificate  bool
}

type VerifyResult struct {
	IsValid           bool   `json:"is_valid"`
	SignerVerified    bool   `json:"signer_verified"`
	DocumentIntegrity bool   `json:"document_integrity"`
	TimestampValid    bool   `json:"timestamp_valid"`
	CertificateValid  bool   `json:"certificate_valid"`
	ErrorCode        string `json:"error_code"`
	ErrorMessage     string `json:"error_message"`
	VerificationDetails string `json:"verification_details"`
}

type SignatureService interface {
	Sign(ctx context.Context, document []byte, opts *SignOptions) (*SignResult, error)
	Verify(ctx context.Context, document []byte, signature []byte, opts *VerifyOptions) (*VerifyResult, error)
	SignXAdES(ctx context.Context, xmlDocument []byte, opts *SignOptions) (*SignResult, error)
	SignPAdES(ctx context.Context, pdfDocument []byte, opts *SignOptions) (*SignResult, error)
	VerifyXAdES(ctx context.Context, xmlDocument []byte, signature []byte) (*VerifyResult, error)
	VerifyPAdES(ctx context.Context, pdfDocument []byte, signature []byte) (*VerifyResult, error)
}

type signatureService struct {
	xadesSigner         *XAdESSigner
	padesSigner         *PAdESSigner
	certificateHandler  *CertificateHandler
}

func NewSignatureService(certHandler *CertificateHandler) (SignatureService, error) {
	xadesSigner, err := NewXAdESSigner(certHandler)
	if err != nil {
		return nil, err
	}

	padesSigner, err := NewPAdESSigner(certHandler)
	if err != nil {
		return nil, err
	}

	return &signatureService{
		xadesSigner:        xadesSigner,
		padesSigner:        padesSigner,
		certificateHandler: certHandler,
	}, nil
}

func (s *signatureService) Sign(ctx context.Context, document []byte, opts *SignOptions) (*SignResult, error) {
	if opts == nil {
		opts = &SignOptions{
			SignatureType: SignatureTypeXAdES,
			XAdESLevel:    XAdESLevelBES,
			AddTimestamp:  false,
		}
	}

	switch opts.SignatureType {
	case SignatureTypeXAdES:
		return s.SignXAdES(ctx, document, opts)
	case SignatureTypePAdES:
		return s.SignPAdES(ctx, document, opts)
	default:
		return s.SignXAdES(ctx, document, opts)
	}
}

func (s *signatureService) Verify(ctx context.Context, document []byte, signature []byte, opts *VerifyOptions) (*VerifyResult, error) {
	if opts == nil {
		opts = &VerifyOptions{
			CheckTimestamp:   true,
			CheckCertificate: true,
		}
	}

	signatureType := detectSignatureType(signature)

	switch signatureType {
	case SignatureTypeXAdES:
		return s.VerifyXAdES(ctx, document, signature)
	case SignatureTypePAdES:
		return s.VerifyPAdES(ctx, document, signature)
	default:
		return s.VerifyXAdES(ctx, document, signature)
	}
}

func (s *signatureService) SignXAdES(ctx context.Context, xmlDocument []byte, opts *SignOptions) (*SignResult, error) {
	if opts == nil {
		opts = &SignOptions{
			XAdESLevel:   XAdESLevelBES,
			AddTimestamp: false,
		}
	}

	switch opts.XAdESLevel {
	case XAdESLevelBES:
		return s.xadesSigner.SignBES(ctx, xmlDocument, opts)
	case XAdESLevelT:
		return s.xadesSigner.SignT(ctx, xmlDocument, opts)
	default:
		return s.xadesSigner.SignBES(ctx, xmlDocument, opts)
	}
}

func (s *signatureService) SignPAdES(ctx context.Context, pdfDocument []byte, opts *SignOptions) (*SignResult, error) {
	if opts == nil {
		opts = &SignOptions{
			PAdESLevel: PAdESLevelBES,
			AddTimestamp: false,
		}
	}

	return s.padesSigner.SignBES(ctx, pdfDocument, opts)
}

func (s *signatureService) VerifyXAdES(ctx context.Context, xmlDocument []byte, signature []byte) (*VerifyResult, error) {
	return s.xadesSigner.Verify(ctx, xmlDocument, signature)
}

func (s *signatureService) VerifyPAdES(ctx context.Context, pdfDocument []byte, signature []byte) (*VerifyResult, error) {
	return s.padesSigner.Verify(ctx, pdfDocument, signature)
}

func detectSignatureType(signature []byte) SignatureType {
	if isXMLSignature(signature) {
		return SignatureTypeXAdES
	}
	if isPDFSignature(signature) {
		return SignatureTypePAdES
	}
	return SignatureTypeXAdES
}

func isXMLSignature(data []byte) bool {
	if len(data) < 5 {
		return false
	}
	return data[0] == '<' && data[1] == '?'
}

func isPDFSignature(data []byte) bool {
	if len(data) < 5 {
		return false
	}
	return data[0] == '%' && data[1] == 'P' && data[2] == 'D' && data[3] == 'F'
}

func NewSignatureRequest(userID uint, documentHash, signerName, signerRFCCURP, signatureType string) *models.DigitalSignatureRequest {
	return &models.DigitalSignatureRequest{
		UserID:        userID,
		DocumentHash:  documentHash,
		SignerName:    signerName,
		SignerRFCCURP: signerRFCCURP,
		SignatureType: signatureType,
		Status:        "pending",
	}
}

func NewVerificationRequest(signatureID uint, documentHash, signature string) *models.SignatureVerificationRequest {
	return &models.SignatureVerificationRequest{
		SignatureID:  signatureID,
		DocumentHash: documentHash,
		Signature:    signature,
		Status:       "pending",
	}
}