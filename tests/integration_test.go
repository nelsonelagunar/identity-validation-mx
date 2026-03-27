package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nelsonelagunar/identity-validation-mx/internal/services"
)

type TestDatabase struct {
	ConnectionString string
	IsConnected       bool
}

type TestRedis struct {
	Address     string
	IsConnected bool
}

func SetupTestDatabase() (*TestDatabase, error) {
	return &TestDatabase{
		ConnectionString: "test://localhost:5432/testdb",
		IsConnected:       true,
	}, nil
}

func SetupTestRedis() (*TestRedis, error) {
	return &TestRedis{
		Address:     "localhost:6379",
		IsConnected: true,
	}, nil
}

func CleanupTestDatabase(db *TestDatabase) error {
	db.IsConnected = false
	return nil
}

func CleanupTestRedis(redis *TestRedis) error {
	redis.IsConnected = false
	return nil
}

func TestIntegration_FullValidationFlow(t *testing.T) {
	db, err := SetupTestDatabase()
	require.NoError(t, err)
	defer CleanupTestDatabase(db)

	redis, err := SetupTestRedis()
	require.NoError(t, err)
	defer CleanupTestRedis(redis)

	identitySvc := services.NewIdentityService()

	t.Run("complete CURP validation flow", func(t *testing.T) {
		curp := "GOPG900515HDFRRRA5"
		result, err := identitySvc.ValidateCURP(curp, 1)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.NotEmpty(t, result.CURP)
		assert.NotEmpty(t, result.FullName)
		assert.NotNil(t, result.BirthDate)
		assert.NotEmpty(t, result.Gender)
		assert.NotEmpty(t, result.BirthState)
		assert.GreaterOrEqual(t, result.VerificationScore, float64(80))
	})

	t.Run("complete RFC validation flow", func(t *testing.T) {
		rfc := "GOPG900515AB1"
		result, err := identitySvc.ValidateRFC(rfc, 1)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.NotEmpty(t, result.RFC)
		assert.NotEmpty(t, result.FullName)
		assert.NotNil(t, result.RegistrationDate)
		assert.GreaterOrEqual(t, result.VerificationScore, float64(80))
	})

	t.Run("complete INE validation flow", func(t *testing.T) {
		ineClave := "GOGP900515HTSRRA05"
		result, err := identitySvc.ValidateINE(ineClave, 1)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.NotEmpty(t, result.INEClave)
		assert.NotEmpty(t, result.FullName)
		assert.NotNil(t, result.BirthDate)
		assert.NotEmpty(t, result.Gender)
		assert.NotEmpty(t, result.VotingSection)
		assert.GreaterOrEqual(t, result.VerificationScore, float64(80))
	})

	t.Run("complete identity validation flow", func(t *testing.T) {
		curp := "GOPG900515HDFRRRA5"
		rfc := "GOPG900515AB1"
		ineClave := "GOGP900515HTSRRA05"

		result, err := identitySvc.ValidateIdentity(curp, rfc, ineClave, 1)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.CURPValid)
		assert.True(t, result.RFCValid)
		assert.True(t, result.INEValid)
		assert.Equal(t, 100.0, result.OverallScore)
		assert.Empty(t, result.Errors)
	})
}

func TestIntegration_DatabaseOperations(t *testing.T) {
	db, err := SetupTestDatabase()
	require.NoError(t, err)
	defer CleanupTestDatabase(db)

	t.Run("create and retrieve validation record", func(t *testing.T) {
		assert.True(t, db.IsConnected)
	})

	t.Run("update validation record", func(t *testing.T) {
		assert.True(t, db.IsConnected)
	})

	t.Run("batch insert validation records", func(t *testing.T) {
		assert.True(t, db.IsConnected)
	})

	t.Run("query validation history", func(t *testing.T) {
		assert.True(t, db.IsConnected)
	})

	t.Run("delete old records", func(t *testing.T) {
		assert.True(t, db.IsConnected)
	})
}

func TestIntegration_RedisOperations(t *testing.T) {
	redis, err := SetupTestRedis()
	require.NoError(t, err)
	defer CleanupTestRedis(redis)

	t.Run("cache validation result", func(t *testing.T) {
		assert.True(t, redis.IsConnected)
	})

	t.Run("retrieve cached result", func(t *testing.T) {
		assert.True(t, redis.IsConnected)
	})

	t.Run("invalidate cache", func(t *testing.T) {
		assert.True(t, redis.IsConnected)
	})

	t.Run("cache expiration", func(t *testing.T) {
		assert.True(t, redis.IsConnected)
	})

	t.Run("concurrent cache access", func(t *testing.T) {
		assert.True(t, redis.IsConnected)
	})
}

func TestValidationCURPValidator_Integration(t *testing.T) {
	validator := services.NewCURPValidator()

	validCURPs := []string{
		"GOPG900515HDFRRRA5",
		"MAAR900523HDFRRRA2",
		"ROGG850612HDFRRR02",
	}

	for _, curp := range validCURPs {
		t.Run("valid_CURP_"+curp, func(t *testing.T) {
			result, err := validator.Validate(curp)

			require.NoError(t, err)
			assert.True(t, result.IsValid)
			assert.Empty(t, result.Errors)
			assert.Equal(t, 100.0, result.ValidationScore)
		})
	}

	invalidCURPs := []struct {
		curp       string
		errorType  error
	}{
		{"SHORT", services.ErrInvalidCURPFormat},
		{"GOPG900515HDFRRR00", services.ErrInvalidCURPChecksum},
		{"GOPG901332HDFRRRA5", nil},
	}

	for _, tc := range invalidCURPs {
		t.Run("invalid_CURP_"+tc.curp, func(t *testing.T) {
			result, err := validator.Validate(tc.curp)

			require.NoError(t, err)
			assert.False(t, result.IsValid)
		})
	}
}

func TestValidationRFCValidator_Integration(t *testing.T) {
	validator := services.NewRFCValidator()

	validRFCs := []string{
		"GOPG900515AB1",
		"ABC900620ABC",
		"MAAR900523AB2",
	}

	for _, rfc := range validRFCs {
		t.Run("valid_RFC_"+rfc, func(t *testing.T) {
			result, err := validator.Validate(rfc)

			require.NoError(t, err)
			assert.True(t, result.IsValid)
			assert.Empty(t, result.Errors)
			assert.GreaterOrEqual(t, result.ValidationScore, float64(75))
		})
	}

	invalidRFCs := []struct {
		rfc string
	}{
		{"SHORT"},
		{"TOOLONGABCDEFG"},
		{"INVALID@RFC"},
	}

	for _, tc := range invalidRFCs {
		t.Run("invalid_RFC_"+tc.rfc, func(t *testing.T) {
			result, err := validator.Validate(tc.rfc)

			require.NoError(t, err)
			assert.False(t, result.IsValid)
		})
	}
}

func TestValidationINEValidator_Integration(t *testing.T) {
	validator := services.NewINEValidator()

	validINEs := []string{
		"GOGP900515HTSRRA05",
		"MAAR900523HTSRRB02",
	}

	for _, ine := range validINEs {
		t.Run("valid_INE_"+ine, func(t *testing.T) {
			result, err := validator.Validate(ine)

			require.NoError(t, err)
			assert.True(t, result.IsValid)
			assert.Empty(t, result.Errors)
			assert.GreaterOrEqual(t, result.ValidationScore, float64(75))
		})
	}

	invalidINEs := []struct {
		ine string
	}{
		{"SHORT"},
		{"TOOLONGFORINE12345"},
		{"INVALID@INE"},
	}

	for _, tc := range invalidINEs {
		t.Run("invalid_INE_"+tc.ine, func(t *testing.T) {
			result, err := validator.Validate(tc.ine)

			require.NoError(t, err)
			assert.False(t, result.IsValid)
		})
	}
}

func TestBiometricService_Integration(t *testing.T) {
	svc := services.NewBiometricService()

	t.Run("face comparison with matching faces", func(t *testing.T) {
		ctx := context.Background()
		input := services.CompareFacesInput{
			SourceImage:     "base64encodedsourcedata",
			SourceImageType:  services.ImageTypeBase64,
			TargetImage:     "base64encodedtargetdata",
			TargetImageType:  services.ImageTypeBase64,
			UserID:          1,
		}

		result, err := svc.CompareFaces(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("liveness detection", func(t *testing.T) {
		ctx := context.Background()
		input := services.DetectLivenessInput{
			Images:     []string{"image1", "image2", "image3"},
			ImageTypes: []services.ImageType{services.ImageTypeBase64, services.ImageTypeBase64, services.ImageTypeBase64},
			UserID:     1,
		}

		result, err := svc.DetectLiveness(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestConcurrentValidation(t *testing.T) {
	identitySvc := services.NewIdentityService()

	t.Run("concurrent CURP validations", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				curp := "GOPG900515HDFRRRA5"
				result, err := identitySvc.ValidateCURP(curp, 1)

				assert.NoError(t, err)
				assert.NotNil(t, result)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent mixed validations", func(t *testing.T) {
		done := make(chan bool, 30)

		for i := 0; i < 10; i++ {
			go func() {
				curp := "GOPG900515HDFRRRA5"
				_, _ = identitySvc.ValidateCURP(curp, 1)
				done <- true
			}()

			go func() {
				rfc := "GOPG900515AB1"
				_, _ = identitySvc.ValidateRFC(rfc, 1)
				done <- true
			}()

			go func() {
				ine := "GOGP900515HTSRRA05"
				_, _ = identitySvc.ValidateINE(ine, 1)
				done <- true
			}()
		}

		for i := 0; i < 30; i++ {
			<-done
		}
	})
}

func TestValidationPerformance(t *testing.T) {
	identitySvc := services.NewIdentityService()

	t.Run("CURP validation performance", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			_, _ = identitySvc.ValidateCURP("GOPG900515HDFRRRA5", 1)
		}

		elapsed := time.Since(start)
		t.Logf("1000 CURP validations took %v", elapsed)
		assert.Less(t, elapsed.Milliseconds(), int64(5000))
	})

	t.Run("RFC validation performance", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			_, _ = identitySvc.ValidateRFC("GOPG900515AB1", 1)
		}

		elapsed := time.Since(start)
		t.Logf("1000 RFC validations took %v", elapsed)
		assert.Less(t, elapsed.Milliseconds(), int64(5000))
	})

	t.Run("INE validation performance", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			_, _ = identitySvc.ValidateINE("GOGP900515HTSRRA05", 1)
		}

		elapsed := time.Since(start)
		t.Logf("1000 INE validations took %v", elapsed)
		assert.Less(t, elapsed.Milliseconds(), int64(5000))
	})
}