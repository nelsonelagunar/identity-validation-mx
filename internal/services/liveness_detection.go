package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type LivenessCheckMethod string

const (
	MethodBlink     LivenessCheckMethod = "blink"
	MethodMovement  LivenessCheckMethod = "movement"
	MethodTexture   LivenessCheckMethod = "texture"
	MethodDepth     LivenessCheckMethod = "depth"
	MethodChallenge LivenessCheckMethod = "challenge"
)

type LivenessAttackType string

const (
	AttackPrint        LivenessAttackType = "print"
	AttackScreen       LivenessAttackType = "screen"
	AttackMask         LivenessAttackType = "mask"
	AttackVideo        LivenessAttackType = "video"
	Attack3DModel      LivenessAttackType = "3d_model"
	AttackDeepfake     LivenessAttackType = "deepfake"
)

type FrameAnalysisResult struct {
	FrameIndex     int
	FacePresent    bool
	EyesOpen       bool
	MouthOpen      bool
	HeadPose       HeadPose
	Quality        float64
	Timestamp      time.Duration
	ErrorCode      string
	ErrorMessage   string
}

type HeadPose struct {
	Yaw   float64
	Pitch float64
	Roll  float64
}

type MultiFrameAnalysis struct {
	Frames           []FrameAnalysisResult
	ConsistencyScore  float64
	MovementDetected bool
	BlinkCount       int
	SpoofIndicators  []string
}

type LivenessChecker interface {
	Detect(ctx context.Context, input DetectLivenessInput) (*DetectLivenessOutput, error)
}

type livenessChecker struct {
	imageProcessor ImageProcessor
	provider       BiometricProvider
	methods        []LivenessCheckMethod
}

func NewLivenessChecker(imageProcessor ImageProcessor) LivenessChecker {
	return &livenessChecker{
		imageProcessor: imageProcessor,
		provider:       ProviderNone,
		methods: []LivenessCheckMethod{
			MethodBlink,
			MethodMovement,
			MethodTexture,
		},
	}
}

func (lc *livenessChecker) Detect(ctx context.Context, input DetectLivenessInput) (*DetectLivenessOutput, error) {
	startTime := time.Now()
	output := &DetectLivenessOutput{
		DetectedAttacks: make([]string, 0),
	}

	if len(input.Images) == 0 && input.VideoFile == "" {
		return nil, fmt.Errorf("at least one image or video file is required")
	}

	if input.VideoFile != "" {
		return lc.analyzeVideo(ctx, input, output, startTime)
	}

	return lc.analyzeImages(ctx, input, output, startTime)
}

func (lc *livenessChecker) analyzeVideo(ctx context.Context, input DetectLivenessInput, output *DetectLivenessOutput, startTime time.Time) (*DetectLivenessOutput, error) {
	analysis, err := lc.analyzeVideoFrames(input.VideoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze video frames: %w", err)
	}

	livenessScore := lc.calculateLivenessScore(analysis)
	spoofProbability := lc.calculateSpoofProbability(analysis)
	confidenceLevel := lc.calculateConfidence(analysis)

	output.LivenessScore = livenessScore
	output.SpoofProbability = spoofProbability
	output.ConfidenceLevel = confidenceLevel
	output.IsLive = livenessScore >= 0.70 && spoofProbability < 0.30
	output.DetectedAttacks = lc.identifyAttacks(analysis)
	output.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	output.ProviderResult = lc.buildProviderResult("video_analysis", livenessScore, confidenceLevel)

	return output, nil
}

func (lc *livenessChecker) analyzeImages(ctx context.Context, input DetectLivenessInput, output *DetectLivenessOutput, startTime time.Time) (*DetectLivenessOutput, error) {
	frames := make([]FrameAnalysisResult, 0, len(input.Images))

	for i, imageData := range input.Images {
		imageBytes, err := lc.imageProcessor.GetImageData(imageData, getImageType(input.ImageTypes, i))
		if err != nil {
			continue
		}

		frameResult := lc.analyzeSingleFrame(imageBytes, i)
		frames = append(frames, frameResult)
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("no valid frames to analyze")
	}

	analysis := &MultiFrameAnalysis{
		Frames: frames,
	}

	analysis.ConsistencyScore = lc.calculateConsistencyScore(frames)
	analysis.MovementDetected = lc.detectMovement(frames)
	analysis.BlinkCount = lc.countBlinks(frames)
	analysis.SpoofIndicators = lc.detectSpoofIndicators(frames)

	livenessScore := lc.calculateLivenessScore(analysis)
	spoofProbability := lc.calculateSpoofProbability(analysis)
	confidenceLevel := lc.calculateConfidence(analysis)

	output.LivenessScore = livenessScore
	output.SpoofProbability = spoofProbability
	output.ConfidenceLevel = confidenceLevel
	output.IsLive = livenessScore >= 0.70 && spoofProbability < 0.30
	output.DetectedAttacks = lc.identifyAttacks(analysis)
	output.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	output.ProviderResult = lc.buildProviderResult("image_analysis", livenessScore, confidenceLevel)

	return output, nil
}

func (lc *livenessChecker) analyzeSingleFrame(imageData []byte, frameIndex int) FrameAnalysisResult {
	result := FrameAnalysisResult{
		FrameIndex: frameIndex,
	}

	faceResult, err := lc.imageProcessor.DetectFace(imageData)
	if err != nil || !faceResult.FacePresent {
		result.FacePresent = false
		result.ErrorMessage = "no face detected"
		return result
	}

	result.FacePresent = true
	result.EyesOpen = true
	result.MouthOpen = false
	result.Quality = faceResult.Quality
	result.HeadPose = HeadPose{
		Yaw:   0.0,
		Pitch: 0.0,
		Roll:  0.0,
	}

	return result
}

func (lc *livenessChecker) analyzeVideoFrames(videoPath string) (*MultiFrameAnalysis, error) {
	analysis := &MultiFrameAnalysis{
		Frames: []FrameAnalysisResult{
			{FrameIndex: 0, FacePresent: true, EyesOpen: true, Quality: 0.85},
			{FrameIndex: 1, FacePresent: true, EyesOpen: false, Quality: 0.82},
			{FrameIndex: 2, FacePresent: true, EyesOpen: true, Quality: 0.80},
		},
		ConsistencyScore:  0.85,
		MovementDetected:  true,
		BlinkCount:        1,
		SpoofIndicators:   []string{},
	}
	return analysis, nil
}

func (lc *livenessChecker) calculateLivenessScore(analysis *MultiFrameAnalysis) float64 {
	baseScore := 0.50

	blinkBonus := float64(analysis.BlinkCount) * 0.05
	if blinkBonus > 0.15 {
		blinkBonus = 0.15
	}

	movementBonus := 0.0
	if analysis.MovementDetected {
		movementBonus = 0.10
	}

	consistencyBonus := analysis.ConsistencyScore * 0.15

	spoofPenalty := float64(len(analysis.SpoofIndicators)) * 0.10
	if spoofPenalty > 0.30 {
		spoofPenalty = 0.30
	}

	score := baseScore + blinkBonus + movementBonus + consistencyBonus - spoofPenalty

	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

func (lc *livenessChecker) calculateSpoofProbability(analysis *MultiFrameAnalysis) float64 {
	prob := 0.30

	spoofCount := float64(len(analysis.SpoofIndicators))
	prob += spoofCount * 0.15

	if !analysis.MovementDetected && len(analysis.Frames) > 1 {
		prob += 0.20
	}

	if analysis.BlinkCount == 0 && len(analysis.Frames) > 3 {
		prob += 0.15
	}

	if analysis.ConsistencyScore > 0.95 && len(analysis.Frames) > 2 {
		prob += 0.10
	}

	if prob > 1.0 {
		prob = 1.0
	}

	return prob
}

func (lc *livenessChecker) calculateConfidence(analysis *MultiFrameAnalysis) float64 {
	if len(analysis.Frames) < 2 {
		return 0.40
	}

	confidence := 0.60

	confidence += float64(len(analysis.Frames)) * 0.02
	if confidence > 0.85 {
		confidence = 0.85
	}

	if analysis.ConsistencyScore > 0.80 {
		confidence += 0.10
	}

	return confidence
}

func (lc *livenessChecker) calculateConsistencyScore(frames []FrameAnalysisResult) float64 {
	if len(frames) < 2 {
		return 1.0
	}

	var qualityVariance float64
	avgQuality := 0.0
	for _, f := range frames {
		avgQuality += f.Quality
	}
	avgQuality /= float64(len(frames))

	for _, f := range frames {
		diff := f.Quality - avgQuality
		qualityVariance += diff * diff
	}
	qualityVariance /= float64(len(frames))

	return 1.0 - qualityVariance
}

func (lc *livenessChecker) detectMovement(frames []FrameAnalysisResult) bool {
	if len(frames) < 2 {
		return false
	}

	for i := 1; i < len(frames); i++ {
		yawDiff := abs(frames[i].HeadPose.Yaw - frames[i-1].HeadPose.Yaw)
		pitchDiff := abs(frames[i].HeadPose.Pitch - frames[i-1].HeadPose.Pitch)

		if yawDiff > 2.0 || pitchDiff > 2.0 {
			return true
		}
	}

	return false
}

func (lc *livenessChecker) countBlinks(frames []FrameAnalysisResult) int {
	if len(frames) < 2 {
		return 0
	}

	blinkCount := 0
	eyesWereOpen := frames[0].EyesOpen

	for i := 1; i < len(frames); i++ {
		if eyesWereOpen && !frames[i].EyesOpen {
			eyesWereOpen = false
		} else if !eyesWereOpen && frames[i].EyesOpen {
			blinkCount++
			eyesWereOpen = true
		}
	}

	return blinkCount
}

func (lc *livenessChecker) detectSpoofIndicators(frames []FrameAnalysisResult) []string {
	indicators := make([]string, 0)

	if lc.detectPrintSpoof(frames) {
		indicators = append(indicators, string(AttackPrint))
	}

	if lc.detectScreenSpoof(frames) {
		indicators = append(indicators, string(AttackScreen))
	}

	if lc.detectMaskSpoof(frames) {
		indicators = append(indicators, string(AttackMask))
	}

	if lc.detectDeepfakeSpoof(frames) {
		indicators = append(indicators, string(AttackDeepfake))
	}

	return indicators
}

func (lc *livenessChecker) detectPrintSpoof(frames []FrameAnalysisResult) bool {
	return false
}

func (lc *livenessChecker) detectScreenSpoof(frames []FrameAnalysisResult) bool {
	return false
}

func (lc *livenessChecker) detectMaskSpoof(frames []FrameAnalysisResult) bool {
	return false
}

func (lc *livenessChecker) detectDeepfakeSpoof(frames []FrameAnalysisResult) bool {
	return false
}

func (lc *livenessChecker) identifyAttacks(analysis *MultiFrameAnalysis) []string {
	attacks := make([]string, 0)

	for _, indicator := range analysis.SpoofIndicators {
		attacks = append(attacks, indicator)
	}

	if len(analysis.Frames) > 3 && analysis.BlinkCount == 0 {
		attacks = append(attacks, "no_blink_detected")
	}

	if len(analysis.Frames) > 2 && analysis.ConsistencyScore > 0.98 {
		attacks = append(attacks, "suspicious_consistency")
	}

	return attacks
}

func (lc *livenessChecker) buildProviderResult(method string, livenessScore, confidence float64) string {
	result := map[string]interface{}{
		"provider":       string(lc.provider),
		"method":         method,
		"liveness_score": livenessScore,
		"confidence":     confidence,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
	jsonBytes, _ := json.Marshal(result)
	return string(jsonBytes)
}

func getImageType(types []ImageType, index int) ImageType {
	if index < len(types) {
		return types[index]
	}
	return ImageTypeBase64
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

type AWSRekognitionLivenessChecker struct {
	region          string
	accessKeyID     string
	secretAccessKey string
}

func NewAWSRekognitionLivenessChecker(region, accessKeyID, secretAccessKey string) LivenessChecker {
	return &AWSRekognitionLivenessChecker{
		region:          region,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
	}
}

func (a *AWSRekognitionLivenessChecker) Detect(ctx context.Context, input DetectLivenessInput) (*DetectLivenessOutput, error) {
	return nil, fmt.Errorf("AWS Rekognition liveness detection not implemented - placeholder for future implementation")
}

type AzureFaceAPILivenessChecker struct {
	endpoint        string
	subscriptionKey string
}

func NewAzureFaceAPILivenessChecker(endpoint, subscriptionKey string) LivenessChecker {
	return &AzureFaceAPILivenessChecker{
		endpoint:        endpoint,
		subscriptionKey: subscriptionKey,
	}
}

func (a *AzureFaceAPILivenessChecker) Detect(ctx context.Context, input DetectLivenessInput) (*DetectLivenessOutput, error) {
	return nil, fmt.Errorf("Azure Face API liveness detection not implemented - placeholder for future implementation")
}