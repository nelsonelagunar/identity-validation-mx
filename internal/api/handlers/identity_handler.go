package handlers

import (
	"strconv"

	apiErrors "identity-validation-mx/internal/api/errors"
	"identity-validation-mx/internal/api/dto"
	"identity-validation-mx/internal/models"
	"identity-validation-mx/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IdentityHandler struct {
	db *repository.Database
}

func NewIdentityHandler(db *repository.Database) *IdentityHandler {
	return &IdentityHandler{db: db}
}

func (h *IdentityHandler) ValidateCURP(c *fiber.Ctx) error {
	var req dto.CURPValidationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid request body"))
	}

	if len(req.CURP) != 18 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.ErrInvalidCURP)
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("user_id is required"))
	}

	validationReq := &models.CURPValidationRequest{
		CURP:   req.CURP,
		UserID: req.UserID,
		Status: "pending",
	}

	if err := h.db.DB.Create(validationReq).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create validation request"))
	}

	isValid := validateCURPFormat(req.CURP)

	validationResp := &models.CURPValidationResponse{
		RequestID:       validationReq.ID,
		IsValid:         isValid,
		RenapoVerified:  false,
		VerificationScore: 0.85,
	}

	if isValid {
		validationResp.FullName = extractNameFromCURP(req.CURP)
		validationResp.Gender = string(req.CURP[10])
	}

	if err := h.db.DB.Create(validationResp).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create validation response"))
	}

	validationReq.Status = "completed"
	h.db.DB.Save(validationReq)

	return c.Status(fiber.StatusOK).JSON(dto.CURPValidationResponse{
		RequestID:         validationResp.RequestID,
		IsValid:          validationResp.IsValid,
		FullName:         validationResp.FullName,
		Gender:           validationResp.Gender,
		BirthState:        validationResp.BirthState,
		ValidationError:   validationResp.ValidationError,
		RenapoVerified:    validationResp.RenapoVerified,
		VerificationScore: validationResp.VerificationScore,
	})
}

func (h *IdentityHandler) ValidateRFC(c *fiber.Ctx) error {
	var req dto.RFCValidationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid request body"))
	}

	if len(req.RFC) < 12 || len(req.RFC) > 13 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.ErrInvalidRFC)
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("user_id is required"))
	}

	validationReq := &models.RFCValidationRequest{
		RFC:    req.RFC,
		UserID: req.UserID,
		Status: "pending",
	}

	if err := h.db.DB.Create(validationReq).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create validation request"))
	}

	isValid := validateRFCCTFormat(req.RFC)

	validationResp := &models.RFCValidationResponse{
		RequestID:       validationReq.ID,
		IsValid:         isValid,
		SatVerified:     false,
		VerificationScore: 0.80,
	}

	if isValid {
		validationResp.FullName = extractNameFromRFC(req.RFC)
	}

	if err := h.db.DB.Create(validationResp).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create validation response"))
	}

	validationReq.Status = "completed"
	h.db.DB.Save(validationReq)

	return c.Status(fiber.StatusOK).JSON(dto.RFCValidationResponse{
		RequestID:         validationResp.RequestID,
		IsValid:          validationResp.IsValid,
		FullName:         validationResp.FullName,
		TaxRegime:        validationResp.TaxRegime,
		StatusSAT:        validationResp.StatusSAT,
		ValidationError:   validationResp.ValidationError,
		SatVerified:      validationResp.SatVerified,
		VerificationScore: validationResp.VerificationScore,
	})
}

func (h *IdentityHandler) ValidateINE(c *fiber.Ctx) error {
	var req dto.INEValidationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid request body"))
	}

	if len(req.INEClave) != 18 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.ErrInvalidINE)
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("user_id is required"))
	}

	validationReq := &models.INEValidationRequest{
		INEClave:    req.INEClave,
		UserID:      req.UserID,
		OCRNumber:   req.OCRNumber,
		ElectionKey: req.ElectionKey,
		Status:      "pending",
	}

	if err := h.db.DB.Create(validationReq).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create validation request"))
	}

	isValid := validateINEFormat(req.INEClave)

	validationResp := &models.INEValidationResponse{
		RequestID:       validationReq.ID,
		IsValid:         isValid,
		INEVerified:     false,
		VerificationScore: 0.75,
	}

	if isValid {
		validationResp.FullName = "Extracted from INE"
	}

	if err := h.db.DB.Create(validationResp).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create validation response"))
	}

	validationReq.Status = "completed"
	h.db.DB.Save(validationReq)

	return c.Status(fiber.StatusOK).JSON(dto.INEValidationResponse{
		RequestID:         validationResp.RequestID,
		IsValid:          validationResp.IsValid,
		FullName:         validationResp.FullName,
		Gender:           validationResp.Gender,
		VotingSection:    validationResp.VotingSection,
		ValidationError:   validationResp.ValidationError,
		INEVerified:      validationResp.INEVerified,
		VerificationScore: validationResp.VerificationScore,
	})
}

func (h *IdentityHandler) GetCURPValidation(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid id parameter"))
	}

	var response models.CURPValidationResponse
	if err := h.db.DB.Where("request_id = ?", id).First(&response).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(apiErrors.NewNotFoundError("validation not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to fetch validation"))
	}

	return c.Status(fiber.StatusOK).JSON(dto.CURPValidationResponse{
		RequestID:         response.RequestID,
		IsValid:          response.IsValid,
		FullName:         response.FullName,
		Gender:           response.Gender,
		BirthState:       response.BirthState,
		ValidationError:   response.ValidationError,
		RenapoVerified:    response.RenapoVerified,
		VerificationScore: response.VerificationScore,
	})
}

func validateCURPFormat(curp string) bool {
	if len(curp) != 18 {
		return false
	}
	return true
}

func validateRFCCTFormat(rfc string) bool {
	l := len(rfc)
	return l == 12 || l == 13
}

func validateINEFormat(ine string) bool {
	return len(ine) == 18
}

func extractNameFromCURP(curp string) string {
	return "Name extracted from CURP"
}

func extractNameFromRFC(rfc string) string {
	return "Name extracted from RFC"
}