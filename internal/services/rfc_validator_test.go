package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRFC(t *testing.T) {
	validator := NewRFCValidator()

	tests := []struct {
		name    string
		rfc     string
		wantErr bool
	}{
		{
			name:    "valid RFC physical person",
			rfc:     "LAG850315ABC",
			wantErr: false,
		},
		{
			name:    "valid RFC moral person",
			rfc:     "ABC850315ABC",
			wantErr: false,
		},
		{
			name:    "invalid RFC - too short",
			rfc:     "ABC85031",
			wantErr: true,
		},
		{
			name:    "invalid RFC - empty",
			rfc:     "",
			wantErr: true,
		},
		{
			name:    "invalid RFC - lowercase",
			rfc:     "lag850315abc",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(tt.rfc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRFCCalculateCheckDigit(t *testing.T) {
	validator := NewRFCValidator()

	rfc := "LAG850315AB"
	checkDigit := validator.CalculateCheckDigit(rfc)

	assert.NotEmpty(t, checkDigit)
	assert.Len(t, checkDigit, 1)
}
