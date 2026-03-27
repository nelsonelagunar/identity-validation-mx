package dto

type CURPValidationRequest struct {
	CURP   string `json:"curp" validate:"required,len=18"`
	UserID uint   `json:"user_id" validate:"required"`
}

type RFCValidationRequest struct {
	RFC    string `json:"rfc" validate:"required,min=12,max=13"`
	UserID uint   `json:"user_id" validate:"required"`
}

type INEValidationRequest struct {
	INEClave    string `json:"ine_clave" validate:"required,len=18"`
	UserID      uint   `json:"user_id" validate:"required"`
	OCRNumber   string `json:"ocr_number,omitempty"`
	ElectionKey string `json:"election_key,omitempty"`
}

type FacialComparisonRequest struct {
	UserID        uint   `json:"user_id" validate:"required"`
	DocumentPhoto string `json:"document_photo" validate:"required"`
	SelfiePhoto   string `json:"selfie_photo" validate:"required"`
}

type LivenessDetectionRequest struct {
	UserID     uint     `json:"user_id" validate:"required"`
	VideoFile  string   `json:"video_file,omitempty"`
	ImageFiles []string `json:"image_files,omitempty"`
}

type SignDocumentRequest struct {
	UserID         uint   `json:"user_id" validate:"required"`
	DocumentHash   string `json:"document_hash" validate:"required"`
	SignerName     string `json:"signer_name" validate:"required"`
	SignerRFCCURP  string `json:"signer_rfc_curp,omitempty"`
	SignatureType  string `json:"signature_type,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
}

type VerifySignatureRequest struct {
	SignatureID  uint   `json:"signature_id" validate:"required"`
	DocumentHash string `json:"document_hash" validate:"required"`
	Signature    string `json:"signature" validate:"required"`
}

type BulkImportRequest struct {
	UserID        uint                   `json:"user_id" validate:"required"`
	FileName      string                 `json:"file_name" validate:"required"`
	FileType      string                 `json:"file_type" validate:"required,oneof=csv xlsx json"`
	FileData      string                 `json:"file_data" validate:"required"`
	ValidationType string                 `json:"validation_type" validate:"required,oneof=CURP RFC INE BIOMETRIC SIGNATURE"`
	Options       map[string]interface{} `json:"options,omitempty"`
}