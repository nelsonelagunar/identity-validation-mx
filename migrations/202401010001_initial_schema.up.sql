-- Initial schema for identity-validation-mx
-- CURP Validation Tables
CREATE TABLE IF NOT EXISTS curp_validation_requests (
    id SERIAL PRIMARY KEY,
    curp VARCHAR(18) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_curp_requests_user_id ON curp_validation_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_curp_requests_curp ON curp_validation_requests(curp);
CREATE INDEX IF NOT EXISTS idx_curp_requests_deleted_at ON curp_validation_requests(deleted_at);

CREATE TABLE IF NOT EXISTS curp_validation_responses (
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
    deleted_at TIMESTAMP,
    CONSTRAINT fk_curp_response_request FOREIGN KEY (request_id) REFERENCES curp_validation_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_curp_responses_deleted_at ON curp_validation_responses(deleted_at);

-- RFC Validation Tables
CREATE TABLE IF NOT EXISTS rfc_validation_requests (
    id SERIAL PRIMARY KEY,
    rfc VARCHAR(13) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rfc_requests_user_id ON rfc_validation_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_rfc_requests_rfc ON rfc_validation_requests(rfc);
CREATE INDEX IF NOT EXISTS idx_rfc_requests_deleted_at ON rfc_validation_requests(deleted_at);

CREATE TABLE IF NOT EXISTS rfc_validation_responses (
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
    deleted_at TIMESTAMP,
    CONSTRAINT fk_rfc_response_request FOREIGN KEY (request_id) REFERENCES rfc_validation_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_rfc_responses_deleted_at ON rfc_validation_responses(deleted_at);

-- INE Validation Tables
CREATE TABLE IF NOT EXISTS ine_validation_requests (
    id SERIAL PRIMARY KEY,
    ine_clave VARCHAR(18) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    ocr_number VARCHAR(13),
    election_key VARCHAR(18),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ine_requests_user_id ON ine_validation_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_ine_requests_clave ON ine_validation_requests(ine_clave);
CREATE INDEX IF NOT EXISTS idx_ine_requests_deleted_at ON ine_validation_requests(deleted_at);

CREATE TABLE IF NOT EXISTS ine_validation_responses (
    id SERIAL PRIMARY KEY,
    request_id INTEGER UNIQUE NOT NULL,
    is_valid BOOLEAN DEFAULT FALSE,
    full_name VARCHAR(200),
    birth_date TIMESTAMP,
    gender VARCHAR(1),
    address VARCHAR(500),
    voting_section VARCHAR(10),
    validation_error VARCHAR(500),
    ine_verified BOOLEAN DEFAULT FALSE,
    verification_score DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_ine_response_request FOREIGN KEY (request_id) REFERENCES ine_validation_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ine_responses_deleted_at ON ine_validation_responses(deleted_at);

-- Biometric Tables
CREATE TABLE IF NOT EXISTS facial_comparison_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    document_photo TEXT,
    selfie_photo TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_facial_requests_user_id ON facial_comparison_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_facial_requests_deleted_at ON facial_comparison_requests(deleted_at);

CREATE TABLE IF NOT EXISTS facial_comparison_responses (
    id SERIAL PRIMARY KEY,
    request_id INTEGER UNIQUE NOT NULL,
    is_match BOOLEAN DEFAULT FALSE,
    similarity_score DECIMAL(5,2),
    confidence_level DECIMAL(5,2),
    detected_anomalies TEXT,
    processing_time_ms BIGINT,
    provider_result TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_facial_response_request FOREIGN KEY (request_id) REFERENCES facial_comparison_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_facial_responses_deleted_at ON facial_comparison_responses(deleted_at);

CREATE TABLE IF NOT EXISTS liveness_detection_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    video_file VARCHAR(500),
    image_files TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_liveness_requests_user_id ON liveness_detection_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_liveness_requests_deleted_at ON liveness_detection_requests(deleted_at);

CREATE TABLE IF NOT EXISTS liveness_detection_responses (
    id SERIAL PRIMARY KEY,
    request_id INTEGER UNIQUE NOT NULL,
    is_live BOOLEAN DEFAULT FALSE,
    liveness_score DECIMAL(5,2),
    confidence_level DECIMAL(5,2),
    spoof_probability DECIMAL(5,2),
    detected_attacks TEXT,
    processing_time_ms BIGINT,
    provider_result TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_liveness_response_request FOREIGN KEY (request_id) REFERENCES liveness_detection_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_liveness_responses_deleted_at ON liveness_detection_responses(deleted_at);

-- Digital Signature Tables
CREATE TABLE IF NOT EXISTS digital_signature_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    document_hash VARCHAR(128) NOT NULL,
    signer_name VARCHAR(200) NOT NULL,
    signer_rfc_curp VARCHAR(18),
    signature_type VARCHAR(20) DEFAULT 'basic',
    expires_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_signature_requests_user_id ON digital_signature_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_signature_requests_hash ON digital_signature_requests(document_hash);
CREATE INDEX IF NOT EXISTS idx_signature_requests_deleted_at ON digital_signature_requests(deleted_at);

CREATE TABLE IF NOT EXISTS digital_signature_responses (
    id SERIAL PRIMARY KEY,
    request_id INTEGER UNIQUE NOT NULL,
    signature TEXT,
    serial_number VARCHAR(64),
    issuer_dn VARCHAR(500),
    subject_dn VARCHAR(500),
    valid_from TIMESTAMP,
    valid_to TIMESTAMP,
    signature_base64 TEXT,
    certificate TEXT,
    provider_result TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_signature_response_request FOREIGN KEY (request_id) REFERENCES digital_signature_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_signature_responses_deleted_at ON digital_signature_responses(deleted_at);

CREATE TABLE IF NOT EXISTS signature_verification_requests (
    id SERIAL PRIMARY KEY,
    signature_id INTEGER NOT NULL,
    document_hash VARCHAR(128) NOT NULL,
    signature TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_verification_requests_signature_id ON signature_verification_requests(signature_id);
CREATE INDEX IF NOT EXISTS idx_verification_requests_deleted_at ON signature_verification_requests(deleted_at);

CREATE TABLE IF NOT EXISTS signature_verification_responses (
    id SERIAL PRIMARY KEY,
    request_id INTEGER UNIQUE NOT NULL,
    is_valid BOOLEAN DEFAULT FALSE,
    signer_verified BOOLEAN DEFAULT FALSE,
    document_integrity BOOLEAN DEFAULT FALSE,
    timestamp_valid BOOLEAN DEFAULT FALSE,
    error_code VARCHAR(50),
    error_message TEXT,
    verification_details TEXT,
    processing_time_ms BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_verification_response_request FOREIGN KEY (request_id) REFERENCES signature_verification_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_verification_responses_deleted_at ON signature_verification_responses(deleted_at);

-- Audit Trail Table
CREATE TABLE IF NOT EXISTS audit_trail (
    audit_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    action VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    request TEXT,
    response TEXT,
    status VARCHAR(20) NOT NULL,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    module VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit_trail(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_trail(action);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_trail(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_module ON audit_trail(module);
CREATE INDEX IF NOT EXISTS idx_audit_deleted_at ON audit_trail(deleted_at);

-- Bulk Import Tables
CREATE TABLE IF NOT EXISTS bulk_import_jobs (
    id SERIAL PRIMARY KEY,
    job_id VARCHAR(64) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    file_hash VARCHAR(128),
    total_records INTEGER DEFAULT 0,
    processed_records INTEGER DEFAULT 0,
    success_records INTEGER DEFAULT 0,
    failed_records INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_import_jobs_user_id ON bulk_import_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_import_jobs_job_id ON bulk_import_jobs(job_id);
CREATE INDEX IF NOT EXISTS idx_import_jobs_deleted_at ON bulk_import_jobs(deleted_at);

CREATE TABLE IF NOT EXISTS import_status_tracking (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL,
    record_number INTEGER NOT NULL,
    record_data TEXT,
    validation_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    processing_time_ms BIGINT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_tracking_job FOREIGN KEY (job_id) REFERENCES bulk_import_jobs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tracking_job_id ON import_status_tracking(job_id);
CREATE INDEX IF NOT EXISTS idx_tracking_status ON import_status_tracking(status);
CREATE INDEX IF NOT EXISTS idx_tracking_deleted_at ON import_status_tracking(deleted_at);

CREATE TABLE IF NOT EXISTS bulk_import_stats (
    job_id INTEGER PRIMARY KEY,
    total_time_ms BIGINT DEFAULT 0,
    average_time_ms BIGINT DEFAULT 0,
    min_time_ms BIGINT DEFAULT 0,
    max_time_ms BIGINT DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 0.00,
    failure_rate DECIMAL(5,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_stats_job FOREIGN KEY (job_id) REFERENCES bulk_import_jobs(id) ON DELETE CASCADE
);

-- Schema migrations tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);