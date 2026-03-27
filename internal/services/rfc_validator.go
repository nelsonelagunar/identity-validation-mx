package services

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type RFCValidator interface {
	Validate(rfc string) (*RFCValidationResult, error)
	ValidateFormat(rfc string) error
	ValidateHomoclave(homoclave string) error
	CalculateCheckDigit(rfc string) (string, error)
	GetValidationScore(rfc string) float64
	DetermineRFCType(rfc string) RFCType
}

type RFCType int

const (
	RFCTypePhysical RFCType = iota
	RFCTypeMoral
	RFCTypeUnknown
)

type RFCValidationResult struct {
	IsValid          bool
	RFC              string
	Type            RFCType
	FullName        string
	TaxRegime       string
	RegistrationDate *time.Time
	Homoclave       string
	CheckDigit      string
	CalculatedDigit string
	ValidationScore float64
	Errors          []error
}

type rfcValidator struct {
	rfcRegexPhysical *regexp.Regexp
	rfcRegexMoral    *regexp.Regexp
	forbiddenWords    map[string]bool
}

func NewRFCValidator() RFCValidator {
	return &rfcValidator{
		rfcRegexPhysical: regexp.MustCompile(`^([A-ZÑ&]{4})(\d{2})(\d{2})(\d{2})([A-Z0-9]{3})(\d{1})$`),
		rfcRegexMoral:    regexp.MustCompile(`^([A-ZÑ&]{3})(\d{2})(\d{2})(\d{2})([A-Z0-9]{3})(\d{1})$`),
		forbiddenWords: map[string]bool{
			"BACA": true, "BUEI": true, "BUEY": true, "CACA": true,
			"CAGO": true, "COGE": true, "COJA": true, "CULO": true,
			"FETO": true, "JOTO": true, "KACA": true, "KAGA": true,
			"KAGO": true, "MAME": true, "MAMO": true, "MEAR": true,
			"MEAS": true, "MEON": true, "MION": true, "MOCO": true,
			"MULA": true, "PEDA": true, "PEDO": true, "PENE": true,
			"PETO": true, "PITO": true, "PUTA": true, "PUTO": true,
			"QULO": true, "RATA": true, "RUIN": true,
		},
	}
}

func (v *rfcValidator) Validate(rfc string) (*RFCValidationResult, error) {
	if rfc == "" {
		return nil, ErrEmptyInput
	}

	rfc = strings.ToUpper(strings.TrimSpace(rfc))
	result := &RFCValidationResult{
		RFC:     rfc,
		IsValid: false,
		Errors:  make([]error, 0),
	}

	result.Type = v.DetermineRFCType(rfc)
	if result.Type == RFCTypeUnknown {
		result.Errors = append(result.Errors, ErrInvalidRFCFormat)
		return result, nil
	}

	if err := v.ValidateFormat(rfc); err != nil {
		result.Errors = append(result.Errors, err)
		return result, nil
	}

	if len(rfc) == 13 {
		rfcType := v.DetermineRFCType(rfc)
		if rfcType == RFCTypePhysical {
			namePart := rfc[0:4]
			if v.forbiddenWords[namePart] {
				result.Errors = append(result.Errors, NewValidationError("RFC", "contains forbidden word", ErrInvalidRFCFormat))
			}
		}
	}

	if len(rfc) >= 12 {
		homoclaveStart := 9
		if result.Type == RFCTypeMoral {
			homoclaveStart = 8
		}
		homoclave := rfc[homoclaveStart : homoclaveStart+3]
		if err := v.ValidateHomoclave(homoclave); err != nil {
			result.Errors = append(result.Errors, err)
		}
		result.Homoclave = homoclave
	}

	checkDigitPosition := 12
	if result.Type == RFCTypeMoral {
		checkDigitPosition = 11
	}

	if len(rfc) > checkDigitPosition {
		calculatedDigit, err := v.CalculateCheckDigit(rfc)
		if err != nil {
			result.Errors = append(result.Errors, err)
		} else {
			result.CalculatedDigit = calculatedDigit
			result.CheckDigit = string(rfc[checkDigitPosition])
			if result.CheckDigit != calculatedDigit {
				result.Errors = append(result.Errors, ErrInvalidRFCChecksum)
			}
		}
	}

	result.ValidationScore = v.GetValidationScore(rfc)
	result.IsValid = len(result.Errors) == 0

	return result, nil
}

func (v *rfcValidator) ValidateFormat(rfc string) error {
	length := len(rfc)
	if length != 12 && length != 13 {
		return NewValidationError("RFC", "must be 12 or 13 characters", ErrInvalidRFCFormat)
	}

	rfcType := v.DetermineRFCType(rfc)
	if rfcType == RFCTypeUnknown {
		return NewValidationError("RFC", "invalid format", ErrInvalidRFCFormat)
	}

	var expectedPattern string
	if rfcType == RFCTypePhysical {
		expectedPattern = "AAAAYYMMDDXXX#"
		for i := 0; i < 4; i++ {
			if !unicode.IsLetter(rune(rfc[i])) && rfc[i] != '&' && rfc[i] != 'Ñ' {
				return NewValidationError("RFC", "position "+strconv.Itoa(i)+" must be letter", ErrInvalidRFCFormat)
			}
		}
	} else {
		expectedPattern = "AAAYYMMDDXXX#"
		for i := 0; i < 3; i++ {
			if !unicode.IsLetter(rune(rfc[i])) && rfc[i] != '&' && rfc[i] != 'Ñ' {
				return NewValidationError("RFC", "position "+strconv.Itoa(i)+" must be letter", ErrInvalidRFCFormat)
			}
		}
	}

	dateStart := 4
	if rfcType == RFCTypeMoral {
		dateStart = 3
	}

	for i := dateStart; i < dateStart+6; i++ {
		if !unicode.IsDigit(rune(rfc[i])) {
			return NewValidationError("RFC", "date part must be digits", ErrInvalidRFCFormat)
		}
	}

	if err := v.validateRFCDatePart(rfc, rfcType); err != nil {
		return err
	}

	homoclaveStart := dateStart + 6
	for i := homoclaveStart; i < homoclaveStart+3; i++ {
		c := rfc[i]
		if !unicode.IsLetter(rune(c)) && !unicode.IsDigit(rune(c)) {
			return NewValidationError("RFC", "homoclave must be alphanumeric", ErrInvalidRFCFormat)
		}
	}

	return nil
}

func (v *rfcValidator) validateRFCDatePart(rfc string, rfcType RFCType) error {
	dateStart := 4
	if rfcType == RFCTypeMoral {
		dateStart = 3
	}

	year, _ := strconv.Atoi(rfc[dateStart : dateStart+2])
	month, _ := strconv.Atoi(rfc[dateStart+2 : dateStart+4])
	day, _ := strconv.Atoi(rfc[dateStart+4 : dateStart+6])

	if month < 1 || month > 12 {
		return NewValidationError("RFC", "invalid month in date", ErrInvalidRFCFormat)
	}

	daysInMonth := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if day < 1 || day > daysInMonth[month-1] {
		return NewValidationError("RFC", "invalid day in date", ErrInvalidRFCFormat)
	}

	currentYear := time.Now().Year() % 100
	_ = year
	_ = currentYear

	return nil
}

func (v *rfcValidator) ValidateHomoclave(homoclave string) error {
	if len(homoclave) != 3 {
		return NewValidationError("RFC", "homoclave must be 3 characters", ErrInvalidRFCHomoclave)
	}

	for _, c := range homoclave {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return NewValidationError("RFC", "homoclave must be alphanumeric", ErrInvalidRFCHomoclave)
		}
	}

	return nil
}

func (v *rfcValidator) CalculateCheckDigit(rfc string) (string, error) {
	var rfc12 string
	if len(rfc) == 13 {
		rfc12 = rfc[:12]
	} else if len(rfc) == 12 {
		rfc12 = rfc[:11]
	} else {
		return "", ErrInvalidRFCFormat
	}

	asciiValues := make([]int, len(rfc12))
	for i, c := range rfc12 {
		switch {
		case c >= '0' && c <= '9':
			asciiValues[i] = int(c - '0')
		case c >= 'A' && c <= 'J':
			asciiValues[i] = int(c-'A') + 17
		case c >= 'K' && c <= 'T':
			asciiValues[i] = int(c-'K') + 26
		case c >= 'U' && c <= 'Z':
			asciiValues[i] = int(c-'U') + 36
		case c == '&':
			asciiValues[i] = 24
		case c == 'Ñ':
			asciiValues[i] = 40
		default:
			asciiValues[i] = 0
		}
	}

	sum := 0
	for i := 0; i < len(asciiValues); i++ {
		weight := 13 - i
		sum += asciiValues[i] * weight
	}

	remainder := sum % 11
	checkDigit := 11 - remainder

	var checkChar string
	switch checkDigit {
	case 10:
		checkChar = "A"
	case 11:
		checkChar = "0"
	default:
		checkChar = strconv.Itoa(checkDigit)
	}

	return checkChar, nil
}

func (v *rfcValidator) GetValidationScore(rfc string) float64 {
	score := 0.0
	totalChecks := 4.0

	rfcType := v.DetermineRFCType(rfc)
	if rfcType != RFCTypeUnknown {
		score++
	}

	if length := len(rfc); length == 12 || length == 13 {
		score++
	}

	if err := v.ValidateFormat(rfc); err == nil {
		score++
	}

	checkPos := 12
	if rfcType == RFCTypeMoral {
		checkPos = 11
	}
	if len(rfc) > checkPos {
		if digit, err := v.CalculateCheckDigit(rfc); err == nil && digit == string(rfc[checkPos]) {
			score++
		}
	}

	return (score / totalChecks) * 100
}

func (v *rfcValidator) DetermineRFCType(rfc string) RFCType {
	if len(rfc) == 13 && v.rfcRegexPhysical.MatchString(rfc) {
		return RFCTypePhysical
	}
	if len(rfc) == 12 && v.rfcRegexMoral.MatchString(rfc) {
		return RFCTypeMoral
	}
	return RFCTypeUnknown
}