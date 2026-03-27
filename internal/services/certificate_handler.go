package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

type CertificateHandler struct {
	certificate     *x509.Certificate
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	certChain       []*x509.Certificate
	caCertPool      *x509.CertPool
}

type CertificateInfo struct {
	SubjectDN       string    `json:"subject_dn"`
	IssuerDN        string    `json:"issuer_dn"`
	SerialNumber    string    `json:"serial_number"`
	ValidFrom       time.Time `json:"valid_from"`
	ValidTo         time.Time `json:"valid_to"`
	IsValid        bool      `json:"is_valid"`
	IsValidChain   bool      `json:"is_valid_chain"`
	IsValidForUse  bool      `json:"is_valid_for_use"`
	KeyUsage       []string  `json:"key_usage"`
	ExtKeyUsage    []string  `json:"ext_key_usage"`
	PublicKeyType  string    `json:"public_key_type"`
	PublicKeySize  int       `json:"public_key_size"`
}

type PKCS12Content struct {
	Certificate  *x509.Certificate
	PrivateKey   *rsa.PrivateKey
	Certificates []*x509.Certificate
}

func NewCertificateHandler() *CertificateHandler {
	return &CertificateHandler{
		caCertPool: x509.NewCertPool(),
	}
}

func (c *CertificateHandler) LoadPKCS12FromFile(filename string, password []byte) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read PKCS12 file: %w", err)
	}
	return c.LoadPKCS12(data, password)
}

func (c *CertificateHandler) LoadPKCS12(data []byte, password []byte) error {
	blocks := c.parsePKCS12Blocks(data, password)
	
	if len(blocks.Certificates) == 0 {
		return fmt.Errorf("no certificates found in PKCS12 data")
	}

	c.certificate = blocks.Certificate
	if blocks.PrivateKey != nil {
		c.privateKey = blocks.PrivateKey
		c.publicKey = &blocks.PrivateKey.PublicKey
	}
	c.certChain = blocks.Certificates

	return nil
}

func (c *CertificateHandler) parsePKCS12Blocks(data []byte, password []byte) *PKCS12Content {
	content := &PKCS12Content{}
	
	var pemBlocks []string
	currentLine := ""
	
	for _, b := range string(data) {
		if b == '\n' || b == '\r' {
			if currentLine != "" {
				pemBlocks = append(pemBlocks, currentLine)
			}
			currentLine = ""
		} else {
			currentLine += string(b)
		}
	}
	if currentLine != "" {
		pemBlocks = append(pemBlocks, currentLine)
	}

	for i, line := range pemBlocks {
		if len(line) >= 27 && line[:27] == "-----BEGIN CERTIFICATE-----" {
			endIdx := -1
			for j := i; j < len(pemBlocks); j++ {
				if len(pemBlocks[j]) >= 25 && pemBlocks[j][:25] == "-----END CERTIFICATE-----" {
					endIdx = j
					break
				}
			}
			if endIdx != -1 {
				certPEM := ""
				for k := i; k <= endIdx; k++ {
					certPEM += pemBlocks[k] + "\n"
				}
				block, _ := pem.Decode([]byte(certPEM))
				if block != nil {
					cert, err := x509.ParseCertificate(block.Bytes)
					if err == nil {
						if content.Certificate == nil {
							content.Certificate = cert
						}
						content.Certificates = append(content.Certificates, cert)
					}
				}
			}
		}
		if len(line) >= 28 && line[:28] == "-----BEGIN PRIVATE KEY-----" {
			endIdx := -1
			for j := i; j < len(pemBlocks); j++ {
				if len(pemBlocks[j]) >= 26 && pemBlocks[j][:26] == "-----END PRIVATE KEY-----" {
					endIdx = j
					break
				}
			}
			if endIdx != -1 {
				keyPEM := ""
				for k := i; k <= endIdx; k++ {
					keyPEM += pemBlocks[k] + "\n"
				}
				block, _ := pem.Decode([]byte(keyPEM))
				if block != nil {
					key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
					if err == nil {
						content.PrivateKey = key
					}
				}
			}
		}
		if len(line) >= 25 && line[:25] == "-----BEGIN RSA PRIVATE KEY-----" {
			endIdx := -1
			for j := i; j < len(pemBlocks); j++ {
				if len(pemBlocks[j]) >= 23 && pemBlocks[j][:23] == "-----END RSA PRIVATE KEY-----" {
					endIdx = j
					break
				}
			}
			if endIdx != -1 {
				keyPEM := ""
				for k := i; k <= endIdx; k++ {
					keyPEM += pemBlocks[k] + "\n"
				}
				block, _ := pem.Decode([]byte(keyPEM))
				if block != nil {
					key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
					if err == nil {
						content.PrivateKey = key
					}
				}
			}
		}
	}
	
	return content
}

func (c *CertificateHandler) LoadCertificateFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}
	return c.LoadCertificate(data)
}

func (c *CertificateHandler) LoadCertificate(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	c.certificate = cert
	if cert.PublicKey != nil {
		if pubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			c.publicKey = pubKey
		}
	}

	return nil
}

func (c *CertificateHandler) LoadPrivateKeyFromFile(filename string, password []byte) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}
	return c.LoadPrivateKey(data, password)
}

func (c *CertificateHandler) LoadPrivateKey(data []byte, password []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	var key *rsa.PrivateKey
	var err error

	if block.Type == "RSA PRIVATE KEY" {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else if block.Type == "PRIVATE KEY" {
		var parsedKey any
		parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			key, ok := parsedKey.(*rsa.PrivateKey)
			if !ok {
				return fmt.Errorf("private key is not RSA")
			}
		}
	} else if block.Type == "ENCRYPTED PRIVATE KEY" && len(password) > 0 {
		var parsedKey any
		parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			key, ok := parsedKey.(*rsa.PrivateKey)
			if !ok {
				return fmt.Errorf("private key is not RSA")
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	c.privateKey = key
	c.publicKey = &key.PublicKey

	return nil
}

func (c *CertificateHandler) GetCertificate() *x509.Certificate {
	return c.certificate
}

func (c *CertificateHandler) GetPrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

func (c *CertificateHandler) GetPublicKey() *rsa.PublicKey {
	return c.publicKey
}

func (c *CertificateHandler) GetCertificateChain() []*x509.Certificate {
	return c.certChain
}

func (c *CertificateHandler) ExtractPublicKey(cert *x509.Certificate) (*rsa.PublicKey, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate is nil")
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate does not contain RSA public key")
	}

	return pubKey, nil
}

func (c *CertificateHandler) VerifyCertificateChain(cert *x509.Certificate) error {
	if len(c.certChain) == 0 {
		return fmt.Errorf("no certificate chain available")
	}

	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	var selfSigned bool
	for _, chainCert := range c.certChain {
		if chainCert.Subject.String() == chainCert.Issuer.String() {
			roots.AddCert(chainCert)
			selfSigned = true
		} else {
			intermediates.AddCert(chainCert)
		}
	}

	if !selfSigned {
		return fmt.Errorf("no root certificate found in chain")
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	_, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}

	return nil
}

func (c *CertificateHandler) CheckCertificateValidity(cert *x509.Certificate, now time.Time) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from %s)", cert.NotBefore.Format(time.RFC3339))
	}

	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (expired at %s)", cert.NotAfter.Format(time.RFC3339))
	}

	return nil
}

func (c *CertificateHandler) GetCertificateInfo(cert *x509.Certificate) (*CertificateInfo, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate is nil")
	}

	now := time.Now()
	isValid := cert.NotBefore.Before(now) && cert.NotAfter.After(now)

	keyUsage := getKeyUsageStrings(cert.KeyUsage)
	extKeyUsage := getExtKeyUsageStrings(cert.ExtKeyUsage)

	publicKeyType := "unknown"
	publicKeySize := 0
	if pubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		publicKeyType = "RSA"
		publicKeySize = pubKey.N.BitLen()
	}

	isValidChain := true
	if err := c.VerifyCertificateChain(cert); err != nil {
		isValidChain = false
	}

	isValidForUse := isValid && isValidChain

	return &CertificateInfo{
		SubjectDN:      cert.Subject.String(),
		IssuerDN:       cert.Issuer.String(),
		SerialNumber:   cert.SerialNumber.String(),
		ValidFrom:      cert.NotBefore,
		ValidTo:        cert.NotAfter,
		IsValid:       isValid,
		IsValidChain:  isValidChain,
		IsValidForUse: isValidForUse,
		KeyUsage:      keyUsage,
		ExtKeyUsage:   extKeyUsage,
		PublicKeyType: publicKeyType,
		PublicKeySize: publicKeySize,
	}, nil
}

func getKeyUsageStrings(ku x509.KeyUsage) []string {
	var usages []string
	if ku&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "DigitalSignature")
	}
	if ku&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "ContentCommitment")
	}
	if ku&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "KeyEncipherment")
	}
	if ku&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "DataEncipherment")
	}
	if ku&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "KeyAgreement")
	}
	if ku&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "CertSign")
	}
	if ku&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRLSign")
	}
	if ku&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "EncipherOnly")
	}
	if ku&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "DecipherOnly")
	}
	return usages
}

func getExtKeyUsageStrings(eku []x509.ExtKeyUsage) []string {
	var usages []string
	for _, usage := range eku {
		switch usage {
		case x509.ExtKeyUsageAny:
			usages = append(usages, "Any")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "ServerAuth")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "ClientAuth")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "CodeSigning")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "EmailProtection")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "TimeStamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSPSigning")
		default:
			usages = append(usages, fmt.Sprintf("ExtKeyUsage(%d)", usage))
		}
	}
	return usages
}

func GenerateSelfSignedCertificate(commonName string, days int) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, days),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, privateKey, nil
}

func CertificateToPEM(cert *x509.Certificate) string {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	return string(pem.EncodeToMemory(pemBlock))
}

func PrivateKeyToPEM(key *rsa.PrivateKey) string {
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return string(pem.EncodeToMemory(pemBlock))
}

func PublicKeyToPEM(pubKey *rsa.PublicKey) string {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return ""
	}
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}
	return string(pem.EncodeToMemory(pemBlock))
}

func CertificateToBase64(cert *x509.Certificate) string {
	return base64.StdEncoding.EncodeToString(cert.Raw)
}

func PrivateKeyToBase64(key *rsa.PrivateKey) string {
	return base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
}

func VerifySignatureWithCertificate(cert *x509.Certificate, publicKey *rsa.PublicKey, data []byte, signature []byte) error {
	hash := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
}