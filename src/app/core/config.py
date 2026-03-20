"""
Application configuration
"""

from pydantic_settings import BaseSettings
from typing import List


class Settings(BaseSettings):
    """Application settings."""
    
    # App
    APP_NAME: str = "Mexico Identity Validation API"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = False
    
    # Database
    DATABASE_URL: str = "postgresql://user:pass@localhost/mexico_identity"
    DATABASE_POOL_SIZE: int = 10
    
    # Redis
    REDIS_URL: str = "redis://localhost:6379/0"
    
    # Security
    API_KEY_HEADER: str = "X-API-Key"
    SECRET_KEY: str = "your-secret-key-change-in-production"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 60
    
    # Rate Limiting
    RATE_LIMIT_PER_MINUTE: int = 1000
    RATE_LIMIT_PER_DAY: int = 10000
    
    # CORS
    CORS_ORIGINS: List[str] = ["*"]
    
    # External APIs (Demo mode - simulated)
    RENAPO_API_URL: str = "https://demo.renapo.gob.mx/api"
    SAT_API_URL: str = "https://demo.sat.gob.mx/api"
    INE_API_URL: str = "https://demo.ine.mx/api"
    
    # Signature
    SIGNATURE_CERTIFICATE_PATH: str = "./certs/demo.p12"
    SIGNATURE_CERTIFICATE_PASSWORD: str = "demo"
    
    # Import
    MAX_IMPORT_FILE_SIZE_MB: int = 50
    IMPORT_BATCH_SIZE: int = 1000
    IMPORT_WEBHOOK_TIMEOUT: int = 30
    
    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()