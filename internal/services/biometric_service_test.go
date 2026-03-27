package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockFaceComparator struct {
	mock.Mock
}

func (m *MockFaceComparator) Compare(ctx context.Context, input CompareFacesInput) (*CompareFacesOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CompareFacesOutput), args.Error(1)
}

type MockLivenessChecker struct {
	mock.Mock
}

func (m *MockLivenessChecker) Detect(ctx context.Context, input DetectLivenessInput) (*DetectLivenessOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DetectLivenessOutput), args.Error(1)
}

type MockImageProcessor struct {
	mock.Mock
}

func (m *MockImageProcessor) Process(imageData string, imageType ImageType) ([]byte, error) {
	args := m.Called(imageData, imageType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageProcessor) ValidateFormat(imageData string, imageType ImageType) error {
	args := m.Called(imageData, imageType)
	return args.Error(0)
}

func (m *MockImageProcessor) DetectFaces(imageData []byte) ([]FaceInfo, error) {
	args := m.Called(imageData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]FaceInfo), args.Error(1)
}

func TestBiometricService_CompareFaces_Success(t *testing.T) {
	mockFaceComparator := new(MockFaceComparator)
	mockLivenessChecker := new(MockLivenessChecker)
	mockImageProcessor := new(MockImageProcessor)

	service := &biometricService{
		faceComparator:  mockFaceComparator,
		livenessChecker: mockLivenessChecker,
		imageProcessor:  mockImageProcessor,
		provider:        ProviderLocal,
	}

	ctx := context.Background()
	input := CompareFacesInput{
		SourceImage:     "base64encodedsourcedata",
		SourceImageType: ImageTypeBase64,
		TargetImage:     "base64encodedtargetdata",
		TargetImageType: ImageTypeBase64,
		UserID:          1,
	}

	expectedOutput := &CompareFacesOutput{
		IsMatch:          true,
		SimilarityScore:  0.98,
		ConfidenceLevel:  0.95,
		ProcessingTimeMs: 150,
	}

	mockFaceComparator.On("Compare", ctx, input).Return(expectedOutput, nil)

	result, err := service.CompareFaces(ctx, input)

	require.NoError(t, err)
	assert.True(t, result.IsMatch)
	assert.Equal(t, 0.98, result.SimilarityScore)
	assert.Equal(t, 0.95, result.ConfidenceLevel)
	mockFaceComparator.AssertExpectations(t)
}

func TestBiometricService_CompareFaces_NoMatch(t *testing.T) {
	mockFaceComparator := new(MockFaceComparator)
	mockLivenessChecker := new(MockLivenessChecker)
	mockImageProcessor := new(MockImageProcessor)

	service := &biometricService{
		faceComparator:  mockFaceComparator,
		livenessChecker: mockLivenessChecker,
		imageProcessor:  mockImageProcessor,
		provider:        ProviderLocal,
	}

	ctx := context.Background()
	input := CompareFacesInput{
		SourceImage:     "differentface1",
		TargetImage:     "differentface2",
		SourceImageType: ImageTypeBase64,
		TargetImageType: ImageTypeBase64,
		UserID:          1,
	}

	expectedOutput := &CompareFacesOutput{
		IsMatch:           false,
		SimilarityScore:   0.45,
		ConfidenceLevel:   0.80,
		DetectedAnomalies: []string{"different_person"},
		ProcessingTimeMs:  120,
	}

	mockFaceComparator.On("Compare", ctx, input).Return(expectedOutput, nil)

	result, err := service.CompareFaces(ctx, input)

	require.NoError(t, err)
	assert.False(t, result.IsMatch)
	assert.Equal(t, 0.45, result.SimilarityScore)
	mockFaceComparator.AssertExpectations(t)
}

func TestBiometricService_CompareFaces_Error(t *testing.T) {
	mockFaceComparator := new(MockFaceComparator)
	mockLivenessChecker := new(MockLivenessChecker)
	mockImageProcessor := new(MockImageProcessor)

	service := &biometricService{
		faceComparator:  mockFaceComparator,
		livenessChecker: mockLivenessChecker,
		imageProcessor:  mockImageProcessor,
		provider:        ProviderLocal,
	}

	ctx := context.Background()
	input := CompareFacesInput{
		SourceImage:     "invalidimage",
		SourceImageType: ImageTypeBase64,
		TargetImage:     "validimage",
		TargetImageType: ImageTypeBase64,
		UserID:          1,
	}

	mockFaceComparator.On("Compare", ctx, input).Return(nil, errors.New("no face detected in source image"))

	result, err := service.CompareFaces(ctx, input)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no face detected")
	mockFaceComparator.AssertExpectations(t)
}

func TestBiometricService_DetectLiveness_Success(t *testing.T) {
	mockFaceComparator := new(MockFaceComparator)
	mockLivenessChecker := new(MockLivenessChecker)
	mockImageProcessor := new(MockImageProcessor)

	service := &biometricService{
		faceComparator:  mockFaceComparator,
		livenessChecker: mockLivenessChecker,
		imageProcessor:  mockImageProcessor,
		provider:        ProviderLocal,
	}

	ctx := context.Background()
	input := DetectLivenessInput{
		Images:     []string{"image1", "image2", "image3"},
		ImageTypes: []ImageType{ImageTypeBase64, ImageTypeBase64, ImageTypeBase64},
		UserID:     1,
	}

	expectedOutput := &DetectLivenessOutput{
		IsLive:           true,
		LivenessScore:    0.97,
		ConfidenceLevel:  0.95,
		SpoofProbability: 0.02,
		ProcessingTimeMs: 250,
	}

	mockLivenessChecker.On("Detect", ctx, input).Return(expectedOutput, nil)

	result, err := service.DetectLiveness(ctx, input)

	require.NoError(t, err)
	assert.True(t, result.IsLive)
	assert.Equal(t, 0.97, result.LivenessScore)
	assert.LessOrEqual(t, result.SpoofProbability, 0.1)
	mockLivenessChecker.AssertExpectations(t)
}

func TestBiometricService_DetectLiveness_SpoofDetected(t *testing.T) {
	mockFaceComparator := new(MockFaceComparator)
	mockLivenessChecker := new(MockLivenessChecker)
	mockImageProcessor := new(MockImageProcessor)

	service := &biometricService{
		faceComparator:  mockFaceComparator,
		livenessChecker: mockLivenessChecker,
		imageProcessor:  mockImageProcessor,
		provider:        ProviderLocal,
	}

	ctx := context.Background()
	input := DetectLivenessInput{
		Images:     []string{"spoofimage"},
		ImageTypes: []ImageType{ImageTypeBase64},
		UserID:     1,
	}

	expectedOutput := &DetectLivenessOutput{
		IsLive:           false,
		LivenessScore:    0.35,
		ConfidenceLevel:  0.90,
		SpoofProbability: 0.78,
		DetectedAttacks:  []string{"photo_replay", "screen_replay"},
		ProcessingTimeMs: 180,
	}

	mockLivenessChecker.On("Detect", ctx, input).Return(expectedOutput, nil)

	result, err := service.DetectLiveness(ctx, input)

	require.NoError(t, err)
	assert.False(t, result.IsLive)
	assert.GreaterOrEqual(t, result.SpoofProbability, 0.5)
	assert.Contains(t, result.DetectedAttacks, "photo_replay")
	mockLivenessChecker.AssertExpectations(t)
}

func TestBiometricService_DetectLiveness_Error(t *testing.T) {
	mockFaceComparator := new(MockFaceComparator)
	mockLivenessChecker := new(MockLivenessChecker)
	mockImageProcessor := new(MockImageProcessor)

	service := &biometricService{
		faceComparator:  mockFaceComparator,
		livenessChecker: mockLivenessChecker,
		imageProcessor:  mockImageProcessor,
		provider:        ProviderLocal,
	}

	ctx := context.Background()
	input := DetectLivenessInput{
		Images:     []string{},
		ImageTypes: []ImageType{},
		UserID:     1,
	}

	mockLivenessChecker.On("Detect", ctx, input).Return(nil, errors.New("no images provided"))

	result, err := service.DetectLiveness(ctx, input)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no images provided")
	mockLivenessChecker.AssertExpectations(t)
}

func TestBiometricService_MockBiometricAPIResponses(t *testing.T) {
	t.Run("mock high confidence match response", func(t *testing.T) {
		mockFaceComparator := new(MockFaceComparator)
		mockLivenessChecker := new(MockLivenessChecker)

		service := &biometricService{
			faceComparator:  mockFaceComparator,
			livenessChecker: mockLivenessChecker,
			provider:        ProviderAWS,
		}

		ctx := context.Background()
		input := CompareFacesInput{
			SourceImage:     "highqualityface",
			TargetImage:     "highqualityface",
			SourceImageType: ImageTypeBase64,
			TargetImageType: ImageTypeBase64,
			UserID:          1,
		}

		expectedOutput := &CompareFacesOutput{
			IsMatch:          true,
			SimilarityScore:  0.99,
			ConfidenceLevel:  0.99,
			ProcessingTimeMs: 100,
			ProviderResult:   "AWS",
		}

		mockFaceComparator.On("Compare", ctx, input).Return(expectedOutput, nil)

		result, err := service.CompareFaces(ctx, input)
		require.NoError(t, err)
		assert.True(t, result.IsMatch)
		assert.Equal(t, "AWS", result.ProviderResult)
	})

	t.Run("mock low confidence response", func(t *testing.T) {
		mockFaceComparator := new(MockFaceComparator)
		mockLivenessChecker := new(MockLivenessChecker)

		service := &biometricService{
			faceComparator:  mockFaceComparator,
			livenessChecker: mockLivenessChecker,
			provider:        ProviderGoogle,
		}

		ctx := context.Background()
		input := CompareFacesInput{
			SourceImage:     "lowqualityface",
			TargetImage:     "lowqualityface",
			SourceImageType: ImageTypeBase64,
			TargetImageType: ImageTypeBase64,
			UserID:          1,
		}

		expectedOutput := &CompareFacesOutput{
			IsMatch:           true,
			SimilarityScore:   0.65,
			ConfidenceLevel:   0.55,
			DetectedAnomalies: []string{"low_quality", "blur_detected"},
			ProcessingTimeMs:  200,
			ProviderResult:    "Google",
		}

		mockFaceComparator.On("Compare", ctx, input).Return(expectedOutput, nil)

		result, err := service.CompareFaces(ctx, input)
		require.NoError(t, err)
		assert.True(t, result.IsMatch)
		assert.Equal(t, "Google", result.ProviderResult)
		assert.Contains(t, result.DetectedAnomalies, "blur_detected")
	})
}

func TestBiometricService_Providers(t *testing.T) {
	mockFaceComparator := new(MockFaceComparator)
	mockLivenessChecker := new(MockLivenessChecker)
	mockImageProcessor := new(MockImageProcessor)

	service := &biometricService{
		faceComparator:  mockFaceComparator,
		livenessChecker: mockLivenessChecker,
		imageProcessor:  mockImageProcessor,
		provider:        ProviderNone,
	}

	t.Run("set and get provider", func(t *testing.T) {
		assert.Equal(t, ProviderNone, service.GetProvider())

		service.SetProvider(ProviderAWS)
		assert.Equal(t, ProviderAWS, service.GetProvider())

		service.SetProvider(ProviderAzure)
		assert.Equal(t, ProviderAzure, service.GetProvider())

		service.SetProvider(ProviderGoogle)
		assert.Equal(t, ProviderGoogle, service.GetProvider())

		service.SetProvider(ProviderLocal)
		assert.Equal(t, ProviderLocal, service.GetProvider())
	})
}

func TestNewBiometricService(t *testing.T) {
	service := NewBiometricService()
	assert.NotNil(t, service)
	assert.Equal(t, ProviderNone, service.GetProvider())
}

func TestNewBiometricServiceWithProvider(t *testing.T) {
	providers := []BiometricProvider{ProviderAWS, ProviderAzure, ProviderGoogle, ProviderLocal}

	for _, provider := range providers {
		service := NewBiometricServiceWithProvider(provider)
		assert.NotNil(t, service)
		assert.Equal(t, provider, service.GetProvider())
	}
}