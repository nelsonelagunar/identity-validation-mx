# Mexico Identity Validation API

<div align="center">

**🇲🇽 API REST para Validación de Identidad Mexicana**

[![Status](https://img.shields.io/badge/Status-Demo-success.svg)](https://github.com/nelsonelagunar/mexico-identity-validation)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.11+-blue.svg)](https://python.org)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.109+-green.svg)](https://fastapi.tiangolo.com)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com)

[Demo](#demo) • [Features](#features) • [Instalación](#instalación) • [API Docs](#api-documentación)

</div>

---

## 📋 Descripción

**Mexico Identity Validation API** es una solución completa para validación de identidad electrónica en México, diseñada para empresas que necesitan:

- ✅ Validar INE/IFE, CURP, RFC
- ✅ Verificación biométrica facial
- ✅ Firma electrónica avanzada (estándar mexicano)
- ✅ Importación masiva de registros
- ✅ Soporte técnico directo

**⚠️ NOTA:** Este es un **proyecto demo** que simula las validaciones. Para uso en producción, necesitas contratar los servicios de:
- [SAT](https://www.sat.gob.mx) - RFC
- [RENAPO](https://www.gob.mx/segob/renapo) - CURP
- [INE](https://www.ine.mx) - INE/IFE
- Proveedores de biometría (Idemia, Onfido, etc.)

---

## 🚀 Features

| Feature | Descripción | Estado |
|---------|-------------|--------|
| **Validación CURP** | Verifica formato y existencia simulada | ✅ |
| **Validación RFC** | Valida estructura y calcula homoclave | ✅ |
| **Validación INE** | Simula validación de credencial electoral | ✅ |
| **Biometría Facial** | Comparación de fotos con score simulado | ✅ |
| **Firma Electrónica** | Genera firmas XAdES/PAdES | ✅ |
| **Importación Masiva** | Procesa CSV/Excel con colas | ✅ |
| **Webhooks** | Notificaciones de eventos | ✅ |
| **API Docs** | OpenAPI/Swagger completo | ✅ |
| **Rate Limiting** | Límites por API Key | ✅ |
| **Audit Logs** | Registro de todas las operaciones | ✅ |

---

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (FastAPI)                     │
│              Rate Limiting • Auth • Validation              │
└─────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│   Identity    │       │   Signature   │       │    Import     │
│   Service     │       │   Service     │       │    Service    │
│               │       │               │       │               │
│ • INE/IFE     │       │ • e.firma     │       │ • CSV/Excel   │
│ • CURP/RFC    │       │ • XAdES       │       │ • Async Queue │
│ • Biometric   │       │ • Timestamp   │       │ • Progress    │
└───────────────┘       └───────────────┘       └───────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────┐
│              Database (PostgreSQL + Redis)                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Instalación

### Requisitos

- Python 3.11+
- PostgreSQL 15+
- Redis 7+
- Docker (opcional)

### Con Docker

```bash
# Clonar repositorio
git clone https://github.com/nelsonelagunar/mexico-identity-validation.git
cd mexico-identity-validation

# Configurar variables de entorno
cp .env.example .env

# Levantar servicios
docker-compose up -d

# Ver logs
docker-compose logs -f api
```

### Sin Docker

```bash
# Crear entorno virtual
python -m venv venv
source venv/bin/activate

# Instalar dependencias
pip install -r requirements.txt

# Configurar base de datos
createdb mexico_identity
python scripts/init_db.py

# Ejecutar
uvicorn app.main:app --reload
```

---

## 🔧 Uso

### Validar CURP

```bash
curl -X POST http://localhost:8000/api/v1/identity/curp/validate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "curp": "LAGN850315HDFABC01",
    "nombres": "NELSON EVERALDO",
    "primer_apellido": "LAGUNA",
    "segundo_apellido": "RIVERA",
    "fecha_nacimiento": "1985-03-15"
  }'
```

**Respuesta:**

```json
{
  "valid": true,
  "curp": "LAGN850315HDFABC01",
  "nombres_match": true,
  "apellidos_match": true,
  "fecha_nacimiento_match": true,
  "entidad_registro": "DF",
  "sexo": "H",
  "score": 0.98,
  "audit_id": "audit_abc123"
}
```

### Validar RFC

```bash
curl -X POST http://localhost:8000/api/v1/identity/rfc/validate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "rfc": "LAGN850315ABC01",
    "nombre": "NELSON EVERALDO LAGUNA RIVERA"
  }'
```

### Validación Biométrica

```bash
curl -X POST http://localhost:8000/api/v1/biometric/compare \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "reference_image": "base64_encoded_image",
    "candidate_image": "base64_encoded_image",
    "curp": "LAGN850315HDFABC01"
  }'
```

**Respuesta:**

```json
{
  "match": true,
  "score": 0.92,
  "threshold": 0.85,
  "processing_time_ms": 245,
  "audit_id": "audit_def456"
}
```

### Firma Electrónica

```bash
curl -X POST http://localhost:8000/api/v1/signature/sign \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "document": "base64_encoded_pdf",
    "signer": {
      "curp": "LAGN850315HDFABC01",
      "rfc": "LAGN850315ABC01",
      "nombre": "NELSON EVERALDO LAGUNA RIVERA"
    },
    "signature_type": "xades"
  }'
```

### Importación Masiva

```bash
# Subir archivo CSV
curl -X POST http://localhost:8000/api/v1/import/bulk \
  -H "X-API-Key: your-api-key" \
  -F "file=@validaciones.csv" \
  -F "webhook_url=https://your-server.com/webhook"

# Respuesta inmediata
{
  "import_id": "imp_abc123",
  "status": "queued",
  "total_records": 10000,
  "estimated_time_seconds": 300,
  "webhook_url": "https://your-server.com/webhook"
}

# Consultar progreso
curl http://localhost:8000/api/v1/import/imp_abc123/status

{
  "import_id": "imp_abc123",
  "status": "processing",
  "processed": 5000,
  "total": 10000,
  "errors": 12,
  "progress_percent": 50
}
```

---

## 📊 API Documentation

- **Swagger UI:** http://localhost:8000/docs
- **ReDoc:** http://localhost:8000/redoc
- **OpenAPI Spec:** http://localhost:8000/openapi.json

---

## 🔐 Seguridad

| Feature | Descripción |
|---------|-------------|
| **API Keys** | Autenticación por API Key |
| **Rate Limiting** | 1000 req/min por API Key |
| **Audit Logs** | Todas las operaciones se registran |
| **Encryption** | TLS 1.3 para todas las conexiones |
| **Data Retention** | 90 días configurable |

---

## 🧪 Tests

```bash
# Ejecutar tests
pytest tests/ -v --cov=app

# Tests de integración
pytest tests/integration/ -v

# Tests de carga
locust -f tests/locustfile.py
```

---

## 📈 Monitoreo

- **Health Check:** `GET /health`
- **Metrics:** `GET /metrics` (Prometheus format)
- **Logs:** JSON structured logging

---

## 🚢 Deploy

### Azure AKS

```bash
# Crear imagen
docker build -t mexico-identity-api:latest .

# Subir a ACR
az acr login --name yourregistry
docker tag mexico-identity-api:latest yourregistry.azurecr.io/mexico-identity-api:latest
docker push yourregistry.azurecr.io/mexico-identity-api:latest

# Desplegar en AKS
kubectl apply -f k8s/
```

### Docker Compose

```bash
docker-compose up -d
```

---

## 📞 Soporte Técnico

Este es un proyecto demo. Para soporte técnico en proyectos reales:

- **Email:** contacto@nelsonlaguna.dev
- **LinkedIn:** [Nelson Laguna](https://linkedin.com/in/nelsonelagunar)
- **GitHub:** Issues en este repositorio

---

## 🤝 Contribuciones

Ver [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📄 Licencia

MIT License - Ver [LICENSE](LICENSE)

---

## 🙏 Créditos

Desarrollado por **Nelson Laguna**

Especialista en:
- Azure DevOps Engineer
- Kubernetes & Microservices
- API Design & Integration
- Identity Verification Systems

---

**⚠️ DISCLAIMER:** Este proyecto es un DEMO con datos simulados. NO usar en producción sin integrar APIs gubernamentales reales.