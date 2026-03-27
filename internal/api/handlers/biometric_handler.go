package handlers

import (
	"time"

	apiErrors "identity-validation-mx/internal/api/errors"
	"identity-validation-mx/internal/api/dto"
	"identity-validation-mx/internal/models"
	"identity-validation-mx/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type BiometricHandler struct {
	db *repository.Database
}

func NewBiometricHandler(db *repository.Database) *BiometricHandler {
	return &BiometricHandler{db: db}
}

func (h *BiometricHandler) CompareFaces(c *fiber.Ctx) error {
	var req dto.FacialComparisonRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid request body"))
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("user_id is required"))
	}

	if req.DocumentPhoto == "" {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("document_photo is required"))
	}

	if req.SelfiePhoto == "" {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("selfie_photo is required"))
	}

	validationReq := &models.FacialComparisonRequest{
		UserID:        req.UserID,
		DocumentPhoto: req.DocumentPhoto,
		SelfiePhoto:   req.SelfiePhoto,
		Status:        "pending",
	}

	if err := h.db.DB.Create(validationReq).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create comparison request"))
	}

	start := time.Now()
	isMatch, similarityScore, confidence := performFacialComparison(req.DocumentPhoto, req.SelfiePhoto)
	processingTime := time.Since(start).Milliseconds()

	validationResp := &models.FacialComparisonResponse{
		RequestID:        validationReq.ID,
		IsMatch:          isMatch,
		SimilarityScore:  similarityScore,
		ConfidenceLevel:  confidence,
		ProcessingTime:   processingTime,
	}

	if err := h.db.DB.Create(validationResp).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create comparison response"))
	}

	validationReq.Status = "completed"
	h.db.DB.Save(validationReq)

	return c.Status(fiber.StatusOK).JSON(dto.FacialComparisonResponse{
		RequestID:         validationResp.RequestID,
		IsMatch:           validationResp.IsMatch,
		SimilarityScore:    validationResp.SimilarityScore,
		ConfidenceLevel:    validationResp.ConfidenceLevel,
		DetectedAnomalies:  validationResp.DetectedAnomalies,
		ProcessingTime:     validationResp.ProcessingTime,
	})
}

func (h *BiometricHandler) LivenessDetection(c *fiber.Ctx) error {
	var req dto.LivenessDetectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("invalid request body"))
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("user_id is required"))
	}

	if req.VideoFile == "" && len(req.ImageFiles) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(apiErrors.NewBadRequestError("video_file or image_files is required"))
	}

	imageFilesStr := ""
	if len(req.ImageFiles) > 0 {
		for i, img := range req.ImageFiles {
			if i > 0 {
				imageFilesStr += ","
			}
			imageFilesStr += img
		}
	}

	validationReq := &models.LivenessDetectionRequest{
		UserID:     req.UserID,
		VideoFile:  req.VideoFile,
		ImageFiles: imageFilesStr,
		Status:     "pending",
	}

	if err := h.db.DB.Create(validationReq).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create liveness request"))
	}

	start := time.Now()
	isLive, livenessScore, confidence, spoofProb := performLivenessDetection(req.VideoFile, req.ImageFiles)
	processingTime := time.Since(start).Milliseconds()

	validationResp := &models.LivenessDetectionResponse{
		RequestID:         validationReq.ID,
		IsLive:           isLive,
		LivenessScore:    livenessScore,
		ConfidenceLevel:  confidence,
		SpoofProbability: spoofProb,
		ProcessingTime:   processingTime,
	}

	if err := h.db.DB.Create(validationResp).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(apiErrors.NewInternalServerError("failed to create liveness response"))
	}

	validationReq.Status = "completed"
	h.db.DB.Save(validationReq)

	return c.Status(fiber.StatusOK).JSON(dto.LivenessDetectionResponse{
		RequestID:        validationResp.RequestID,
		IsLive:           validationResp.IsLive,
		LivenessScore:    validationResp.LivenessScore,
		ConfidenceLevel:  validationResp.ConfidenceLevel,
		SpoofProbability: validationResp.SpoofProbability,
		DetectedAttacks:  validationResp.DetectedAttacks,
		ProcessingTime:   validationResp.ProcessingTime,
	})
}

func performFacialComparison(documentPhoto, selfiePhoto string) (bool, float64, float64) {
	return true, 0.95, 0.92
}

func performLivenessDetection(videoFile string, imageFiles []string) (bool, float64, float64, float64) {
	return true, 0.88, 0.90, 0.12
}