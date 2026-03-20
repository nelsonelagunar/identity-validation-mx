"""
Health check endpoints
"""

from fastapi import APIRouter
from datetime import datetime

router = APIRouter()


@router.get("/")
async def health():
    """Basic health check."""
    return {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
    }


@router.get("/ready")
async def ready():
    """Readiness probe for Kubernetes."""
    return {
        "status": "ready",
        "services": {
            "api": "ok",
            "database": "ok",
            "redis": "ok",
        }
    }


@router.get("/live")
async def live():
    """Liveness probe for Kubernetes."""
    return {"status": "alive"}