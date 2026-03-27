package dto

import "time"

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type HealthResponse struct {
	Status    string            `json:"status"`
 Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
 Checks    map[string]string  `json:"checks,omitempty"`
}

type ReadinessResponse struct {
	Status   string            `json:"status"`
	Checks   map[string]Check  `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

type CURPValidationResponse struct {
	RequestID         uint    `json:"request_id"`
	IsValid           bool    `json:"is_valid"`
	FullName          string  `json:"full_name,omitempty"`
	BirthDate         string  `json:"birth_date,omitempty"`
	Gender            string  `json:"gender,omitempty"`
	BirthState        string  `json:"birth_state,omitempty"`
	ValidationError   string  `json:"validation_error,omitempty"`
	RenapoVerified    bool    `json:"renapo_verified"`
	VerificationScore float64 `json:"verification_score"`
}

type RFCValidationResponse struct {
	RequestID         uint    `json:"request_id"`
	IsValid           bool    `json:"is_valid"`
	FullName          string  `json:"full_name,omitempty"`
	TaxRegime         string  `json:"tax_regime,omitempty"`
	RegistrationDate  string  `json:"registration_date,omitempty"`
	StatusSAT         string  `json:"status_sat,omitempty"`
	ValidationError   string  `json:"validation_error,omitempty"`
	SatVerified       bool    `json:"sat_verified"`
	VerificationScore float64 `json:"verification_score"`
}

type INEValidationResponse struct {
	RequestID         uint    `json:"request_id"`
	IsValid           bool    `json:"is_valid"`
	FullName          string  `json:"full_name,omitempty"`
	BirthDate         string  `json:"birth_date,omitempty"`
	Gender            string  `json:"gender,omitempty"`
	Address           string  `json:"address,omitempty"`
	VotingSection     string  `json:"voting_section,omitempty"`
	ValidationError   string  `json:"validation_error,omitempty"`
	INEVerified       bool    `json:"ine_verified"`
	VerificationScore float64 `json:"verification_score"`
}

type FacialComparisonResponse struct {
	RequestID         uint    `json:"request_id"`
	IsMatch           bool    `json:"is_match"`
	SimilarityScore   float64 `json:"similarity_score"`
	ConfidenceLevel   float64 `json:"confidence_level"`
	DetectedAnomalies string  `json:"detected_anomalies,omitempty"`
	ProcessingTime    int64   `json:"processing_time_ms"`
}

type LivenessDetectionResponse struct {
	RequestID       uint    `json:"request_id"`
	IsLive          bool    `json:"is_live"`
	LivenessScore   float64 `json:"liveness_score"`
	ConfidenceLevel float64 `json:"confidence_level"`
	SpoofProbability float64 `json:"spoof_probability"`
	DetectedAttacks  string  `json:"detected_attacks,omitempty"`
	ProcessingTime  int64   `json:"processing_time_ms"`
}

type SignDocumentResponse struct {
	RequestID       uint   `json:"request_id"`
	Signature       string `json:"signature,omitempty"`
	SerialNumber    string `json:"serial_number,omitempty"`
	IssuerDN        string `json:"issuer_dn,omitempty"`
	SubjectDN       string `json:"subject_dn,omitempty"`
	ValidFrom       string `json:"valid_from,omitempty"`
	ValidTo         string `json:"valid_to,omitempty"`
	SignatureBase64 string `json:"signature_base64,omitempty"`
	Certificate     string `json:"certificate,omitempty"`
}

type VerifySignatureResponse struct {
	RequestID         uint   `json:"request_id"`
	IsValid           bool   `json:"is_valid"`
	SignerVerified    bool   `json:"signer_verified"`
	DocumentIntegrity bool   `json:"document_integrity"`
	TimestampValid    bool   `json:"timestamp_valid"`
	ErrorCode         string `json:"error_code,omitempty"`
	ErrorMessage      string `json:"error_message,omitempty"`
	ProcessingTime    int64  `json:"processing_time_ms"`
}

type BulkImportResponse struct {
	JobID           string `json:"job_id"`
	Status          string `json:"status"`
	TotalRecords    int    `json:"total_records,omitempty"`
	Message         string `json:"message"`
}

type ImportStatusResponse struct {
	JobID            string `json:"job_id"`
	Status           string `json:"status"`
	TotalRecords     int    `json:"total_records"`
	ProcessedRecords int    `json:"processed_records"`
	SuccessRecords   int    `json:"success_records"`
	FailedRecords    int    `json:"failed_records"`
	StartedAt        string `json:"started_at,omitempty"`
	CompletedAt      string `json:"completed_at,omitempty"`
	ErrorMessage     string `json:"error_message,omitempty"`
}