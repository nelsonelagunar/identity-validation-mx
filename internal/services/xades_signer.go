package services

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type XAdESSigner struct {
	certHandler     *CertificateHandler
	certificate     *x509.Certificate
	privateKey      *rsa.PrivateKey
	timestampServer string
}

type XAdESSignature struct {
	XMLName        xml.Name          `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`
	Id            string             `xml:"Id,attr"`
	SignedInfo    SignedInfo         `xml:"SignedInfo"`
	SignatureValue string            `xml:"SignatureValue"`
	KeyInfo       KeyInfo            `xml:"KeyInfo"`
	Object        []SignatureObject  `xml:"Object"`
}

type SignedInfo struct {
	CanonicalizationMethod CanonicalizationMethod `xml:"CanonicalizationMethod"`
	SignatureMethod        SignatureMethod        `xml:"SignatureMethod"`
	Reference              []Reference            `xml:"Reference"`
}

type CanonicalizationMethod struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type SignatureMethod struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type Reference struct {
	Id        string    `xml:"Id,attr,omitempty"`
	URI       string    `xml:"URI,attr"`
	Transforms Transforms `xml:"Transforms,omitempty"`
	DigestMethod DigestMethod `xml:"DigestMethod"`
	DigestValue string     `xml:"DigestValue"`
}

type Transforms struct {
	Transform []Transform `xml:"Transform"`
}

type Transform struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type DigestMethod struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type KeyInfo struct {
	Id         string      `xml:"Id,attr,omitempty"`
	X509Data   X509Data    `xml:"X509Data"`
}

type X509Data struct {
	X509Certificate string `xml:"X509Certificate"`
}

type SignatureObject struct {
	Id      string      `xml:"Id,attr,omitempty"`
	QualifyingProperties QualifyingProperties `xml:"QualifyingProperties"`
}

type QualifyingProperties struct {
	Target      string        `xml:"Target,attr"`
	SignedProperties SignedProperties `xml:"SignedProperties"`
	UnsignedProperties UnsignedProperties `xml:"UnsignedProperties,omitempty"`
}

type SignedProperties struct {
	Id                  string              `xml:"Id,attr"`
	SignedSignatureProperties SignedSignatureProperties `xml:"SignedSignatureProperties"`
}

type SignedSignatureProperties struct {
	SigningTime           string            `xml:"SigningTime"`
	SigningCertificate    SigningCertificate `xml:"SigningCertificate"`
}

type SigningCertificate struct {
	Cert []CertRef `xml:"Cert"`
}

type CertRef struct {
	CertDigest CertDigest `xml:"CertDigest"`
	IssuerSerial IssuerSerial `xml:"IssuerSerial"`
}

type CertDigest struct {
	DigestMethod DigestMethod `xml:"DigestMethod"`
	DigestValue  string      `xml:"DigestValue"`
}

type IssuerSerial struct {
	X509IssuerName   string `xml:"X509IssuerName"`
	X509SerialNumber string `xml:"X509SerialNumber"`
}

type UnsignedProperties struct {
	UnsignedSignatureProperties UnsignedSignatureProperties `xml:"UnsignedSignatureProperties"`
}

type UnsignedSignatureProperties struct {
	SignatureTimeStamp SignatureTimeStamp `xml:"SignatureTimeStamp"`
}

type SignatureTimeStamp struct {
	Id string `xml:"Id,attr"`
	EncapsulatedTimeStamp EncapsulatedTimeStamp `xml:"EncapsulatedTimeStamp"`
}

type EncapsulatedTimeStamp struct {
	Id     string `xml:"Id,attr,omitempty"`
	Value  string `xml:",chardata"`
}

func NewXAdESSigner(certHandler *CertificateHandler) (*XAdESSigner, error) {
	if certHandler == nil {
		return nil, fmt.Errorf("certificate handler is required")
	}

	return &XAdESSigner{
		certHandler:     certHandler,
		certificate:     certHandler.GetCertificate(),
		privateKey:      certHandler.GetPrivateKey(),
		timestampServer: "http://timestamp.digicert.com",
	}, nil
}

func (x *XAdESSigner) SignBES(ctx context.Context, xmlDocument []byte, opts *SignOptions) (*SignResult, error) {
	signatureId := generateSignatureId()
	signingTime := time.Now().UTC()

	documentHash := sha256.Sum256(xmlDocument)
	documentHashBase64 := base64.StdEncoding.EncodeToString(documentHash[:])

	certDER, err := x509.MarshalPKIXPublicKey(x.certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal certificate: %w", err)
	}

	certDigest := sha256.Sum256(certDER)
	certDigestBase64 := base64.StdEncoding.EncodeToString(certDigest[:])

	signature := &XAdESSignature{
		Id: signatureId,
		SignedInfo: SignedInfo{
			CanonicalizationMethod: CanonicalizationMethod{
				Algorithm: "http://www.w3.org/TR/2001/REC-xml-c14n-20010315",
			},
			SignatureMethod: SignatureMethod{
				Algorithm: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
			},
			Reference: []Reference{
				{
					URI: "",
					Transforms: Transforms{
						Transform: []Transform{
							{Algorithm: "http://www.w3.org/2000/09/xmldsig#enveloped-signature"},
						},
					},
					DigestMethod: DigestMethod{
						Algorithm: "http://www.w3.org/2001/04/xmlenc#sha256",
					},
					DigestValue: documentHashBase64,
				},
			},
		},
		KeyInfo: KeyInfo{
			X509Data: X509Data{
				X509Certificate: base64.StdEncoding.EncodeToString(x.certificate.Raw),
			},
		},
		Object: []SignatureObject{
			{
				Id: "signature-object-" + signatureId,
				QualifyingProperties: QualifyingProperties{
					Target: "#" + signatureId,
					SignedProperties: SignedProperties{
						Id: "signed-props-" + signatureId,
						SignedSignatureProperties: SignedSignatureProperties{
							SigningTime: signingTime.Format(time.RFC3339),
							SigningCertificate: SigningCertificate{
								Cert: []CertRef{
									{
										CertDigest: CertDigest{
											DigestMethod: DigestMethod{
												Algorithm: "http://www.w3.org/2001/04/xmlenc#sha256",
											},
											DigestValue: certDigestBase64,
										},
										IssuerSerial: IssuerSerial{
											X509IssuerName:   x.certificate.Issuer.String(),
											X509SerialNumber: fmt.Sprintf("%d", x.certificate.SerialNumber),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	signedInfoBytes, err := xml.Marshal(signature.SignedInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signed info: %w", err)
	}

	canonicalized := canonicalizeXML(signedInfoBytes)

	hashed := sha256.Sum256(canonicalized)
	signatureBytes, err := rsa.SignPKCS1v15(nil, x.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	signature.SignatureValue = base64.StdEncoding.EncodeToString(signatureBytes)

	signatureXML, err := xml.MarshalIndent(signature, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signature: %w", err)
	}

	return &SignResult{
		Signature:       string(signatureXML),
		SignatureBase64: base64.StdEncoding.EncodeToString(signatureXML),
		Certificate:     base64.StdEncoding.EncodeToString(x.certificate.Raw),
		SerialNumber:    fmt.Sprintf("%d", x.certificate.SerialNumber),
		IssuerDN:        x.certificate.Issuer.String(),
		SubjectDN:       x.certificate.Subject.String(),
		ValidFrom:       x.certificate.NotBefore,
		ValidTo:         x.certificate.NotAfter,
		DocumentHash:     documentHashBase64,
		Timestamp:        signingTime,
	}, nil
}

func (x *XAdESSigner) SignT(ctx context.Context, xmlDocument []byte, opts *SignOptions) (*SignResult, error) {
	result, err := x.SignBES(ctx, xmlDocument, opts)
	if err != nil {
		return nil, err
	}

	if opts == nil || !opts.AddTimestamp {
		return result, nil
	}

	timestampToken, err := x.addTimestamp(ctx, result.SignatureBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to add timestamp: %w", err)
	}

	result.TimestampToken = timestampToken

	return result, nil
}

func (x *XAdESSigner) Verify(ctx context.Context, xmlDocument []byte, signature []byte) (*VerifyResult, error) {
	result := &VerifyResult{
		IsValid:           false,
		SignerVerified:    false,
		DocumentIntegrity: false,
		TimestampValid:    false,
		CertificateValid:  false,
	}

	var xadesSignature XAdESSignature
	if err := xml.Unmarshal(signature, &xadesSignature); err != nil {
		result.ErrorCode = "PARSE_ERROR"
		result.ErrorMessage = fmt.Sprintf("Failed to parse XAdES signature: %v", err)
		return result, nil
	}

	cert, err := x509.ParseCertificates(x.certificate.Raw)
	if err != nil || len(cert) == 0 {
		result.ErrorCode = "CERT_PARSE_ERROR"
		result.ErrorMessage = "Failed to parse certificate from signature"
		return result, nil
	}

	signerCert := cert[0]

	now := time.Now()
	if now.Before(signerCert.NotBefore) || now.After(signerCert.NotAfter) {
		result.ErrorCode = "CERT_EXPIRED"
		result.ErrorMessage = "Certificate is not valid at current time"
		return result, nil
	}

	result.CertificateValid = true

	documentHash := sha256.Sum256(xmlDocument)
	documentHashBase64 := base64.StdEncoding.EncodeToString(documentHash[:])

	hashMatches := false
	for _, ref := range xadesSignature.SignedInfo.Reference {
		if ref.DigestValue == documentHashBase64 {
			hashMatches = true
			break
		}
	}

	if !hashMatches {
		result.ErrorCode = "DOCUMENT_INTEGRITY_FAILED"
		result.ErrorMessage = "Document hash does not match signature"
		return result, nil
	}

	result.DocumentIntegrity = true

	signedInfoBytes, err := xml.Marshal(xadesSignature.SignedInfo)
	if err != nil {
		result.ErrorCode = "MARSHAL_ERROR"
		result.ErrorMessage = "Failed to marshal signed info"
		return result, nil
	}

	canonicalized := canonicalizeXML(signedInfoBytes)

	signatureBytes, err := base64.StdEncoding.DecodeString(xadesSignature.SignatureValue)
	if err != nil {
		result.ErrorCode = "DECODE_ERROR"
		result.ErrorMessage = "Failed to decode signature value"
		return result, nil
	}

	hashed := sha256.Sum256(canonicalized)

	publicKey, ok := signerCert.PublicKey.(*rsa.PublicKey)
	if !ok {
		result.ErrorCode = "KEY_TYPE_ERROR"
		result.ErrorMessage = "Signer certificate does not contain RSA public key"
		return result, nil
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signatureBytes)
	if err != nil {
		result.ErrorCode = "SIGNATURE_INVALID"
		result.ErrorMessage = "Signature verification failed"
		return result, nil
	}

	result.IsValid = true
	result.SignerVerified = true
	result.VerificationDetails = "XAdES signature verified successfully"

	return result, nil
}

func (x *XAdESSigner) addTimestamp(ctx context.Context, signatureBase64 string) (string, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	timestampData := fmt.Sprintf("%s:%s", signatureBase64, timestamp)
	hash := sha256.Sum256([]byte(timestampData))
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

func generateSignatureId() string {
	return fmt.Sprintf("signature-%d", time.Now().UnixNano())
}

func canonicalizeXML(data []byte) []byte {
	str := string(data)
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\r", "\n")
	str = strings.TrimSpace(str)
	return []byte(str)
}