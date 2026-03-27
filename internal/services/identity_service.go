package services

import (
	"time"

	"github.com/nelsonelagunar/identity-validation-mx/internal/models"
)

type IdentityService interface {
	ValidateCURP(curp string, userID uint) (*models.CURPValidationResponse, error)
	ValidateRFC(rfc string, userID uint) (*models.RFCValidationResponse, error)
	ValidateINE(ineClave string, userID uint) (*models.INEValidationResponse, error)
	ValidateIdentity(curp, rfc, ineClave string, userID uint) (*IdentityValidationResult, error)
}

type IdentityValidationResult struct {
	CURPValid            bool
	RFCValid             bool
	INEValid             bool
	OverallScore         float64
	CURPValidationScore  float64
	RFCValidationScore   float64
	INEValidationScore   float64
	CURPResponse         *models.CURPValidationResponse
	RFCResponse          *models.RFCValidationResponse
	INEResponse          *models.INEValidationResponse
	Errors               []error
}

type identityService struct {
	curpValidator CURPValidator
	rfcValidator   RFCValidator
	ineValidator   INEValidator
}

func NewIdentityService() IdentityService {
	return &identityService{
		curpValidator: NewCURPValidator(),
		rfcValidator:  NewRFCValidator(),
		ineValidator:  NewINEValidator(),
	}
}

func NewIdentityServiceWithValidators(curpValidator CURPValidator, rfcValidator RFCValidator, ineValidator INEValidator) IdentityService {
	return &identityService{
		curpValidator: curpValidator,
		rfcValidator:  rfcValidator,
		ineValidator:  ineValidator,
	}
}

func (s *identityService) ValidateCURP(curp string, userID uint) (*models.CURPValidationResponse, error) {
	if curp == "" {
		return nil, ErrEmptyInput
	}

	result, err := s.curpValidator.Validate(curp)
	if err != nil {
		return nil, err
	}

	response := &models.CURPValidationResponse{
		IsValid:          result.IsValid,
		FullName:         result.FullName,
		BirthDate:        result.BirthDate,
		Gender:           result.Gender,
		BirthState:       result.BirthState,
		VerificationScore: result.ValidationScore,
		RenapoVerified:    false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if !result.IsValid && len(result.Errors) > 0 {
		errorMessages := make([]string, 0, len(result.Errors))
		for _, e := range result.Errors {
			errorMessages = append(errorMessages, e.Error())
		}
		response.ValidationError = errorMessages[0]
	}

	return response, nil
}

func (s *identityService) ValidateRFC(rfc string, userID uint) (*models.RFCValidationResponse, error) {
	if rfc == "" {
		return nil, ErrEmptyInput
	}

	result, err := s.rfcValidator.Validate(rfc)
	if err != nil {
		return nil, err
	}

	response := &models.RFCValidationResponse{
		IsValid:          result.IsValid,
		FullName:         result.FullName,
		TaxRegime:        result.TaxRegime,
		RegistrationDate: result.RegistrationDate,
		VerificationScore: result.ValidationScore,
		SatVerified:      false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if !result.IsValid && len(result.Errors) > 0 {
		errorMessages := make([]string, 0, len(result.Errors))
		for _, e := range result.Errors {
			errorMessages = append(errorMessages, e.Error())
		}
		response.ValidationError = errorMessages[0]
	}

	return response, nil
}

func (s *identityService) ValidateINE(ineClave string, userID uint) (*models.INEValidationResponse, error) {
	if ineClave == "" {
		return nil, ErrEmptyInput
	}

	result, err := s.ineValidator.Validate(ineClave)
	if err != nil {
		return nil, err
	}

	response := &models.INEValidationResponse{
		IsValid:          result.IsValid,
		FullName:         result.FullName,
		BirthDate:        result.BirthDate,
		Gender:           result.Gender,
		VotingSection:    result.VotingSection,
		VerificationScore: result.ValidationScore,
		INEVerified:      false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if !result.IsValid && len(result.Errors) > 0 {
		errorMessages := make([]string, 0, len(result.Errors))
		for _, e := range result.Errors {
			errorMessages = append(errorMessages, e.Error())
		}
		response.ValidationError = errorMessages[0]
	}

	return response, nil
}

func (s *identityService) ValidateIdentity(curp, rfc, ineClave string, userID uint) (*IdentityValidationResult, error) {
	result := &IdentityValidationResult{
		Errors: make([]error, 0),
	}

	var curpErr, rfcErr, ineErr error
	var curpResult *CURPValidationResult
	var rfcResult *RFCValidationResult
	var ineResult *INEValidationResult

	if curp != "" {
		curpResult, curpErr = s.curpValidator.Validate(curp)
		if curpErr != nil {
			result.Errors = append(result.Errors, curpErr)
		} else {
			result.CURPValid = curpResult.IsValid
			result.CURPValidationScore = curpResult.ValidationScore
			result.CURPResponse, _ = s.ValidateCURP(curp, userID)
		}
	}

	if rfc != "" {
		rfcResult, rfcErr = s.rfcValidator.Validate(rfc)
		if rfcErr != nil {
			result.Errors = append(result.Errors, rfcErr)
		} else {
			result.RFCValid = rfcResult.IsValid
			result.RFCValidationScore = rfcResult.ValidationScore
			result.RFCResponse, _ = s.ValidateRFC(rfc, userID)
		}
	}

	if ineClave != "" {
		ineResult, ineErr = s.ineValidator.Validate(ineClave)
		if ineErr != nil {
			result.Errors = append(result.Errors, ineErr)
		} else {
			result.INEValid = ineResult.IsValid
			result.INEValidationScore = ineResult.ValidationScore
			result.INEResponse, _ = s.ValidateINE(ineClave, userID)
		}
	}

	totalValid := 0
	totalScore := 0.0
	count := 0

	if curp != "" {
		count++
		if result.CURPValid {
			totalValid++
		}
		totalScore += result.CURPValidationScore
	}

	if rfc != "" {
		count++
		if result.RFCValid {
			totalValid++
		}
		totalScore += result.RFCValidationScore
	}

	if ineClave != "" {
		count++
		if result.INEValid {
			totalValid++
		}
		totalScore += result.INEValidationScore
	}

	if count > 0 {
		result.OverallScore = totalScore / float64(count)
	} else {
		result.OverallScore = 0
	}

	if curp != "" && rfc != "" && curpResult != nil && rfcResult != nil {
		if curpResult.IsValid && rfcResult.IsValid {
			if curpResult.BirthDate != nil && rfcResult.BirthDate != nil {
				curpDate := *curpResult.BirthDate
				rfcDate := *rfcResult.BirthDate
				if !curpDate.Equal(rfcDate) {
					validationErr := NewValidationError("CURP/RFC", "birth dates do not match", ErrValidationFailed)
					result.Errors = append(result.Errors, validationErr)
				}
			}
		}
	}

	return result, nil
}