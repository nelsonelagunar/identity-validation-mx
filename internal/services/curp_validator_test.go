package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCURP(t *testing.T) {
	validator := NewCURPValidator()

	tests := []struct {
		name    string
		curp    string
		wantErr bool
	}{
		{
			name:    "valid CURP",
			curp:    "LAGN850315HDFABC01",
			wantErr: false,
		},
		{
			name:    "invalid CURP - too short",
			curp:    "LAGN850315",
			wantErr: true,
		},
		{
			name:    "invalid CURP - too long",
			curp:    "LAGN850315HDFABC010",
			wantErr: true,
		},
		{
			name:    "invalid CURP - empty",
			curp:    "",
			wantErr: true,
		},
		{
			name:    "invalid CURP - lowercase",
			curp:    "lagn850315hdfabc01",
			wantErr: false,
		},
		{
			name:    "invalid CURP - special characters",
			curp:    "LAGN850315HDFABC@01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(tt.curp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCURPFormat(t *testing.T) {
	validator := NewCURPValidator()

	curp := "LAGN850315HDFABC01"
	result, err := validator.Validate(curp)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, curp, result.CURP)
}

func TestCURPExtractData(t *testing.T) {
	validator := NewCURPValidator()

	curp := "LAGN850315HDFABC01"
	data, err := validator.ExtractData(curp)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "LA", data.Initials)
	assert.Equal(t, "GN", data.LastName)
}
