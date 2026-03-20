"""
Mexico Identity Validation API
Main application entry point
"""

from contextlib import asynccontextmanager
from datetime import datetime
from typing import Optional

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from app.core.config import settings
from app.core.database import database
from app.api.v1.api import api_router
from app.core.logging import setup_logging

import logging

# Setup logging
setup_logging()
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    logger.info("🇲🇽 Starting Mexico Identity Validation API...")
    
    # Initialize database
    await database.connect()
    logger.info("✅ Database connected")
    
    # Initialize Redis
    # await redis.connect()
    # logger.info("✅ Redis connected")
    
    yield
    
    # Cleanup
    logger.info("🛑 Shutting down Mexico Identity Validation API...")
    await database.disconnect()
    # await redis.disconnect()


# Create FastAPI app
app = FastAPI(
    title="Mexico Identity Validation API",
    description="""
    API REST para validación de identidad mexicana.
    
    ## Features
    
    * **Validación INE/IFE** - Credencial electoral
    * **Validación CURP** - Clave Única de Registro de Población
    * **Validación RFC** - Registro Federal de Contribuyentes
    * **Biometría Facial** - Comparación de fotografías
    * **Firma Electrónica** - XAdES/PAdES
    * **Importación Masiva** - CSV/Excel con webhooks
    
    ## Autenticación
    
    Todas las peticiones requieren un API Key en el header:
    `X-API-Key: your-api-key`
    
    ## Rate Limiting
    
    * 1000 peticiones/minuto por API Key
    * 10000 peticiones/día por API Key
    
    ## Nota
    
    ⚠️ Este es un proyecto DEMO con datos simulados.
    Para uso en producción, integra APIs gubernamentales reales.
    """,
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
    lifespan=lifespan,
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API router
app.include_router(api_router, prefix="/api/v1")


# Health check endpoint
@app.get("/health", tags=["Health"])
async def health_check():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "service": "mexico-identity-validation",
        "version": "1.0.0",
        "timestamp": datetime.utcnow().isoformat(),
    }


# Metrics endpoint (Prometheus format)
@app.get("/metrics", tags=["Monitoring"])
async def metrics():
    """Prometheus metrics endpoint."""
    return {
        "requests_total": 1000,
        "requests_successful": 985,
        "requests_failed": 15,
        "avg_response_time_ms": 145,
    }


# Root endpoint
@app.get("/", tags=["Root"])
async def root():
    """Root endpoint with API info."""
    return {
        "name": "Mexico Identity Validation API",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health",
        "country": "🇲🇽",
    }


# Exception handlers
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """Global exception handler."""
    logger.error(f"Unhandled exception: {exc}", exc_info=True)
    return JSONResponse(
        status_code=500,
        content={
            "error": "internal_server_error",
            "message": "An unexpected error occurred",
            "request_id": request.headers.get("X-Request-ID", "unknown"),
        },
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
    )