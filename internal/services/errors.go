package services

import "errors"

var (
	ErrInvalidCURPFormat      = errors.New("invalid CURP format")
	ErrInvalidCURPChecksum    = errors.New("invalid CURP check digit")
	ErrInvalidCURPDate        = errors.New("invalid date in CURP")
	ErrInvalidCURPState       = errors.New("invalid state code in CURP")
	ErrInvalidRFCFormat       = errors.New("invalid RFC format")
	ErrInvalidRFCChecksum     = errors.New("invalid RFC check digit")
	ErrInvalidRFCHomoclave    = errors.New("invalid RFC homoclave")
	ErrInvalidINEFormat       = errors.New("invalid INE format")
	ErrInvalidINEOCR          = errors.New("invalid INE OCR code")
	ErrInvalidINEElectionKey  = errors.New("invalid INE election key")
	ErrInvalidINEChecksum     = errors.New("invalid INE check digit")
	ErrValidationFailed       = errors.New("validation failed")
	ErrEmptyInput             = errors.New("input cannot be empty")
)

type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return e.Field + ": " + e.Message + " - " + e.Err.Error()
	}
	return e.Field + ": " + e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}