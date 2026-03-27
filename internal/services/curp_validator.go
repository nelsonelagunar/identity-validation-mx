package services

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type CURPValidator interface {
	Validate(curp string) (*CURPValidationResult, error)
	ValidateFormat(curp string) error
	ValidateDate(curp string) error
	ValidateState(stateCode string) error
	CalculateCheckDigit(curp string) (string, error)
	GetValidationScore(curp string) float64
}

type CURPValidationResult struct {
	IsValid          bool
	CURP             string
	FullName         string
	BirthDate       *time.Time
	Gender          string
	BirthState      string
	ValidationScore float64
	CheckDigit      string
	CalculatedDigit string
	Errors          []error
}

type curpValidator struct {
	curpRegex      *regexp.Regexp
	stateCodes     map[string]string
	forbiddenWords map[string]bool
}

func NewCURPValidator() CURPValidator {
	curpRegex := regexp.MustCompile(`^([A-Z]{4})(\d{2})(\d{2})(\d{2})([HM])([A-Z]{2})([A-Z]{3})([A-Z\d]{1})(\d{1})$`)
	return &curpValidator{
		curpRegex: curpRegex,
		stateCodes: map[string]string{
			"AS": "Aguascalientes", "BC": "Baja California", "BS": "Baja California Sur",
			"CC": "Campeche", "CL": "Coahuila", "CM": "Colima", "CS": "Chiapas",
			"CH": "Chihuahua", "DF": "Ciudad de México", "DG": "Durango", "GT": "Guanajuato",
			"GR": "Guerrero", "HG": "Hidalgo", "JA": "Jalisco", "MC": "México",
			"MN": "Michoacán", "MS": "Morelos", "NT": "Nayarit", "NL": "Nuevo León",
			"OC": "Oaxaca", "PL": "Puebla", "QO": "Querétaro", "QR": "Quintana Roo",
			"SP": "San Luis Potosí", "SL": "Sinaloa", "SR": "Sonora", "TC": "Tabasco",
			"TS": "Tamaulipas", "TL": "Tlaxcala", "VZ": "Veracruz", "YN": "Yucatán",
			"ZS": "Zacatecas", "NE": "Nacido en el Extranjero",
		},
		forbiddenWords: map[string]bool{
			"BACA": true, "BAKA": true, "BUEI": true, "BUEY": true, "CACA": true,
			"CAKA": true, "COGE": true, "COJA": true, "COJE": true, "COLI": true,
			"CULO": true, "FETO": true, "GETA": true, "JOTO": true, "KACA": true,
			"KACO": true, "KAGA": true, "KAGO": true, "KAKA": true, "KULO": true,
			"MAME": true, "MAMO": true, "MEAR": true, "MEAS": true, "MEON": true,
			"MION": true, "MOCO": true, "MULA": true, "PEDA": true, "PEDO": true,
			"PENE": true, "PITO": true, "PUTA": true, "PUTO": true, "QULO": true,
			"RATA": true, "RUIN": true,
		},
	}
}

func (v *curpValidator) Validate(curp string) (*CURPValidationResult, error) {
	if curp == "" {
		return nil, ErrEmptyInput
	}

	curp = strings.ToUpper(strings.TrimSpace(curp))
	result := &CURPValidationResult{
		CURP:    curp,
		IsValid: false,
		Errors:  make([]error, 0),
	}

	if err := v.ValidateFormat(curp); err != nil {
		result.Errors = append(result.Errors, err)
		return result, nil
	}

	if err := v.ValidateDate(curp); err != nil {
		result.Errors = append(result.Errors, err)
	}

	if err := v.ValidateState(curp[11:13]); err != nil {
		result.Errors = append(result.Errors, err)
	}

	calculatedDigit, err := v.CalculateCheckDigit(curp)
	if err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		result.CalculatedDigit = calculatedDigit
		result.CheckDigit = string(curp[17])
		if result.CheckDigit != calculatedDigit {
			result.Errors = append(result.Errors, ErrInvalidCURPChecksum)
		}
	}

	result.Gender = string(curp[10])
	if birthState, ok := v.stateCodes[curp[11:13]]; ok {
		result.BirthState = birthState
	}

	year, _ := strconv.Atoi(curp[4:6])
	month, _ := strconv.Atoi(curp[6:8])
	day, _ := strconv.Atoi(curp[8:10])
	
	var fullYear int
	if year >= 0 && year <= 99 {
		if year >= 0 && year <= 25 {
			fullYear = 2000 + year
		} else {
			fullYear = 1900 + year
		}
	}
	
	birthDate := time.Date(fullYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	result.BirthDate = &birthDate

	result.ValidationScore = v.GetValidationScore(curp)
	result.IsValid = len(result.Errors) == 0

	return result, nil
}

func (v *curpValidator) ValidateFormat(curp string) error {
	if len(curp) != 18 {
		return NewValidationError("CURP", "must be exactly 18 characters", ErrInvalidCURPFormat)
	}

	for i, c := range curp {
		switch {
		case i < 4:
			if !unicode.IsUpper(c) {
				return NewValidationError("CURP", "position "+fmt.Sprint(i)+" must be uppercase letter", ErrInvalidCURPFormat)
			}
		case i >= 4 && i < 10:
			if !unicode.IsDigit(c) {
				return NewValidationError("CURP", "position "+fmt.Sprint(i)+" must be digit", ErrInvalidCURPFormat)
			}
		case i == 10:
			if c != 'H' && c != 'M' {
				return NewValidationError("CURP", "position 10 must be H or M", ErrInvalidCURPFormat)
			}
		case i >= 11 && i < 13:
			if !unicode.IsUpper(c) {
				return NewValidationError("CURP", "position "+fmt.Sprint(i)+" must be uppercase letter", ErrInvalidCURPFormat)
			}
		case i >= 13 && i < 16:
			if !unicode.IsUpper(c) {
				return NewValidationError("CURP", "position "+fmt.Sprint(i)+" must be uppercase letter", ErrInvalidCURPFormat)
			}
		case i == 16:
			if !unicode.IsUpper(c) && !unicode.IsDigit(c) {
				return NewValidationError("CURP", "position 16 must be uppercase letter or digit", ErrInvalidCURPFormat)
			}
		case i == 17:
			if !unicode.IsDigit(c) {
				return NewValidationError("CURP", "position 17 must be digit", ErrInvalidCURPFormat)
			}
		}
	}

	namePart := curp[0:4]
	if v.forbiddenWords[namePart] {
		return NewValidationError("CURP", "contains forbidden word", ErrInvalidCURPFormat)
	}

	return nil
}

func (v *curpValidator) ValidateDate(curp string) error {
	year, _ := strconv.Atoi(curp[4:6])
	month, _ := strconv.Atoi(curp[6:8])
	day, _ := strconv.Atoi(curp[8:10])

	if month < 1 || month > 12 {
		return NewValidationError("CURP", "invalid month", ErrInvalidCURPDate)
	}

	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	
	var fullYear int
	if year >= 0 && year <= 25 {
		fullYear = 2000 + year
	} else if year >= 26 && year <= 99 {
		fullYear = 1900 + year
	} else {
		return NewValidationError("CURP", "invalid year", ErrInvalidCURPDate)
	}

	isLeapYear := (fullYear%4 == 0 && fullYear%100 != 0) || (fullYear%400 == 0)
	if isLeapYear && month == 2 {
		if day < 1 || day > 29 {
			return NewValidationError("CURP", "invalid day for month", ErrInvalidCURPDate)
		}
	} else {
		if day < 1 || day > daysInMonth[month-1] {
			return NewValidationError("CURP", "invalid day for month", ErrInvalidCURPDate)
		}
	}

	return nil
}

func (v *curpValidator) ValidateState(stateCode string) error {
	if _, exists := v.stateCodes[stateCode]; !exists {
		return NewValidationError("CURP", "invalid state code: "+stateCode, ErrInvalidCURPState)
	}
	return nil
}

func (v *curpValidator) CalculateCheckDigit(curp string) (string, error) {
	if len(curp) != 18 {
		return "", ErrInvalidCURPFormat
	}

	curp17 := curp[:17]
	asciiValues := make([]int, len(curp17))
	
	for i, c := range curp17 {
		if unicode.IsDigit(c) {
			asciiValues[i] = int(c) - 48
		} else {
			asciiValues[i] = int(c) - 55
		}
	}

	sum := 0
	for i := 0; i < len(asciiValues); i++ {
		weight := 18 - i
		sum += asciiValues[i] * weight
	}

	remainder := sum % 10
	checkDigit := (10 - remainder) % 10

	return strconv.Itoa(checkDigit), nil
}

func (v *curpValidator) GetValidationScore(curp string) float64 {
	score := 0.0
	totalChecks := 5.0

	if len(curp) == 18 {
		score++
	}

	if v.curpRegex.MatchString(curp) {
		score++
	}

	if err := v.ValidateDate(curp); err == nil {
		score++
	}

	if err := v.ValidateState(curp[11:13]); err == nil {
		score++
	}

	if digit, err := v.CalculateCheckDigit(curp); err == nil && digit == string(curp[17]) {
		score++
	}

	return (score / totalChecks) * 100
}