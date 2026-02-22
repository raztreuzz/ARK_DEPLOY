# ARK_DEPLOY Backend

Sistema de gestión de despliegues que integra automatización Jenkins con infraestructura de red Tailscale. API REST para gestionar productos, despliegues y dispositivos de red.

## Inicio Rápido

### Docker
```bash
docker-compose up --build
```
- Backend: http://localhost:5050
- Frontend: http://localhost:3000

### Desarrollo Local
```bash
go run cmd/api/main.go
```

---

## Stack de Tecnologías

**Backend:**
- Go 1.26
- Framework Gin
- UUIDs: github.com/google/uuid (IDs de instancias)
- Docker con Alpine Linux

**Infraestructura:**
- Jenkins CI/CD
- VPN Tailscale
- Docker Compose

---

## Arquitectura de Componentes

![Diagrama de componentes](docs/Diagramas/Diagrama%20de%20componentes.png)

---

## Estructura del Proyecto

```
cmd/
  api/main.go              # Punto de entrada
  test_api/               # Tests de integración

internal/
  config/config.go        # Configuración
  deployments/            # Gestores de despliegue
  jenkins/                # Cliente Jenkins
  products/               # CRUD de productos
  server/routes.go        # Rutas
  storage/product.go      # Almacenamiento
  tailscale/              # Integración Tailscale
```

---

## Instalación

### Requisitos Previos

- Go 1.26+
- Docker 20.10+ / Docker Compose V2
- Servidor Jenkins con acceso API
- Cuenta Tailscale con clave API

### Configuración

```bash
git clone <repository-url>
cd ARK_DEPLOY
go mod download
cp .env.example .env
```

---

## Configuración

### Variables de Entorno (.env)

```bash
ARK_PORT=5050
ARK_PUBLIC_HOST=100.103.47.3:3000    # Public host for reverse proxy URLs (optional)
JENKINS_BASE_URL=http://jenkins.example.com:8080
JENKINS_USER=admin
JENKINS_API_TOKEN=tu_token
TAILSCALE_API_KEY=tskey-api-xxxxxxxxxxxxx
TAILSCALE_TAILNET=example.com
```

### Obtener Credenciales

- **Token Jenkins**: Usuario > Configurar > API Token
- **Clave Tailscale**: https://login.tailscale.com/admin/settings/keys
- **Tailnet Tailscale**: https://login.tailscale.com/admin/settings/general

---

## Endpoints API

### URLs Base

- **Dev**: http://localhost:5050
- **Prod**: http://127.0.0.1:5050 (localhost), http://ark-deploy:5050 (contenedor)

### Endpoints Principales

| Módulo | Endpoint | Método | Estado |
|--------|----------|--------|--------|
| Health | `/health` | GET | Implementado |
| Productos | `/products` | GET/POST/PUT/DELETE | Parcial |
| Despliegues | `/deployments` | POST | Implementado |
| Jenkins | `/jobs` | GET | Implementado |
| Tailscale | `/tailscale/devices` | GET | Implementado |
| Tailscale | `/tailscale/devices/:id` | DELETE | Parcial |

---

## Jenkins Jobs Configuration

**⚠️ IMPORTANT**: Product jobs must publish an ephemeral host port and register instance routing in ARK.

### Container Deployment Requirements

✅ **CORRECT**:
```bash
docker run -d --name <instance_id> -p 0:80 nginx:alpine
PORT=$(docker port <instance_id> 80/tcp | awk -F: '{print $2}' | tail -n 1)
curl -X POST http://<ARK_BACKEND>:5050/instances/register \
  -H "Content-Type: application/json" \
  -d "{\"instance_id\":\"<instance_id>\",\"target_host\":\"<tailscale_ip>\",\"target_port\":$PORT}"
```

❌ **INCORRECT**:
```bash
docker run -d --name <instance_id> --network ark_production nginx:alpine
```

### Architecture

- Nginx in ARK only forwards `/instances/...` to backend ARK (`:5050`)
- Backend ARK resolves `instance_id -> target_host:target_port` from Redis
- Upstream container keeps running on the client node, not on Vault
- Public access remains: `http://<ARK_PUBLIC_HOST>:3000/instances/<instance_id>/`

---

## Testing

```bash
# Tests unitarios
go test ./... -v
go test ./... -cover

# Tests de integración (servidor debe estar ejecutándose)
go test ./cmd/test_api/... -v

# Reporte de cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Docker

```bash
# Todos los servicios
docker-compose up --build

# Solo backend
docker-compose up ark-deploy

# Solo frontend
docker-compose up ark-frontend

# Detener
docker-compose down
```

---

## Desarrollo

### Backend

```bash
# Desarrollo
go run cmd/api/main.go

# Estándares
go fmt ./...
go test ./... -v
go mod tidy
```

### Organización del Código

- `cmd/` - Puntos de entrada
- `internal/config/` - Configuración
- `internal/server/` - Rutas y configuración
- `internal/{module}/handler.go` - Manejadores
- `internal/{module}/handler_test.go` - Tests

---

## Flujo de Eventos del Sistema

![Diagrama de eventos](docs/Diagramas/Diagrama%20de%20eventos.png)

---

## Solución de Problemas

| Problema | Solución |
|----------|----------|
| Variables de entorno faltantes | Configura todas las variables en `.env` |
| Conexión Jenkins falla | Verifica URL, credenciales, red |
| Error API Tailscale | Verifica clave API |
| Tests fallan | Asegúrate de que el servidor se está ejecutando |

---

## Recursos

- [Documentación Jenkins](https://www.jenkins.io/doc/)
- [Documentación Go](https://golang.org/doc/)
- [Documentación Docker](https://docs.docker.com/)
- [Documentación Tailscale](https://tailscale.com/docs/)
- [GitHub Container Registry](https://docs.github.com/es/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

---

**Versión**: 1.0.0 | **Última actualización**: Febrero 2026
