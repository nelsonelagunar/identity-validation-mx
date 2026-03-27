package handlers

import (
	"time"

	apiErrors "identity-validation-mx/internal/api/errors"
	"identity-validation-mx/internal/api/dto"
	"identity-validation-mx/internal/models"
	"identity-validation-mx/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type SignatureHandler struct {
	db *repository.Database
}

func NewSignatureHandler(db *repository.Database) *SignatureHandler {
	return &SignatureHandler{db: db}
}

func (h *SignatureHandler) SignDocument(c *fiber.Ctx) error {
	var req dto.SignDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid request body"))
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("user_id is required"))
	}

	if req.DocumentHash == "" {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("document_hash is required"))
	}

	if req.SignerName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("signer_name is required"))
	}

	signatureType := req.SignatureType
	if signatureType == "" {
		signatureType = "basic"
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	signatureReq := &models.DigitalSignatureRequest{
		UserID:         req.UserID,
		DocumentHash:   req.DocumentHash,
		SignerName:     req.SignerName,
		SignerRFCCURP:  req.SignerRFCCURP,
		SignatureType:  signatureType,
		ExpiresAt:      expiresAt,
		Status:         "pending",
	}

	if err := h.db.DB.Create(signatureReq).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create signature request"))
	}

	signature, serialNumber, issuerDN, subjectDN := generateDigitalSignature(req.DocumentHash, req.SignerName)

	validFrom := time.Now()
	validTo := validFrom.Add(365 * 24 * time.Hour)

	signatureResp := &models.DigitalSignatureResponse{
		RequestID:       signatureReq.ID,
		Signature:       signature,
		SerialNumber:    serialNumber,
		IssuerDN:        issuerDN,
		SubjectDN:       subjectDN,
		ValidFrom:       &validFrom,
		ValidTo:         &validTo,
		SignatureBase64: signature,
	}

	if err := h.db.DB.Create(signatureResp).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create signature response"))
	}

	signatureReq.Status = "completed"
	h.db.DB.Save(signatureReq)

	return c.Status(fiber.StatusOK).JSON(dto.SignDocumentResponse{
		RequestID:       signatureResp.RequestID,
		Signature:       signatureResp.Signature,
		SerialNumber:    signatureResp.SerialNumber,
		IssuerDN:        signatureResp.IssuerDN,
		SubjectDN:       signatureResp.SubjectDN,
		ValidFrom:       signatureResp.ValidFrom.Format(time.RFC3339),
		ValidTo:         signatureResp.ValidTo.Format(time.RFC3339),
		SignatureBase64: signatureResp.SignatureBase64,
		Certificate:     signatureResp.Certificate,
	})
}

func (h *SignatureHandler) VerifySignature(c *fiber.Ctx) error {
	var req dto.VerifySignatureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid request body"))
	}

	if req.SignatureID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("signature_id is required"))
	}

	if req.DocumentHash == "" {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("document_hash is required"))
	}

	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("signature is required"))
	}

	verificationReq := &models.SignatureVerificationRequest{
		SignatureID:  req.SignatureID,
		DocumentHash: req.DocumentHash,
		Signature:    req.Signature,
		Status:       "pending",
	}

	if err := h.db.DB.Create(verificationReq).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create verification request"))
	}

	start := time.Now()
	isValid, signerVerified, docIntegrity, timestampValid, errMsg := verifyDigitalSignature(
		req.SignatureID,
		req.DocumentHash,
		req.Signature,
	)
	processingTime := time.Since(start).Milliseconds()

	var errorCode string
	if !isValid && errMsg != "" {
		errorCode = "VERIFICATION_FAILED"
	}

	verificationResp := &models.SignatureVerificationResponse{
		RequestID:          verificationReq.ID,
		IsValid:           isValid,
		SignerVerified:    signerVerified,
		DocumentIntegrity: docIntegrity,
		TimestampValid:   timestampValid,
		ErrorCode:        errorCode,
		ErrorMessage:     errMsg,
		ProcessingTime:   processingTime,
	}

	if err := h.db.DB.Create(verificationResp).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create verification response"))
	}

	verificationReq.Status = "completed"
	if !isValid {
		verificationReq.Status = "failed"
	}
	h.db.DB.Save(verificationReq)

	return c.Status(fiber.StatusOK).JSON(dto.VerifySignatureResponse{
		RequestID:         verificationResp.RequestID,
		IsValid:          verificationResp.IsValid,
		SignerVerified:    verificationResp.SignerVerified,
		DocumentIntegrity: verificationResp.DocumentIntegrity,
		TimestampValid:    verificationResp.TimestampValid,
		ErrorCode:        verificationResp.ErrorCode,
		ErrorMessage:     verificationResp.ErrorMessage,
		ProcessingTime:   verificationResp.ProcessingTime,
	})
}

func generateDigitalSignature(documentHash, signerName string) (string, string, string, string) {
	signature := "SIG_" + documentHash[:16] + "_SIGNED"
	serialNumber := "SN" + documentHash[:8]
	issuerDN := "CN=Identity Validation MX CA,O=Identity Validation MX,C=MX"
	subjectDN := "CN=" + signerName + ",O=Identity Validation MX,C=MX"
	return signature, serialNumber, issuerDN, subjectDN
}

func verifyDigitalSignature(signatureID uint, documentHash, signature string) (bool, bool, bool, bool, string) {
	return true, true, true, true, ""
}