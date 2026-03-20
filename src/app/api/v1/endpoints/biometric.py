"""
Biometric validation endpoints - Facial comparison
"""

import time
import uuid
from typing import Optional

from fastapi import APIRouter, Header, HTTPException
from pydantic import BaseModel, Field

router = APIRouter()


class BiometricCompareRequest(BaseModel):
    """Request para comparación biométrica facial."""
    reference_image: str = Field(..., description="Imagen de referencia en base64")
    candidate_image: str = Field(..., description="Imagen candidata en base64")
    curp: str = Field(..., min_length=18, max_length=18, description="CURP del usuario")
    threshold: Optional[float] = Field(0.85, ge=0.0, le=1.0, description="Umbral de matching")
    liveness_check: Optional[bool] = Field(False, description="Verificar que la imagen es real")


class BiometricCompareResponse(BaseModel):
    """Response de comparación biométrica."""
    match: bool
    score: float = Field(..., ge=0.0, le=1.0)
    threshold: float
    liveness_passed: Optional[bool] = None
    processing_time_ms: int
    audit_id: str
    confidence: str


class BiometricLivenessRequest(BaseModel):
    """Request para verificar liveness."""
    image: str = Field(..., description="Imagen en base64")
    challenge_type: Optional[str] = Field("blink", description="Tipo de challenge")


class BiometricLivenessResponse(BaseModel):
    """Response de verificación de liveness."""
    liveness: bool
    confidence: float
    challenge_passed: bool
    audit_id: str


# Endpoints
@router.post("/compare", response_model=BiometricCompareResponse)
async def compare_faces(
    request: BiometricCompareRequest,
    x_api_key: str = Header(..., description="API Key"),
):
    """
    Comparar dos imágenes faciales.
    
    Este endpoint compara una imagen de referencia (INE, pasaporte, etc.)
    con una imagen candidata (selfie) para verificar identidad.
    
    ⚠️ **DEMO:** Esta comparación es simulada. Para producción, integra con
    proveedores de biometría como Idemia, Onfido, Jumio, etc.
    """
    start_time = time.time()
    audit_id = f"audit_{uuid.uuid4().hex[:12]}"
    
    # Simulated face comparison
    # In production, use a face recognition library like:
    # - face_recognition (Python)
    # - AWS Rekognition
    # - Azure Face API
    # - Idemia API
    # - Onfido API
    
    # Simulated score (0.0 to 1.0)
    import random
    score = round(random.uniform(0.75, 0.99), 2)
    
    # Determine match
    match = score >= request.threshold
    
    # Liveness check (simulated)
    liveness_passed = None
    if request.liveness_check:
        liveness_passed = score >= 0.80  # Simplified
    
    processing_time = int((time.time() - start_time) * 1000)
    
    # Determine confidence level
    if score >= 0.95:
        confidence = "very_high"
    elif score >= 0.90:
        confidence = "high"
    elif score >= 0.85:
        confidence = "medium"
    else:
        confidence = "low"
    
    return BiometricCompareResponse(
        match=match,
        score=score,
        threshold=request.threshold,
        liveness_passed=liveness_passed,
        processing_time_ms=processing_time,
        audit_id=audit_id,
        confidence=confidence,
    )


@router.post("/liveness", response_model=BiometricLivenessResponse)
async def check_liveness(
    request: BiometricLivenessRequest,
    x_api_key: str = Header(..., description="API Key"),
):
    """
    Verificar que la imagen es real (no una foto/video).
    
    ⚠️ **DEMO:** Verificación simulada.
    """
    audit_id = f"audit_{uuid.uuid4().hex[:12]}"
    
    # Simulated liveness check
    # In production, use:
    # - AWS Rekognition Liveness
    # - Azure Face Liveness
    # - Idemia Liveness
    # - Onfido Liveness
    
    confidence = round(0.85 + (hash(request.image) % 15) / 100, 2)
    
    return BiometricLivenessResponse(
        liveness=confidence >= 0.80,
        confidence=confidence,
        challenge_passed=confidence >= 0.80,
        audit_id=audit_id,
    )