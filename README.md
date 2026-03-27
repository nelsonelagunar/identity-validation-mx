# Identity Validation MX

🇲🇽 **API REST para validación de identidad mexicana** - CURP, RFC, INE, Biometría, Firma Electrónica

## 📋 Descripción

API de grado enterprise para validación de identidad mexicana con soporte para CURP, RFC, INE/IFE, comparación biométrica facial y firmas digitales.

## 🛠️ Stack Tecnológico

- **Lenguaje:** Go 1.21+
- **Framework:** Fiber v2
- **Base de datos:** PostgreSQL (GORM)
- **Cache/Queue:** Redis
- **Containerización:** Docker, Kubernetes
- **Logging:** Zerolog

## 📦 Requisitos

- Go 1.21 o superior
- PostgreSQL 14+
- Redis 7+
- Docker y Docker Compose (opcional)
- Make (opcional)

## 🚀 Inicio Rápido

### Sin Docker

```bash
# Clonar repositorio
git clone https://github.com/nelsonelagunar/identity-validation-mx.git
cd identity-validation-mx

# Descargar dependencias
go mod download
go mod tidy

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus configuraciones

# Ejecutar
make run
# o directamente:
go run ./cmd/server
```

### Con Docker

```bash
# Construir y ejecutar
make docker-build
make docker-up

# Ver logs
docker-compose -f deployments/docker/docker-compose.yml logs -f
```

## 📡 Endpoints API

### Health Check
```
GET  /health
GET  /ready
```

### Validación de Identidad
```
POST /api/v1/identity/curp/validate
POST /api/v1/identity/rfc/validate
POST /api/v1/identity/ine/validate
```

### Biometría
```
POST /api/v1/biometric/compare
POST /api/v1/biometric/liveness
```

### Firma Digital
```
POST /api/v1/signature/sign
POST /api/v1/signature/verify
```

### Import Bulk
```
POST /api/v1/import/bulk
GET  /api/v1/import/:id/status
```

### Ejemplo de Request CURP

```bash
curl -X POST http://localhost:8080/api/v1/identity/curp/validate \
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

### Ejemplo de Response

```json
{
  "valid": true,
  "curp": "LAGN850315HDFABC01",
  "score": 0.98,
  "nombres_match": true,
  "apellidos_match": true,
  "audit_id": "audit_abc123"
}
```

## 🔐 Autenticación

Todas las requests deben incluir el header `X-API-Key` con una API key válida.

```bash
X-API-Key: your-api-key
```

## ⚙️ Configuración

Ver `.env.example` para todas las variables de configuración disponibles:

| Variable | Descripción | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Puerto del servidor | `8080` |
| `DB_HOST` | Host de PostgreSQL | `localhost` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `REDIS_HOST` | Host de Redis | `localhost` |
| `API_RATE_LIMIT_REQUESTS` | Límite de requests/min | `100` |

## 🧪 Tests

```bash
# Ejecutar todos los tests
make test

# Tests con cobertura
make test-coverage
```

## 🐳 Docker & Kubernetes

### Docker Compose

```bash
# Iniciar servicios
docker-compose -f deployments/docker/docker-compose.yml up -d

# Detener servicios
docker-compose -f deployments/docker/docker-compose.yml down
```

### Kubernetes

```bash
# Desplegar
kubectl apply -f deployments/kubernetes/namespace.yaml
kubectl apply -f deployments/kubernetes/configmap.yaml
kubectl apply -f deployments/kubernetes/secrets.yaml
kubectl apply -f deployments/kubernetes/deployment.yaml
kubectl apply -f deployments/kubernetes/service.yaml
kubectl apply -f deployments/kubernetes/ingress.yaml
```

## 📁 Estructura del Proyecto

```
identity-validation-mx/
├── cmd/
│   └── server/           # Entry point
├── internal/
│   ├── api/              # Handlers, routes, DTOs
│   ├── config/           # Configuración
│   ├── middleware/       # Auth, logging, CORS
│   ├── models/           # Data models
│   ├── repository/       # Database access
│   └── services/         # Business logic
├── pkg/
│   └── utils/            # Utilidades compartidas
├── deployments/
│   ├── docker/           # Docker files
│   └── kubernetes/       # K8s manifests
├── migrations/           # SQL migrations
├── tests/               # Integration tests
└── go.mod
```

## 📊 Monitoreo

Métricas disponibles en:
```
GET /metrics
```

## 🔄 Rate Limiting

Rate limiting configurado por defecto:
- **Tier Free:** 100 requests/minuto
- **Tier Pro:** 1000 requests/minuto
- **Tier Enterprise:** Ilimitado

## 📝 Licencia

Proprietary Software - Todos los derechos reservados.

Para consultas de licenciamiento, contactar al autor.

## 👤 Contacto

**Nelson Laguna**
- LinkedIn: [linkedin.com/in/nelsonelagunar](https://linkedin.com/in/nelsonelagunar)
- Email: nlaguna@mykeepper.com

## ⚠️ Disclaimer

Esta es una demostración de capacidades de API. Para implementación en producción:

1. Integración con APIs gubernamentales oficiales (RENAPO, SAT, INE)
2. Cumplimiento con LFPDPPP
3. Proveedores biométricos certificados
4. Auditorías de seguridad apropiadas
