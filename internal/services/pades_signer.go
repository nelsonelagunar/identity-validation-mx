package services

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"
)

type PAdESSigner struct {
	certHandler     *CertificateHandler
	certificate     *x509.Certificate
	privateKey      *rsa.PrivateKey
	timestampServer string
}

type PAdESSignature struct {
	SignatureBytes  []byte
	Certificate     *x509.Certificate
	SignerName      string
	SignerRFCCURP   string
	Reason          string
	Location        string
	Timestamp       time.Time
	DocumentHash    []byte
	VisibleBounds   SignatureBounds
}

type SignatureBounds struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Page   int
}

type PDFSignatureContainer struct {
	ByteRange      [2]int
	Contents       []byte
	SubFilter      string
	SignatureType  string
	M              string
	SignerInfo     SignerInfoBlock
}

type SignerInfoBlock struct {
	Name        string
	Location    string
	Reason      string
	SigningTime time.Time
}

type PDFSignatureField struct {
	Name      string
	Page      int
	Rect      [4]float64
	Visible   bool
	Appearances PDFAppearance
}

type PDFAppearance struct {
	Normal []byte
}

type TimestampToken struct {
	Token     []byte
	Timestamp time.Time
	Hash      []byte
}

func NewPAdESSigner(certHandler *CertificateHandler) (*PAdESSigner, error) {
	if certHandler == nil {
		return nil, fmt.Errorf("certificate handler is required")
	}

	return &PAdESSigner{
		certHandler:     certHandler,
		certificate:     certHandler.GetCertificate(),
		privateKey:      certHandler.GetPrivateKey(),
		timestampServer: "http://timestamp.digicert.com",
	}, nil
}

func (p *PAdESSigner) SignBES(ctx context.Context, pdfDocument []byte, opts *SignOptions) (*SignResult, error) {
	signingTime := time.Now().UTC()

	documentHash := sha256.Sum256(pdfDocument)
	documentHashBase64 := base64.StdEncoding.EncodeToString(documentHash[:])

	signerName := "Unknown"
	signerRFC := ""
	if opts != nil {
		if opts.SignerName != "" {
			signerName = opts.SignerName
		}
		if opts.SignerRFCCURP != "" {
			signerRFC = opts.SignerRFCCURP
		}
	}

	signerInfo := SignerInfoBlock{
		Name:        signerName,
		Location:    getLocationString(opts),
		Reason:      getReasonString(opts),
		SigningTime: signingTime,
	}

	signatureData := p.buildSignatureData(pdfDocument, signerInfo)

	hashToSign := sha256.Sum256(signatureData)
	signatureBytes, err := rsa.SignPKCS1v15(nil, p.privateKey, crypto.SHA256, hashToSign[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign PDF: %w", err)
	}

	signatureContainer := p.buildSignatureContainer(signatureBytes, signerInfo, documentHash[:])

	signatureBase64 := base64.StdEncoding.EncodeToString(signatureContainer)

	var timestampToken string
	if opts != nil && opts.AddTimestamp {
		timestamp, err := p.addTimestamp(ctx, signatureContainer)
		if err != nil {
			return nil, fmt.Errorf("failed to add timestamp: %w", err)
		}
		timestampToken = timestamp
	}

	return &SignResult{
		Signature:       string(signatureContainer),
		SignatureBase64: signatureBase64,
		Certificate:     base64.StdEncoding.EncodeToString(p.certificate.Raw),
		SerialNumber:    fmt.Sprintf("%d", p.certificate.SerialNumber),
		IssuerDN:        p.certificate.Issuer.String(),
		SubjectDN:       p.certificate.Subject.String(),
		ValidFrom:       p.certificate.NotBefore,
		ValidTo:         p.certificate.NotAfter,
		DocumentHash:     documentHashBase64,
		Timestamp:        signingTime,
		TimestampToken:   timestampToken,
	}, nil
}

func (p *PAdESSigner) Verify(ctx context.Context, pdfDocument []byte, signature []byte) (*VerifyResult, error) {
	result := &VerifyResult{
		IsValid:           false,
		SignerVerified:    false,
		DocumentIntegrity: false,
		TimestampValid:    false,
		CertificateValid:  false,
	}

	signatureContainer, err := p.parseSignatureContainer(signature)
	if err != nil {
		result.ErrorCode = "PARSE_ERROR"
		result.ErrorMessage = fmt.Sprintf("Failed to parse PAdES signature: %v", err)
		return result, nil
	}

	cert := p.certificate

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		result.ErrorCode = "CERT_EXPIRED"
		result.ErrorMessage = "Certificate is not valid at current time"
		return result, nil
	}

	result.CertificateValid = true

	documentHash := sha256.Sum256(pdfDocument)
	documentHashBase64 := base64.StdEncoding.EncodeToString(documentHash[:])

	storedHash := base64.StdEncoding.EncodeToString(signatureContainer.DocumentHash)
	if storedHash != documentHashBase64 {
		result.ErrorCode = "DOCUMENT_INTEGRITY_FAILED"
		result.ErrorMessage = "Document hash does not match signature"
		return result, nil
	}

	result.DocumentIntegrity = true

	hashOfSignatureData := sha256.Sum256(p.buildSignatureData(pdfDocument, signatureContainer.SignerInfo))

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		result.ErrorCode = "KEY_TYPE_ERROR"
		result.ErrorMessage = "Signer certificate does not contain RSA public key"
		return result, nil
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashOfSignatureData[:], signatureContainer.SignatureBytes)
	if err != nil {
		result.ErrorCode = "SIGNATURE_INVALID"
		result.ErrorMessage = "Signature verification failed"
		return result, nil
	}

	result.IsValid = true
	result.SignerVerified = true
	result.VerificationDetails = "PAdES signature verified successfully"

	return result, nil
}

func (p *PAdESSigner) buildSignatureData(pdfDocument []byte, signerInfo SignerInfoBlock) []byte {
	data := fmt.Sprintf("PDF Signature\nDocument: %d bytes\nSigner: %s\nReason: %s\nLocation: %s\nTime: %s",
		len(pdfDocument),
		signerInfo.Name,
		signerInfo.Reason,
		signerInfo.Location,
		signerInfo.SigningTime.Format(time.RFC3339),
	)
	return []byte(data)
}

func (p *PAdESSigner) buildSignatureContainer(signatureBytes []byte, signerInfo SignerInfoBlock, documentHash []byte) []byte {
	container := fmt.Sprintf(
		"------BEGIN PAdES SIGNATURE------\n"+
			"Version: 1.0\n"+
			"Signer: %s\n"+
			"Reason: %s\n"+
			"Location: %s\n"+
			"SigningTime: %s\n"+
			"DocumentHash: %s\n"+
			"Signature: %s\n"+
			"Certificate: %s\n"+
			"------END PAdES SIGNATURE------",
		signerInfo.Name,
		signerInfo.Reason,
		signerInfo.Location,
		signerInfo.SigningTime.Format(time.RFC3339),
		base64.StdEncoding.EncodeToString(documentHash),
		base64.StdEncoding.EncodeToString(signatureBytes),
		base64.StdEncoding.EncodeToString(p.certificate.Raw),
	)
	return []byte(container)
}

func (p *PAdESSigner) parseSignatureContainer(signature []byte) (*PDFSignatureContainer, error) {
	container := &PDFSignatureContainer{
		SubFilter:     "adbe.pkcs7.detached",
		SignatureType: "PAdES-BES",
		SignerInfo: SignerInfoBlock{
			SigningTime: time.Now(),
		},
	}

	parts := splitSignatureContainer(string(signature))
	if len(parts) >= 8 {
		container.SignerInfo.Name = parts["Signer"]
		container.SignerInfo.Reason = parts["Reason"]
		container.SignerInfo.Location = parts["Location"]
		
		if signingTime, err := time.Parse(time.RFC3339, parts["SigningTime"]); err == nil {
			container.SignerInfo.SigningTime = signingTime
		}
		
		if sigBytes, err := base64.StdEncoding.DecodeString(parts["Signature"]); err == nil {
			container.Contents = sigBytes
		}
		
		if docHash, err := base64.StdEncoding.DecodeString(parts["DocumentHash"]); err == nil {
			container.ByteRange[0] = len(docHash)
		}
	}

	return container, nil
}

func (p *PAdESSigner) addTimestamp(ctx context.Context, signatureContainer []byte) (string, error) {
	now := time.Now().UTC()
	timestampToken := fmt.Sprintf("Timestamp: %s\nData: %s",
		now.Format(time.RFC3339),
		base64.StdEncoding.EncodeToString(signatureContainer),
	)
	
	hash := sha256.Sum256([]byte(timestampToken))
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

func splitSignatureContainer(signature string) map[string]string {
	parts := make(map[string]string)
	lines := splitLines(signature)
	
	for _, line := range lines {
		idx := indexOf(line, ":")
		if idx > 0 {
			key := trim(line[:idx])
			value := trim(line[idx+1:])
			parts[key] = value
		}
	}
	
	return parts
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func indexOf(s string, c string) int {
	for i := 0; i < len(s); i++ {
		if s[i:i+1] == c {
			return i
		}
	}
	return -1
}

func trim(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func getReasonString(opts *SignOptions) string {
	if opts == nil {
		return "Digital Signature"
	}
	if opts.Reason != "" {
		return opts.Reason
	}
	return "Digital Signature"
}

func getLocationString(opts *SignOptions) string {
	if opts == nil {
		return ""
	}
	return opts.Location
}