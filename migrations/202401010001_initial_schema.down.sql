-- Rollback initial schema for identity-validation-mx

DROP TABLE IF EXISTS bulk_import_stats CASCADE;
DROP TABLE IF EXISTS import_status_tracking CASCADE;
DROP TABLE IF EXISTS bulk_import_jobs CASCADE;

DROP TABLE IF EXISTS audit_trail CASCADE;

DROP TABLE IF EXISTS signature_verification_responses CASCADE;
DROP TABLE IF EXISTS signature_verification_requests CASCADE;
DROP TABLE IF EXISTS digital_signature_responses CASCADE;
DROP TABLE IF EXISTS digital_signature_requests CASCADE;

DROP TABLE IF EXISTS liveness_detection_responses CASCADE;
DROP TABLE IF EXISTS liveness_detection_requests CASCADE;
DROP TABLE IF EXISTS facial_comparison_responses CASCADE;
DROP TABLE IF EXISTS facial_comparison_requests CASCADE;

DROP TABLE IF EXISTS ine_validation_responses CASCADE;
DROP TABLE IF EXISTS ine_validation_requests CASCADE;
DROP TABLE IF EXISTS rfc_validation_responses CASCADE;
DROP TABLE IF EXISTS rfc_validation_requests CASCADE;
DROP TABLE IF EXISTS curp_validation_responses CASCADE;
DROP TABLE IF EXISTS curp_validation_requests CASCADE;

DROP TABLE IF EXISTS schema_migrations CASCADE;