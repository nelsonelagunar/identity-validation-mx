package repository

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationConfig struct {
	MigrationsPath string
	DatabaseURL    string
}

type MigrationRunner struct {
	config MigrationConfig
	db     *sql.DB
}

func NewMigrationRunner(config MigrationConfig, db *Database) (*MigrationRunner, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return &MigrationRunner{
		config: config,
		db:     sqlDB,
	}, nil
}

func (m *MigrationRunner) RunMigrations() error {
	log.Println("Running migrations...")

	rows, err := m.db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'schema_migrations'
	`)
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if _, err := m.db.Exec(`
			CREATE TABLE schema_migrations (
				version VARCHAR(255) PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`); err != nil {
			return fmt.Errorf("failed to create migrations table: %w", err)
		}
	}

	return nil
}

func (m *MigrationRunner) GetCurrentVersion() (string, error) {
	var version string
	err := m.db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		return "", nil
	}
	return version, nil
}

func (m *MigrationRunner) RecordMigration(version string) error {
	_, err := m.db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	return err
}

func CreateMigrationTables(db *Database) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS curp_validation_requests (
			id SERIAL PRIMARY KEY,
			curp VARCHAR(18) UNIQUE NOT NULL,
			user_id INTEGER NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_curp_requests_user_id ON curp_validation_requests(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_curp_requests_deleted_at ON curp_validation_requests(deleted_at)`,

		`CREATE TABLE IF NOT EXISTS curp_validation_responses (
			id SERIAL PRIMARY KEY,
			request_id INTEGER UNIQUE NOT NULL,
			is_valid BOOLEAN DEFAULT FALSE,
			full_name VARCHAR(200),
			birth_date TIMESTAMP,
			gender VARCHAR(1),
			birth_state VARCHAR(50),
			validation_error VARCHAR(500),
			renapo_verified BOOLEAN DEFAULT FALSE,
			verification_score DECIMAL(5,2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_curp_responses_deleted_at ON curp_validation_responses(deleted_at)`,

		`CREATE TABLE IF NOT EXISTS rfc_validation_requests (
			id SERIAL PRIMARY KEY,
			rfc VARCHAR(13) UNIQUE NOT NULL,
			user_id INTEGER NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rfc_requests_user_id ON rfc_validation_requests(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rfc_requests_deleted_at ON rfc_validation_requests(deleted_at)`,

		`CREATE TABLE IF NOT EXISTS rfc_validation_responses (
			id SERIAL PRIMARY KEY,
			request_id INTEGER UNIQUE NOT NULL,
			is_valid BOOLEAN DEFAULT FALSE,
			full_name VARCHAR(200),
			tax_regime VARCHAR(50),
			registration_date TIMESTAMP,
			status_sat VARCHAR(50),
			validation_error VARCHAR(500),
			sat_verified BOOLEAN DEFAULT FALSE,
			verification_score DECIMAL(5,2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rfc_responses_deleted_at ON rfc_validation_responses(deleted_at)`,
	}

	for _, q := range queries {
		if _, err := sqlDB.Exec(q); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	log.Println("Migration tables created successfully")
	return nil
}