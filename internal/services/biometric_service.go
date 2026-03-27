package services

import (
	"context"

	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type BiometricProvider string

const (
	ProviderNone       BiometricProvider = "none"
	ProviderAWS       BiometricProvider = "aws"
	ProviderAzure     BiometricProvider = "azure"
	ProviderGoogle    BiometricProvider = "google"
	ProviderLocal     BiometricProvider = "local"
)

type CompareFacesInput struct {
	SourceImage      string
	SourceImageType  ImageType
	TargetImage      string
	TargetImageType  ImageType
	UserID           uint
}

type CompareFacesOutput struct {
	IsMatch          bool
	SimilarityScore  float64
	ConfidenceLevel  float64
	DetectedAnomalies []string
	ProcessingTimeMs int64
	ProviderResult   string
}

type DetectLivenessInput struct {
	Images      []string
	ImageTypes  []ImageType
	VideoFile   string
	UserID      uint
}

type DetectLivenessOutput struct {
	IsLive           bool
	LivenessScore    float64
	ConfidenceLevel  float64
	SpoofProbability float64
	DetectedAttacks  []string
	ProcessingTimeMs int64
	ProviderResult   string
}

type ImageType string

const (
	ImageTypeBase64 ImageType = "base64"
	ImageTypeURL    ImageType = "url"
	ImageTypePath   ImageType = "path"
)

type BiometricService interface {
	CompareFaces(ctx context.Context, input CompareFacesInput) (*CompareFacesOutput, error)
	DetectLiveness(ctx context.Context, input DetectLivenessInput) (*DetectLivenessOutput, error)
	SetProvider(provider BiometricProvider)
	GetProvider() BiometricProvider
}

type biometricService struct {
	faceComparator  FaceComparator
	livenessChecker LivenessChecker
	imageProcessor  ImageProcessor
	provider        BiometricProvider
}

func NewBiometricService() BiometricService {
	imageProcessor := NewImageProcessor()
	return &biometricService{
		faceComparator:  NewFaceComparator(imageProcessor),
		livenessChecker: NewLivenessChecker(imageProcessor),
		imageProcessor:  imageProcessor,
		provider:        ProviderNone,
	}
}

func NewBiometricServiceWithProvider(provider BiometricProvider) BiometricService {
	imageProcessor := NewImageProcessor()
	return &biometricService{
		faceComparator:  NewFaceComparator(imageProcessor),
		livenessChecker: NewLivenessChecker(imageProcessor),
		imageProcessor:  imageProcessor,
		provider:        provider,
	}
}

func (s *biometricService) CompareFaces(ctx context.Context, input CompareFacesInput) (*CompareFacesOutput, error) {
	return s.faceComparator.Compare(ctx, input)
}

func (s *biometricService) DetectLiveness(ctx context.Context, input DetectLivenessInput) (*DetectLivenessOutput, error) {
	return s.livenessChecker.Detect(ctx, input)
}

func (s *biometricService) SetProvider(provider BiometricProvider) {
	s.provider = provider
}

func (s *biometricService) GetProvider() BiometricProvider {
	return s.provider
}

func (s *biometricService) SaveFacialComparisonResult(request *models.FacialComparisonRequest, output *CompareFacesOutput) *models.FacialComparisonResponse {
	response := &models.FacialComparisonResponse{
		RequestID:        request.ID,
		IsMatch:          output.IsMatch,
		SimilarityScore:  output.SimilarityScore,
		ConfidenceLevel:  output.ConfidenceLevel,
		DetectedAnomalies: stringSliceToText(output.DetectedAnomalies),
		ProcessingTime:   output.ProcessingTimeMs,
		ProviderResult:   output.ProviderResult,
	}
	return response
}

func (s *biometricService) SaveLivenessDetectionResult(request *models.LivenessDetectionRequest, output *DetectLivenessOutput) *models.LivenessDetectionResponse {
	response := &models.LivenessDetectionResponse{
		RequestID:          request.ID,
		IsLive:             output.IsLive,
		LivenessScore:      output.LivenessScore,
		ConfidenceLevel:    output.ConfidenceLevel,
		SpoofProbability:   output.SpoofProbability,
		DetectedAttacks:    stringSliceToText(output.DetectedAttacks),
		ProcessingTime:     output.ProcessingTimeMs,
		ProviderResult:     output.ProviderResult,
	}
	return response
}

func stringSliceToText(slice []string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += ";"
		}
		result += s
	}
	return result
}