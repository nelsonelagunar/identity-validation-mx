#!/bin/bash
# Mexico Identity Validation API - Setup Script

set -e

echo "🇲🇽 Mexico Identity Validation API - Setup"
echo "============================================"

# Check Python version
PYTHON_VERSION=$(python3 --version 2>&1 | awk '{print $2}')
echo "✓ Python version: $PYTHON_VERSION"

# Create virtual environment
if [ ! -d "venv" ]; then
    echo "📦 Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "🔄 Activating virtual environment..."
source venv/bin/activate

# Install dependencies
echo "📥 Installing dependencies..."
pip install --upgrade pip
pip install -r requirements.txt

# Create .env if not exists
if [ ! -f ".env" ]; then
    echo "📋 Creating .env file..."
    cp .env.example .env
    echo "⚠️  Edit .env with your configuration"
fi

# Create directories
mkdir -p logs certs

# Initialize database (if PostgreSQL is running)
echo ""
echo "📊 Database setup:"
echo "  PostgreSQL: docker-compose up -d db"
echo "  Or set DATABASE_URL in .env"
echo ""

# Run tests
echo "🧪 Running tests..."
pytest tests/ -v --tb=short || true

# Done
echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env with your configuration"
echo "  2. Start services: docker-compose up -d"
echo "  3. Run: uvicorn app.main:app --reload"
echo "  4. Open: http://localhost:8000/docs"
echo ""
echo "📚 Documentation: docs/"
echo "🐛 Issues: https://github.com/nelsonelagunar/mexico-identity-validation/issues"