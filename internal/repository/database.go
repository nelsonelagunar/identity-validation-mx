package repository

import (
	"fmt"
	"log"
	"time"

	"identity-validation-mx/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode        string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime time.Duration
	LogLevel       logger.LogLevel
}

type Database struct {
	DB *gorm.DB
}

func NewDatabase(config DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}

	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}

	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	return &Database{DB: db}, nil
}

func (d *Database) AutoMigrate() error {
	err := d.DB.AutoMigrate(
		&models.CURPValidationRequest{},
		&models.CURPValidationResponse{},
		&models.RFCValidationRequest{},
		&models.RFCValidationResponse{},
		&models.INEValidationRequest{},
		&models.INEValidationResponse{},
		&models.FacialComparisonRequest{},
		&models.FacialComparisonResponse{},
		&models.LivenessDetectionRequest{},
		&models.LivenessDetectionResponse{},
		&models.DigitalSignatureRequest{},
		&models.DigitalSignatureResponse{},
		&models.SignatureVerificationRequest{},
		&models.SignatureVerificationResponse{},
		&models.AuditTrail{},
		&models.BulkImportJob{},
		&models.ImportStatusTracking{},
		&models.BulkImportStats{},
	)

	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

func (d *Database) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}