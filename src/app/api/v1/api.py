"""
API v1 Router
"""

from fastapi import APIRouter

from app.api.v1.endpoints import identity, biometric, signature, import_bulk, health

api_router = APIRouter()

# Include endpoints
api_router.include_router(health.router, prefix="/health", tags=["Health"])
api_router.include_router(identity.router, prefix="/identity", tags=["Identity"])
api_router.include_router(biometric.router, prefix="/biometric", tags=["Biometric"])
api_router.include_router(signature.router, prefix="/signature", tags=["Signature"])
api_router.include_router(import_bulk.router, prefix="/import", tags=["Import"])