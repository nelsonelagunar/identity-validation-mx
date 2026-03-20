"""Tests for Identity endpoints."""

import pytest
from fastapi.testclient import TestClient
from datetime import date

from app.main import app


client = TestClient(app)


class TestHealthEndpoints:
    """Tests for health endpoints."""
    
    def test_root(self):
        """Test root endpoint."""
        response = client.get("/")
        assert response.status_code == 200
        assert "name" in response.json()
    
    def test_health(self):
        """Test health endpoint."""
        response = client.get("/health")
        assert response.status_code == 200
        assert response.json()["status"] == "healthy"


class TestCURPValidation:
    """Tests for CURP validation."""
    
    def test_valid_curp(self):
        """Test valid CURP."""
        response = client.post(
            "/api/v1/identity/curp/validate",
            headers={"X-API-Key": "test-key"},
            json={
                "curp": "LAGN850315HDFABC01",
                "nombres": "NELSON EVERALDO",
                "primer_apellido": "LAGUNA",
                "segundo_apellido": "RIVERA",
                "fecha_nacimiento": "1985-03-15"
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert "valid" in data
        assert "curp" in data
        assert "score" in data
    
    def test_invalid_curp_format(self):
        """Test invalid CURP format."""
        response = client.post(
            "/api/v1/identity/curp/validate",
            headers={"X-API-Key": "test-key"},
            json={
                "curp": "INVALID",
                "nombres": "TEST",
                "primer_apellido": "TEST",
                "fecha_nacimiento": "1990-01-01"
            }
        )
        assert response.status_code == 422  # Validation error
    
    def test_curp_missing_fields(self):
        """Test CURP validation with missing fields."""
        response = client.post(
            "/api/v1/identity/curp/validate",
            headers={"X-API-Key": "test-key"},
            json={
                "curp": "LAGN850315HDFABC01"
            }
        )
        assert response.status_code == 422


class TestRFCValidation:
    """Tests for RFC validation."""
    
    def test_valid_rfc_persona_fisica(self):
        """Test valid RFC for persona física."""
        response = client.post(
            "/api/v1/identity/rfc/validate",
            headers={"X-API-Key": "test-key"},
            json={
                "rfc": "LAGN850315ABC01",
                "nombre": "NELSON EVERALDO LAGUNA RIVERA",
                "rfc_tipo": "fisica"
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert "valid" in data
        assert "rfc" in data


class TestINEValidation:
    """Tests for INE validation."""
    
    def test_valid_ine(self):
        """Test valid INE."""
        response = client.post(
            "/api/v1/identity/ine/validate",
            headers={"X-API-Key": "test-key"},
            json={
                "clave_elector": "123456789012345678",
                "numero_emision": "01",
                "ocr": "1234567890123"
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert "valid" in data
        assert "clave_elector" in data


class TestBiometricComparison:
    """Tests for biometric comparison."""
    
    def test_face_comparison(self):
        """Test face comparison."""
        import base64
        # Dummy base64 image
        dummy_image = base64.b64encode(b"fake_image_data").decode()
        
        response = client.post(
            "/api/v1/biometric/compare",
            headers={"X-API-Key": "test-key"},
            json={
                "reference_image": dummy_image,
                "candidate_image": dummy_image,
                "curp": "LAGN850315HDFABC01",
                "threshold": 0.85
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert "match" in data
        assert "score" in data


class TestSignature:
    """Tests for digital signature."""
    
    def test_create_signature(self):
        """Test creating a signature."""
        import base64
        dummy_doc = base64.b64encode(b"document_content").decode()
        
        response = client.post(
            "/api/v1/signature/sign",
            headers={"X-API-Key": "test-key"},
            json={
                "document": dummy_doc,
                "signer": {
                    "curp": "LAGN850315HDFABC01",
                    "rfc": "LAGN850315ABC01",
                    "nombre": "NELSON EVERALDO LAGUNA RIVERA"
                },
                "signature_type": "xades"
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert "signature_id" in data
        assert "status" in data
    
    def test_verify_signature(self):
        """Test verifying a signature."""
        import base64
        dummy_doc = base64.b64encode(b"signed_document").decode()
        
        response = client.post(
            "/api/v1/signature/verify",
            headers={"X-API-Key": "test-key"},
            json={
                "document": dummy_doc
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert "valid" in data


class TestBulkImport:
    """Tests for bulk import."""
    
    def test_create_import(self):
        """Test creating a bulk import."""
        csv_content = "curp,nombres,primer_apellido\nLAGN850315HDFABC01,NELSON,LAGUNA\n"
        
        response = client.post(
            "/api/v1/import/bulk",
            headers={"X-API-Key": "test-key"},
            files={"file": ("test.csv", csv_content, "text/csv")},
            data={"webhook_url": "https://example.com/webhook"}
        )
        assert response.status_code == 200
        data = response.json()
        assert "import_id" in data
        assert "status" in data
    
    def test_get_import_status(self):
        """Test getting import status."""
        # First create an import
        csv_content = "curp,nombres\nTEST123456HDFABC01,TEST\n"
        
        create_response = client.post(
            "/api/v1/import/bulk",
            headers={"X-API-Key": "test-key"},
            files={"file": ("test.csv", csv_content, "text/csv")}
        )
        import_id = create_response.json()["import_id"]
        
        # Then check status
        response = client.get(
            f"/api/v1/import/{import_id}/status",
            headers={"X-API-Key": "test-key"}
        )
        assert response.status_code == 200
        data = response.json()
        assert "import_id" in data
        assert "status" in data