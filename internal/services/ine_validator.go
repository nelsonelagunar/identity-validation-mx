package services

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type INEValidator interface {
	Validate(ineClave string) (*INEValidationResult, error)
	ValidateOCR(ocr string) error
	ValidateElectionKey(key string) error
	ValidateCheckDigits(ineClave string) error
	GetValidationScore(ineClave string) float64
}

type INEType int

const (
	INETypeIFE INEType = iota
	INETypeINE
	INETypeUnknown
)

type INEValidationResult struct {
	IsValid          bool
	INEClave         string
	INEType          INEType
	OCRNumber       string
	ElectionKey     string
	FullName        string
	BirthDate       *time.Time
	Gender          string
	VotingSection   string
	ValidationScore float64
	Errors          []error
}

type ineValidator struct {
	ocrRegex       *regexp.Regexp
	electionRegex  *regexp.Regexp
	claveRegex     *regexp.Regexp
	stateCodes     map[string]string
}

func NewINEValidator() INEValidator {
	return &ineValidator{
		ocrRegex:      regexp.MustCompile(`^\d{13}$`),
		electionRegex: regexp.MustCompile(`^[A-Z]{6}\d{8}[A-Z]\d$`),
		claveRegex:    regexp.MustCompile(`^[A-Z]{6}\d{8}[A-Z]\d{2,3}$`),
		stateCodes: map[string]string{
			"AG": "Aguascalientes", "BS": "Baja California Sur", "BC": "Baja California",
			"CM": "Campeche", "CS": "Chiapas", "CH": "Chihuahua", "CL": "Coahuila",
			"CO": "Colima", "DF": "Ciudad de México", "DG": "Durango", "GT": "Guanajuato",
			"GR": "Guerrero", "HG": "Hidalgo", "JA": "Jalisco", "MC": "México",
			"MN": "Michoacán", "MS": "Morelos", "NT": "Nayarit", "NL": "Nuevo León",
			"OC": "Oaxaca", "PL": "Puebla", "QO": "Querétaro", "QR": "Quintana Roo",
			"SP": "San Luis Potosí", "SL": "Sinaloa", "SR": "Sonora", "TB": "Tabasco",
			"TC": "Tlaxcala", "TS": "Tamaulipas", "VZ": "Veracruz", "YN": "Yucatán",
			"ZS": "Zacatecas",
		},
	}
}

func (v *ineValidator) Validate(ineClave string) (*INEValidationResult, error) {
	if ineClave == "" {
		return nil, ErrEmptyInput
	}

	ineClave = strings.ToUpper(strings.TrimSpace(ineClave))
	result := &INEValidationResult{
		INEClave: ineClave,
		IsValid:  false,
		Errors:   make([]error, 0),
	}

	result.INEType = v.determineINEType(ineClave)
	if result.INEType == INETypeUnknown {
		result.Errors = append(result.Errors, NewValidationError("INE", "unknown INE format", ErrInvalidINEFormat))
	}

	if len(ineClave) < 18 || len(ineClave) > 20 {
		result.Errors = append(result.Errors, NewValidationError("INE", "invalid length", ErrInvalidINEFormat))
		return result, nil
	}

	for i, c := range ineClave {
		switch {
		case i < 6:
			if !unicode.IsLetter(c) {
				result.Errors = append(result.Errors, NewValidationError("INE", "positions 0-5 must be letters", ErrInvalidINEFormat))
			}
		case i >= 6 && i < 8:
			if !unicode.IsDigit(c) {
				result.Errors = append(result.Errors, NewValidationError("INE", "positions 6-7 must be digits (year)", ErrInvalidINEFormat))
			}
		case i == 8 || i == 9:
			if !unicode.IsDigit(c) {
				result.Errors = append(result.Errors, NewValidationError("INE", "positions 8-9 must be digits (month)", ErrInvalidINEFormat))
			}
		case i == 10 || i == 11:
			if !unicode.IsDigit(c) {
				result.Errors = append(result.Errors, NewValidationError("INE", "positions 10-11 must be digits (day)", ErrInvalidINEFormat))
			}
		case i == 12:
			if !unicode.IsLetter(c) {
				result.Errors = append(result.Errors, NewValidationError("INE", "position 12 must be letter (gender)", ErrInvalidINEFormat))
			}
		case i == 13:
			if !unicode.IsLetter(c) {
				result.Errors = append(result.Errors, NewValidationError("INE", "position 13 must be letter (state)", ErrInvalidINEFormat))
			}
		case i >= 14 && i <= 17:
			if !unicode.IsDigit(c) {
				result.Errors = append(result.Errors, NewValidationError("INE", "positions 14-17 must be digits", ErrInvalidINEFormat))
			}
		}
	}

	if len(result.Errors) == 0 {
		if err := v.validateINEDate(ineClave); err != nil {
			result.Errors = append(result.Errors, err)
		}
	}

	if err := v.ValidateCheckDigits(ineClave); err != nil {
		result.Errors = append(result.Errors, err)
	}

	if len(ineClave) >= 18 {
		ocr := ineClave[0:13]
		if v.ocrRegex.MatchString(ocr) {
			result.OCRNumber = ocr
		}
	}

	if len(ineClave) >= 18 {
		result.ElectionKey = ineClave
	}

	year, _ := strconv.Atoi(ineClave[6:8])
	month, _ := strconv.Atoi(ineClave[8:10])
	day, _ := strconv.Atoi(ineClave[10:12])

	var fullYear int
	currentYear := time.Now().Year() % 100
	if year <= currentYear {
		fullYear = 2000 + year
	} else {
		fullYear = 1900 + year
	}

	birthDate := time.Date(fullYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	result.BirthDate = &birthDate

	if len(ineClave) > 12 {
		result.Gender = v.parseGender(string(ineClave[12]))
	}

	result.ValidationScore = v.GetValidationScore(ineClave)
	result.IsValid = len(result.Errors) == 0

	return result, nil
}

func (v *ineValidator) ValidateOCR(ocr string) error {
	if len(ocr) != 13 {
		return NewValidationError("OCR", "must be 13 characters", ErrInvalidINEOCR)
	}

	for i, c := range ocr {
		if !unicode.IsDigit(c) {
			return NewValidationError("OCR", "position "+strconv.Itoa(i)+" must be digit", ErrInvalidINEOCR)
		}
	}

	if err := v.validateOCRChecksum(ocr); err != nil {
		return err
	}

	return nil
}

func (v *ineValidator) validateOCRChecksum(ocr string) error {
	if len(ocr) != 13 {
		return NewValidationError("OCR", "invalid length for checksum", ErrInvalidINEOCR)
	}

	sum := 0
	for i := 0; i < 12; i++ {
		digit, _ := strconv.Atoi(string(ocr[i]))
		weight := 13 - i
		sum += digit * weight
	}

	checksum := sum % 10
	if checksum != 0 {
		checksum = 10 - checksum
	}

	lastDigit, _ := strconv.Atoi(string(ocr[12]))
	if lastDigit != checksum {
		return NewValidationError("OCR", "checksum mismatch", ErrInvalidINEOCR)
	}

	return nil
}

func (v *ineValidator) ValidateElectionKey(key string) error {
	if len(key) < 18 || len(key) > 20 {
		return NewValidationError("ElectionKey", "must be 18-20 characters", ErrInvalidINEElectionKey)
	}

	for i := 0; i < 6; i++ {
		if !unicode.IsLetter(rune(key[i])) {
			return NewValidationError("ElectionKey", "positions 0-5 must be letters", ErrInvalidINEElectionKey)
		}
	}

	for i := 6; i < 10; i++ {
		if !unicode.IsDigit(rune(key[i])) {
			return NewValidationError("ElectionKey", "positions 6-9 must be digits", ErrInvalidINEElectionKey)
		}
	}

	return nil
}

func (v *ineValidator) ValidateCheckDigits(ineClave string) error {
	if len(ineClave) < 18 {
		return NewValidationError("INE", "insufficient length for check digits", ErrInvalidINEChecksum)
	}

	checkDigits := ineClave[len(ineClave)-2:]

	for _, c := range checkDigits {
		if !unicode.IsDigit(c) {
			return NewValidationError("INE", "check digits must be numeric", ErrInvalidINEChecksum)
		}
	}

	calculatedChecksum := v.calculateINEChecksum(ineClave[:len(ineClave)-2])
	expectedChecksum := calculatedChecksum[len(calculatedChecksum)-2:]

	if checkDigits != expectedChecksum {
		return NewValidationError("INE", "check digits mismatch", ErrInvalidINEChecksum)
	}

	return nil
}

func (v *ineValidator) calculateINEChecksum(ineClaveBase string) string {
	asciiValues := make([]int, len(ineClaveBase))
	for i, c := range ineClaveBase {
		switch {
		case c >= 'A' && c <= 'Z':
			asciiValues[i] = int(c - 'A' + 10)
		case c >= '0' && c <= '9':
			asciiValues[i] = int(c - '0')
		default:
			asciiValues[i] = 0
		}
	}

	sum := 0
	for i := 0; i < len(asciiValues); i++ {
		weight := len(asciiValues) - i
		sum += asciiValues[i] * weight
	}

	remainder := sum % 11
	checksum1 := remainder / 10
	checksum2 := remainder % 10

	return strconv.Itoa(checksum1) + strconv.Itoa(checksum2)
}

func (v *ineValidator) GetValidationScore(ineClave string) float64 {
	score := 0.0
	totalChecks := 4.0

	if len(ineClave) >= 18 && len(ineClave) <= 20 {
		score++
	}

	if v.claveRegex.MatchString(ineClave) || v.electionRegex.MatchString(ineClave) {
		score++
	}

	if err := v.validateINEDate(ineClave); err == nil {
		score++
	}

	if err := v.ValidateCheckDigits(ineClave); err == nil {
		score++
	}

	return (score / totalChecks) * 100
}

func (v *ineValidator) determineINEType(ineClave string) INEType {
	if len(ineClave) == 18 {
		return INETypeIFE
	} else if len(ineClave) >= 19 && len(ineClave) <= 20 {
		return INETypeINE
	}
	return INETypeUnknown
}

func (v *ineValidator) validateINEDate(ineClave string) error {
	if len(ineClave) < 12 {
		return NewValidationError("INE", "insufficient length for date", ErrInvalidINEFormat)
	}

	year, _ := strconv.Atoi(ineClave[6:8])
	month, _ := strconv.Atoi(ineClave[8:10])
	day, _ := strconv.Atoi(ineClave[10:12])

	if month < 1 || month > 12 {
		return NewValidationError("INE", "invalid month", ErrInvalidINEFormat)
	}

	daysInMonth := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if day < 1 || day > daysInMonth[month-1] {
		return NewValidationError("INE", "invalid day", ErrInvalidINEFormat)
	}

	var fullYear int
	currentYear := time.Now().Year() % 100
	if year <= currentYear {
		fullYear = 2000 + year
	} else {
		fullYear = 1900 + year
	}

	_ = time.Date(fullYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if month == 2 && day == 29 {
		isLeapYear := (fullYear%4 == 0 && fullYear%100 != 0) || (fullYear%400 == 0)
		if !isLeapYear {
			return NewValidationError("INE", "invalid leap year date", ErrInvalidINEFormat)
		}
	}

	return nil
}

func (v *ineValidator) parseGender(genderCode string) string {
	switch genderCode {
	case "H", "M":
		if genderCode == "H" {
			return "M"
		}
		return "F"
	default:
		return ""
	}
}