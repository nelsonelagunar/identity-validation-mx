package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCURPValidator struct {
	mock.Mock
}

func (m *MockCURPValidator) Validate(curp string) (*CURPValidationResult, error) {
	args := m.Called(curp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CURPValidationResult), args.Error(1)
}

func (m *MockCURPValidator) ValidateFormat(curp string) error {
	args := m.Called(curp)
	return args.Error(0)
}

func (m *MockCURPValidator) ValidateDate(curp string) error {
	args := m.Called(curp)
	return args.Error(0)
}

func (m *MockCURPValidator) ValidateState(stateCode string) error {
	args := m.Called(stateCode)
	return args.Error(0)
}

func (m *MockCURPValidator) CalculateCheckDigit(curp string) (string, error) {
	args := m.Called(curp)
	return args.String(0), args.Error(1)
}

func (m *MockCURPValidator) GetValidationScore(curp string) float64 {
	args := m.Called(curp)
	return args.Get(0).(float64)
}

type MockRFCValidator struct {
	mock.Mock
}

func (m *MockRFCValidator) Validate(rfc string) (*RFCValidationResult, error) {
	args := m.Called(rfc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RFCValidationResult), args.Error(1)
}

func (m *MockRFCValidator) ValidateFormat(rfc string) error {
	args := m.Called(rfc)
	return args.Error(0)
}

func (m *MockRFCValidator) ValidateHomoclave(homoclave string) error {
	args := m.Called(homoclave)
	return args.Error(0)
}

func (m *MockRFCValidator) CalculateCheckDigit(rfc string) (string, error) {
	args := m.Called(rfc)
	return args.String(0), args.Error(1)
}

func (m *MockRFCValidator) GetValidationScore(rfc string) float64 {
	args := m.Called(rfc)
	return args.Get(0).(float64)
}

func (m *MockRFCValidator) DetermineRFCType(rfc string) RFCType {
	args := m.Called(rfc)
	return args.Get(0).(RFCType)
}

type MockINEValidator struct {
	mock.Mock
}

func (m *MockINEValidator) Validate(ineClave string) (*INEValidationResult, error) {
	args := m.Called(ineClave)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*INEValidationResult), args.Error(1)
}

func (m *MockINEValidator) ValidateOCR(ocr string) error {
	args := m.Called(ocr)
	return args.Error(0)
}

func (m *MockINEValidator) ValidateElectionKey(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockINEValidator) ValidateCheckDigits(ineClave string) error {
	args := m.Called(ineClave)
	return args.Error(0)
}

func (m *MockINEValidator) GetValidationScore(ineClave string) float64 {
	args := m.Called(ineClave)
	return args.Get(0).(float64)
}

func TestIdentityService_ValidateCURP_ValidCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	validResult := &CURPValidationResult{
		IsValid:        true,
		CURP:           "GOPG900515HDFRRRA5",
		FullName:       "GARCIA OROZCO PEDRO",
		BirthDate:      &birthDate,
		Gender:         "M",
		BirthState:     "Ciudad de México",
		ValidationScore: 100.0,
		CheckDigit:     "5",
		Errors:         []error{},
	}

	mockCURPValidator.On("Validate", "GOPG900515HDFRRRA5").Return(validResult, nil)

	result, err := service.ValidateCURP("GOPG900515HDFRRRA5", 1)

	require.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Equal(t, "GOPG900515HDFRRRA5", result.CURP)
	assert.Equal(t, float64(100), result.VerificationScore)
	mockCURPValidator.AssertExpectations(t)
}

func TestIdentityService_ValidateCURP_InvalidCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("empty CURP", func(t *testing.T) {
		result, err := service.ValidateCURP("", 1)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrEmptyInput)
	})

	t.Run("invalid CURP format", func(t *testing.T) {
		invalidResult := &CURPValidationResult{
			IsValid: false,
			CURP:    "INVALID",
			Errors:  []error{ErrInvalidCURPFormat},
		}
		mockCURPValidator.On("Validate", "INVALID").Return(invalidResult, nil)

		result, err := service.ValidateCURP("INVALID", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
		assert.Contains(t, result.ValidationError, "invalid")
	})

	t.Run("wrong check digit", func(t *testing.T) {
		invalidResult := &CURPValidationResult{
			IsValid:        false,
			CURP:           "GOPG900515HDFRRRA0",
			ValidationScore: 80.0,
			Errors:         []error{ErrInvalidCURPChecksum},
		}
		mockCURPValidator.On("Validate", "GOPG900515HDFRRRA0").Return(invalidResult, nil)

		result, err := service.ValidateCURP("GOPG900515HDFRRRA0", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})
}

func TestIdentityService_ValidateCURP_EdgeCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("CURP with special characters", func(t *testing.T) {
		specialResult := &CURPValidationResult{
			IsValid: false,
			CURP:    "GOPG900515HDFRRR@5",
			Errors:  []error{ErrInvalidCURPFormat},
		}
		mockCURPValidator.On("Validate", "GOPG900515HDFRRR@5").Return(specialResult, nil)

		result, err := service.ValidateCURP("GOPG900515HDFRRR@5", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})

	t.Run("CURP with future date", func(t *testing.T) {
		futureDate := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
		futureResult := &CURPValidationResult{
			IsValid:   false,
			CURP:      "GOPG991231HDFRRRA5",
			BirthDate: &futureDate,
			Errors:    []error{ErrInvalidCURPDate},
		}
		mockCURPValidator.On("Validate", "GOPG991231HDFRRRA5").Return(futureResult, nil)

		result, err := service.ValidateCURP("GOPG991231HDFRRRA5", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})

	t.Run("CURP with leap year date", func(t *testing.T) {
		leapDate := time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC)
		leapResult := &CURPValidationResult{
			IsValid:   true,
			CURP:      "GOPG000229HDFRRRA5",
			BirthDate: &leapDate,
			Errors:    []error{},
		}
		mockCURPValidator.On("Validate", "GOPG000229HDFRRRA5").Return(leapResult, nil)

		result, err := service.ValidateCURP("GOPG000229HDFRRRA5", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
	})
}

func TestIdentityService_ValidateRFC_ValidCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("valid physical RFC", func(t *testing.T) {
		regDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)
		validResult := &RFCValidationResult{
			IsValid:          true,
			RFC:              "GOPG900515AB1",
			Type:             RFCTypePhysical,
			FullName:         "GARCIA OROZCO PEDRO",
			TaxRegime:        "605",
			RegistrationDate: &regDate,
			ValidationScore:  100.0,
			Errors:           []error{},
		}
		mockRFCValidator.On("Validate", "GOPG900515AB1").Return(validResult, nil)

		result, err := service.ValidateRFC("GOPG900515AB1", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Equal(t, "GOPG900515AB1", result.RFC)
	})

	t.Run("valid moral RFC", func(t *testing.T) {
		regDate := time.Date(2019, 6, 20, 0, 0, 0, 0, time.UTC)
		validResult := &RFCValidationResult{
			IsValid:          true,
			RFC:              "ABC900620ABC",
			Type:             RFCTypeMoral,
			FullName:         "EMPRESA SA DE CV",
			TaxRegime:        "601",
			RegistrationDate: &regDate,
			ValidationScore:  100.0,
			Errors:           []error{},
		}
		mockRFCValidator.On("Validate", "ABC900620ABC").Return(validResult, nil)

		result, err := service.ValidateRFC("ABC900620ABC", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
	})
}

func TestIdentityService_ValidateRFC_InvalidCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("empty RFC", func(t *testing.T) {
		result, err := service.ValidateRFC("", 1)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrEmptyInput)
	})

	t.Run("invalid RFC format", func(t *testing.T) {
		invalidResult := &RFCValidationResult{
			IsValid: false,
			RFC:     "INVALID",
			Type:    RFCTypeUnknown,
			Errors:  []error{ErrInvalidRFCFormat},
		}
		mockRFCValidator.On("Validate", "INVALID").Return(invalidResult, nil)

		result, err := service.ValidateRFC("INVALID", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})

	t.Run("wrong check digit", func(t *testing.T) {
		invalidResult := &RFCValidationResult{
			IsValid:        false,
			RFC:            "GOPG900515AB0",
			ValidationScore: 75.0,
			Errors:         []error{ErrInvalidRFCChecksum},
		}
		mockRFCValidator.On("Validate", "GOPG900515AB0").Return(invalidResult, nil)

		result, err := service.ValidateRFC("GOPG900515AB0", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})
}

func TestIdentityService_ValidateRFC_EdgeCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("RFC with special characters", func(t *testing.T) {
		specialResult := &RFCValidationResult{
			IsValid: false,
			RFC:     "GOPG900515@B1",
			Errors:  []error{ErrInvalidRFCFormat},
		}
		mockRFCValidator.On("Validate", "GOPG900515@B1").Return(specialResult, nil)

		result, err := service.ValidateRFC("GOPG900515@B1", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})

	t.Run("RFC with Ñ character", func(t *testing.T) {
		nResult := &RFCValidationResult{
			IsValid: true,
			RFC:     "NUÑZ900515AB1",
			Type:    RFCTypePhysical,
			Errors:  []error{},
		}
		mockRFCValidator.On("Validate", "NUÑZ900515AB1").Return(nResult, nil)

		result, err := service.ValidateRFC("NUÑZ900515AB1", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
	})

	t.Run("RFC with & character", func(t *testing.T) {
		ampResult := &RFCValidationResult{
			IsValid: true,
			RFC:     "&BPG900515AB1",
			Type:    RFCTypePhysical,
			Errors:  []error{},
		}
		mockRFCValidator.On("Validate", "&BPG900515AB1").Return(ampResult, nil)

		result, err := service.ValidateRFC("&BPG900515AB1", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
	})
}

func TestIdentityService_ValidateINE_ValidCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)

	t.Run("valid INE IFE format (18 chars)", func(t *testing.T) {
		validResult := &INEValidationResult{
			IsValid:         true,
			INEClave:        "GOGP900515HTSRRA05",
			INEType:         INETypeIFE,
			FullName:        "GARCIA OROZCO PEDRO",
			BirthDate:       &birthDate,
			Gender:          "M",
			VotingSection:   "1234",
			ValidationScore: 100.0,
			Errors:          []error{},
		}
		mockINEValidator.On("Validate", "GOGP900515HTSRRA05").Return(validResult, nil)

		result, err := service.ValidateINE("GOGP900515HTSRRA05", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Equal(t, "GOGP900515HTSRRA05", result.INEClave)
	})

	t.Run("valid INE format (19-20 chars)", func(t *testing.T) {
		validResult := &INEValidationResult{
			IsValid:         true,
			INEClave:        "GOGP900515HTSRRA050",
			INEType:         INETypeINE,
			FullName:        "GARCIA OROZCO PEDRO",
			BirthDate:       &birthDate,
			Gender:          "M",
			VotingSection:   "12345",
			ValidationScore: 100.0,
			Errors:          []error{},
		}
		mockINEValidator.On("Validate", "GOGP900515HTSRRA050").Return(validResult, nil)

		result, err := service.ValidateINE("GOGP900515HTSRRA050", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
	})
}

func TestIdentityService_ValidateINE_InvalidCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("empty INE", func(t *testing.T) {
		result, err := service.ValidateINE("", 1)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrEmptyInput)
	})

	t.Run("invalid INE format", func(t *testing.T) {
		invalidResult := &INEValidationResult{
			IsValid:  false,
			INEClave: "INVALID",
			Errors:   []error{ErrInvalidINEFormat},
		}
		mockINEValidator.On("Validate", "INVALID").Return(invalidResult, nil)

		result, err := service.ValidateINE("INVALID", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})

	t.Run("wrong check digits", func(t *testing.T) {
		invalidResult := &INEValidationResult{
			IsValid:         false,
			INEClave:        "GOGP900515HTSRRA00",
			ValidationScore: 75.0,
			Errors:          []error{ErrInvalidINEChecksum},
		}
		mockINEValidator.On("Validate", "GOGP900515HTSRRA00").Return(invalidResult, nil)

		result, err := service.ValidateINE("GOGP900515HTSRRA00", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})
}

func TestIdentityService_ValidateINE_EdgeCases(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("INE with special characters", func(t *testing.T) {
		specialResult := &INEValidationResult{
			IsValid:  false,
			INEClave: "GOGP900515H@SRRA05",
			Errors:   []error{ErrInvalidINEFormat},
		}
		mockINEValidator.On("Validate", "GOGP900515H@SRRA05").Return(specialResult, nil)

		result, err := service.ValidateINE("GOGP900515H@SRRA05", 1)

		require.NoError(t, err)
		assert.False(t, result.IsValid)
	})

	t.Run("INE with leap year date", func(t *testing.T) {
		leapDate := time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC)
		leapResult := &INEValidationResult{
			IsValid:   true,
			INEClave:  "GOGP000229HTSRRA05",
			BirthDate: &leapDate,
			Errors:    []error{},
		}
		mockINEValidator.On("Validate", "GOGP000229HTSRRA05").Return(leapResult, nil)

		result, err := service.ValidateINE("GOGP000229HTSRRA05", 1)

		require.NoError(t, err)
		assert.True(t, result.IsValid)
	})
}

func TestIdentityService_ValidateIdentity_FullValidation(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)

	curpResult := &CURPValidationResult{
		IsValid:        true,
		CURP:           "GOPG900515HDFRRRA5",
		BirthDate:      &birthDate,
		ValidationScore: 100.0,
		Errors:         []error{},
	}

	rfcResult := &RFCValidationResult{
		IsValid:        true,
		RFC:            "GOPG900515AB1",
		Type:           RFCTypePhysical,
		BirthDate:      &birthDate,
		ValidationScore: 100.0,
		Errors:         []error{},
	}

	ineResult := &INEValidationResult{
		IsValid:         true,
		INEClave:        "GOGP900515HTSRRA05",
		BirthDate:       &birthDate,
		ValidationScore: 100.0,
		Errors:          []error{},
	}

	mockCURPValidator.On("Validate", "GOPG900515HDFRRRA5").Return(curpResult, nil)
	mockRFCValidator.On("Validate", "GOPG900515AB1").Return(rfcResult, nil)
	mockINEValidator.On("Validate", "GOGP900515HTSRRA05").Return(ineResult, nil)

	result, err := service.ValidateIdentity("GOPG900515HDFRRRA5", "GOPG900515AB1", "GOGP900515HTSRRA05", 1)

	require.NoError(t, err)
	assert.True(t, result.CURPValid)
	assert.True(t, result.RFCValid)
	assert.True(t, result.INEValid)
	assert.Equal(t, 100.0, result.OverallScore)
	assert.Empty(t, result.Errors)
}

func TestIdentityService_ValidateIdentity_PartialValidation(t *testing.T) {
	mockCURPValidator := new(MockCURPValidator)
	mockRFCValidator := new(MockRFCValidator)
	mockINEValidator := new(MockINEValidator)

	service := NewIdentityServiceWithValidators(mockCURPValidator, mockRFCValidator, mockINEValidator)

	t.Run("only CURP provided", func(t *testing.T) {
		birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		curpResult := &CURPValidationResult{
			IsValid:        true,
			CURP:           "GOPG900515HDFRRRA5",
			BirthDate:      &birthDate,
			ValidationScore: 100.0,
			Errors:         []error{},
		}

		mockCURPValidator.On("Validate", "GOPG900515HDFRRRA5").Return(curpResult, nil)

		result, err := service.ValidateIdentity("GOPG900515HDFRRRA5", "", "", 1)

		require.NoError(t, err)
		assert.True(t, result.CURPValid)
		assert.False(t, result.RFCValid)
		assert.False(t, result.INEValid)
	})

	t.Run("CURP and RFC mismatched birth dates", func(t *testing.T) {
		curpBirthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		rfcBirthDate := time.Date(1985, 3, 20, 0, 0, 0, 0, time.UTC)

		curpResult := &CURPValidationResult{
			IsValid:        true,
			CURP:           "GOPG900515HDFRRRA5",
			BirthDate:      &curpBirthDate,
			ValidationScore: 100.0,
			Errors:         []error{},
		}

		rfcResult := &RFCValidationResult{
			IsValid:        true,
			RFC:            "GOPG850320AB1",
			Type:           RFCTypePhysical,
			BirthDate:      &rfcBirthDate,
			ValidationScore: 100.0,
			Errors:         []error{},
		}

		mockCURPValidator.On("Validate", "GOPG900515HDFRRRA5").Return(curpResult, nil)
		mockRFCValidator.On("Validate", "GOPG850320AB1").Return(rfcResult, nil)

		result, err := service.ValidateIdentity("GOPG900515HDFRRRA5", "GOPG850320AB1", "", 1)

		require.NoError(t, err)
		assert.NotEmpty(t, result.Errors)
	})
}

func TestNewIdentityService(t *testing.T) {
	service := NewIdentityService()
	assert.NotNil(t, service)
}