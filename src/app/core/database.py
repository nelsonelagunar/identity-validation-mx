"""
Database configuration
"""

import databases
import sqlalchemy
from sqlalchemy.ext.declarative import declarative_base

from app.core.config import settings

# Database instance
database = databases.Database(settings.DATABASE_URL)

# SQLAlchemy metadata
metadata = sqlalchemy.MetaData()

# Base for models
Base = declarative_base()


async def get_database():
    """Get database connection."""
    return database