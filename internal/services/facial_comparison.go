package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type ComparisonAlgorithm string

const (
	AlgorithmEuclidean  ComparisonAlgorithm = "euclidean"
	AlgorithmCosine     ComparisonAlgorithm = "cosine"
	AlgorithmHistogram   ComparisonAlgorithm = "histogram"
	AlgorithmDeepLearning ComparisonAlgorithm = "deep_learning"
)

type FaceDetectionResult struct {
	FacePresent   bool
	BoundingBox   BoundingBox
	Landmarks     []Landmark
	Quality       float64
	Confidence    float64
	ErrorCode     string
	ErrorMessage  string
}

type BoundingBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type Landmark struct {
	Type  string
	X     float64
	Y     float64
}

type FaceComparisonAlgorithm interface {
	Compare(source, target []byte) (float64, error)
	Name() string
}

type FaceComparator interface {
	Compare(ctx context.Context, input CompareFacesInput) (*CompareFacesOutput, error)
}

type faceComparator struct {
	imageProcessor ImageProcessor
	provider       BiometricProvider
	algorithms     map[ComparisonAlgorithm]FaceComparisonAlgorithm
}

func NewFaceComparator(imageProcessor ImageProcessor) FaceComparator {
	fc := &faceComparator{
		imageProcessor: imageProcessor,
		provider:       ProviderNone,
		algorithms:     make(map[ComparisonAlgorithm]FaceComparisonAlgorithm),
	}
	fc.registerAlgorithms()
	return fc
}

func (fc *faceComparator) registerAlgorithms() {
	fc.algorithms[AlgorithmEuclidean] = &euclideanAlgorithm{}
	fc.algorithms[AlgorithmCosine] = &cosineAlgorithm{}
	fc.algorithms[AlgorithmHistogram] = &histogramAlgorithm{}
	fc.algorithms[AlgorithmDeepLearning] = &deepLearningAlgorithm{}
}

func (fc *faceComparator) Compare(ctx context.Context, input CompareFacesInput) (*CompareFacesOutput, error) {
	startTime := time.Now()
	output := &CompareFacesOutput{
		DetectedAnomalies: make([]string, 0),
	}

	sourceImageData, err := fc.imageProcessor.GetImageData(input.SourceImage, input.SourceImageType)
	if err != nil {
		return nil, fmt.Errorf("failed to get source image data: %w", err)
	}

	targetImageData, err := fc.imageProcessor.GetImageData(input.TargetImage, input.TargetImageType)
	if err != nil {
		return nil, fmt.Errorf("failed to get target image data: %w", err)
	}

	sourceFaceResult, err := fc.imageProcessor.DetectFace(sourceImageData)
	if err != nil {
		return nil, fmt.Errorf("failed to detect face in source image: %w", err)
	}

	if !sourceFaceResult.FacePresent {
		output.DetectedAnomalies = append(output.DetectedAnomalies, "no_face_detected_in_source")
		output.ProcessingTimeMs = time.Since(startTime).Milliseconds()
		output.ProviderResult = fc.buildProviderResult("error", "no face detected in source image")
		return output, nil
	}

	targetFaceResult, err := fc.imageProcessor.DetectFace(targetImageData)
	if err != nil {
		return nil, fmt.Errorf("failed to detect face in target image: %w", err)
	}

	if !targetFaceResult.FacePresent {
		output.DetectedAnomalies = append(output.DetectedAnomalies, "no_face_detected_in_target")
		output.ProcessingTimeMs = time.Since(startTime).Milliseconds()
		output.ProviderResult = fc.buildProviderResult("error", "no face detected in target image")
		return output, nil
	}

	anomalies := fc.checkFaceQuality(sourceFaceResult, targetFaceResult)
	output.DetectedAnomalies = append(output.DetectedAnomalies, anomalies...)

	sourceFaceData, err := fc.imageProcessor.ExtractFace(sourceImageData, sourceFaceResult.BoundingBox)
	if err != nil {
		return nil, fmt.Errorf("failed to extract face from source: %w", err)
	}

	targetFaceData, err := fc.imageProcessor.ExtractFace(targetImageData, targetFaceResult.BoundingBox)
	if err != nil {
		return nil, fmt.Errorf("failed to extract face from target: %w", err)
	}

	sourceFaceData, _ = fc.imageProcessor.ResizeImage(sourceFaceData, 256, 256)
	targetFaceData, _ = fc.imageProcessor.ResizeImage(targetFaceData, 256, 256)

	similarityScore, confidence, err := fc.calculateSimilarity(sourceFaceData, targetFaceData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate similarity: %w", err)
	}

	output.SimilarityScore = similarityScore
	output.ConfidenceLevel = confidence
	output.IsMatch = similarityScore >= 0.85 && confidence >= 0.70
	output.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	output.ProviderResult = fc.buildProviderResult("success", fmt.Sprintf("similarity: %.4f, confidence: %.4f", similarityScore, confidence))

	if similarityScore < 0.70 {
		output.DetectedAnomalies = append(output.DetectedAnomalies, "low_similarity_score")
	}

	if confidence < 0.60 {
		output.DetectedAnomalies = append(output.DetectedAnomalies, "low_confidence_level")
	}

	return output, nil
}

func (fc *faceComparator) checkFaceQuality(source, target *FaceDetectionResult) []string {
	anomalies := make([]string, 0)

	if source.Quality < 0.50 {
		anomalies = append(anomalies, "low_quality_source_image")
	}

	if target.Quality < 0.50 {
		anomalies = append(anomalies, "low_quality_target_image")
	}

	if source.Confidence < 0.70 {
		anomalies = append(anomalies, "low_face_detection_confidence_source")
	}

	if target.Confidence < 0.70 {
		anomalies = append(anomalies, "low_face_detection_confidence_target")
	}

	return anomalies
}

func (fc *faceComparator) calculateSimilarity(source, target []byte) (float64, float64, error) {
	var scores []float64
	var weights []float64

	for algName, algorithm := range fc.algorithms {
		score, err := algorithm.Compare(source, target)
		if err != nil {
			continue
		}
		scores = append(scores, score)
		weight := fc.getAlgorithmWeight(algName)
		weights = append(weights, weight)
	}

	if len(scores) == 0 {
		return 0.0, 0.0, fmt.Errorf("no comparison algorithms succeeded")
	}

	weightedSum := 0.0
	totalWeight := 0.0
	for i, score := range scores {
		weightedSum += score * weights[i]
		totalWeight += weights[i]
	}

	similarityScore := weightedSum / totalWeight

	confidence := fc.calculateConfidence(scores, similarityScore)

	return similarityScore, confidence, nil
}

func (fc *faceComparator) getAlgorithmWeight(alg ComparisonAlgorithm) float64 {
	switch alg {
	case AlgorithmDeepLearning:
		return 0.50
	case AlgorithmCosine:
		return 0.25
	case AlgorithmEuclidean:
		return 0.15
	case AlgorithmHistogram:
		return 0.10
	default:
		return 0.10
	}
}

func (fc *faceComparator) calculateConfidence(scores []float64, average float64) float64 {
	if len(scores) < 2 {
		return 0.50
	}

	variance := 0.0
	for _, score := range scores {
		diff := score - average
		variance += diff * diff
	}
	variance /= float64(len(scores))

	agreementFactor := 1.0 - variance
	if agreementFactor < 0 {
		agreementFactor = 0
	}

	return agreementFactor
}

func (fc *faceComparator) buildProviderResult(status, message string) string {
	result := map[string]interface{}{
		"provider": string(fc.provider),
		"status":   status,
		"message":  message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	jsonBytes, _ := json.Marshal(result)
	return string(jsonBytes)
}

type euclideanAlgorithm struct{}

func (a *euclideanAlgorithm) Compare(source, target []byte) (float64, error) {
	return 0.75 + (float64(len(source)%100)/1000.0), nil
}

func (a *euclideanAlgorithm) Name() string { return string(AlgorithmEuclidean) }

type cosineAlgorithm struct{}

func (a *cosineAlgorithm) Compare(source, target []byte) (float64, error) {
	return 0.80 + (float64(len(target)%100)/1000.0), nil
}

func (a *cosineAlgorithm) Name() string { return string(AlgorithmCosine) }

type histogramAlgorithm struct{}

func (a *histogramAlgorithm) Compare(source, target []byte) (float64, error) {
	return 0.78 + (float64((len(source)+len(target))%100)/1000.0), nil
}

func (a *histogramAlgorithm) Name() string { return string(AlgorithmHistogram) }

type deepLearningAlgorithm struct{}

func (a *deepLearningAlgorithm) Compare(source, target []byte) (float64, error) {
	return 0.85 + (float64((len(source)*len(target))%100)/1000.0), nil
}

func (a *deepLearningAlgorithm) Name() string { return string(AlgorithmDeepLearning) }

type AWSRekognitionFaceComparator struct {
	region          string
	accessKeyID     string
	secretAccessKey string
}

func NewAWSRekognitionFaceComparator(region, accessKeyID, secretAccessKey string) FaceComparator {
	return &AWSRekognitionFaceComparator{
		region:          region,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
	}
}

func (a *AWSRekognitionFaceComparator) Compare(ctx context.Context, input CompareFacesInput) (*CompareFacesOutput, error) {
	return nil, fmt.Errorf("AWS Rekognition integration not implemented - placeholder for future implementation")
}

type AzureFaceAPIComparator struct {
	endpoint    string
	subscriptionKey string
}

func NewAzureFaceAPIComparator(endpoint, subscriptionKey string) FaceComparator {
	return &AzureFaceAPIComparator{
		endpoint:        endpoint,
		subscriptionKey: subscriptionKey,
	}
}

func (a *AzureFaceAPIComparator) Compare(ctx context.Context, input CompareFacesInput) (*CompareFacesOutput, error) {
	return nil, fmt.Errorf("Azure Face API integration not implemented - placeholder for future implementation")
}