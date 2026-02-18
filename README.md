# ARK_DEPLOY

**ARK_DEPLOY** is a deployment management system that integrates Jenkins automation with Tailscale network infrastructure. Built with Go backend and React frontend, it provides a modern web interface and RESTful API for managing products, deployments, and network devices in a secure, containerized environment.

## Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Project Structure](#project-structure)
- [API Endpoints](#api-endpoints)
- [Configuration](#configuration)
- [Installation](#installation)
- [Running the Application](#running-the-application)
- [Testing](#testing)
- [Docker Deployment](#docker-deployment)
- [Production Deployment (Ansible + Jenkins)](#production-deployment-ansible--jenkins) 🆕
- [Development](#development)

---

## Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        WEB[Web Browser]
        CLI[CLI/Frontend]
        API_TESTS[Integration Tests]
    end

    subgraph "Frontend - React + Vite"
        REACT[React Application]
        NGINX[Nginx Server]
    end

    subgraph "ARK_DEPLOY Backend"
        ROUTER[Gin HTTP Router]
        
        subgraph "Handlers"
            HEALTH[Health Check]
            PRODUCTS[Products Handler]
            DEPLOYMENTS[Deployments Handler]
            TAILSCALE[Tailscale Handler]
        end
        
        subgraph "Services"
            JENKINS_CLIENT[Jenkins Client]
            TS_CLIENT[Tailscale Client]
            PRODUCT_STORE[Product Store]
        end
        
        subgraph "Business Logic"
            SANITIZER[Description Sanitizer]
            VALIDATOR[Input Validator]
        end
    end

    subgraph "External Services"
        JENKINS[Jenkins Server]
        TS_API[Tailscale API]
    end

    WEB --> NGINX
    NGINX --> REACT
    REACT --> ROUTER
    CLI --> ROUTER
    API_TESTS --> ROUTER
    
    ROUTER --> HEALTH
    ROUTER --> PRODUCTS
    ROUTER --> DEPLOYMENTS
    ROUTER --> TAILSCALE
    
    PRODUCTS --> PRODUCT_STORE
    DEPLOYMENTS --> JENKINS_CLIENT
    DEPLOYMENTS --> PRODUCT_STORE
    TAILSCALE --> TS_CLIENT
    TAILSCALE --> SANITIZER
    
    JENKINS_CLIENT --> JENKINS
    TS_CLIENT --> TS_API
```

### Technology Stack

**Backend:**
- **Language**: Go 1.26
- **Web Framework**: Gin
- **Container**: Docker with Alpine Linux
- **Testing**: testify/assert

**Frontend:**
- **Framework**: React 18
- **Build Tool**: Vite
- **UI Library**: Lucide React (icons)
- **Styling**: Tailwind CSS
- **Web Server**: Nginx (production)

**Infrastructure:**
- **CI/CD**: Jenkins
- **Network**: Tailscale VPN
- **Orchestration**: Docker Compose

---

## Features

### Web Interface (Frontend)
- Modern, responsive React dashboard
- Real-time deployment monitoring
- Product catalog management
- Tailscale nodes visualization
- Live log streaming
- Deployment modal workflow

### Product Management
- Create, read, update, and delete products
- Map products to environment-specific Jenkins jobs
- In-memory storage with concurrent access control

### Deployment Orchestration
- Trigger Jenkins jobs via REST API
- Support for product-based or direct job deployment
- Real-time deployment status monitoring
- Build log streaming
- Queue management and pending jobs tracking

### Tailscale Integration
- List connected devices in the Tailscale network
- Generate authentication keys for new devices
- Device management (view details, remove devices)
- Automatic description sanitization for API compliance

### Security
- Environment-based configuration
- No hardcoded credentials
- Input sanitization and validation
- Docker containerization

---

## Project Structure

```
ARK_DEPLOY/
├── cmd/
│   ├── api/
│   │   └── main.go              # Application entry point
│   └── test_api/
│       ├── tailscale_test.go    # Integration tests for Tailscale endpoints
│       └── README.md            # Test documentation
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration loader
│   ├── deployments/
│   │   ├── handler.go           # Deployment HTTP handlers
│   │   ├── handler_test.go      # Unit tests for handlers
│   │   └── monitor.go           # Build monitoring handlers
│   ├── jenkins/
│   │   ├── client.go            # Jenkins API client
│   │   └── reads.go             # Jenkins read operations
│   ├── products/
│   │   ├── handler.go           # Product CRUD handlers
│   │   └── handler_test.go      # Unit tests
│   ├── server/
│   │   └── routes.go            # Route registration
│   ├── storage/
│   │   └── product.go           # In-memory product store
│   └── tailscale/
│       ├── client.go            # Tailscale API client
│       ├── devices.go           # Device management operations
│       ├── handler.go           # Tailscale HTTP handlers
│       ├── handler_test.go      # Unit tests
│       ├── sanitize.go          # Description sanitizer
│       └── sanitize_test.go     # Sanitizer tests
├── frontend/
│   ├── src/
│   │   ├── App.jsx              # Main React component
│   │   ├── main.jsx             # React entry point
│   │   └── index.css            # Global styles
│   ├── Dockerfile               # Multi-stage frontend build
│   ├── nginx.conf               # Nginx configuration
│   ├── vite.config.js           # Vite build configuration
│   ├── package.json             # Node dependencies
│   ├── index.html               # HTML template
│   └── README.md                # Frontend documentation
├── Dockerfile                   # Backend Docker build
├── docker-compose.yml           # Multi-service orchestration
├── .env                         # Environment variables (not in VCS)
├── .dockerignore               # Docker build exclusions
├── go.mod                      # Go module dependencies
└── go.sum                      # Dependency checksums
```

---

## API Endpoints

### Health Check

```http
GET /health
```

Returns server status.

### Products API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/products` | Create a new product |
| GET | `/products` | List all products |
| GET | `/products/:id` | Get product by ID |
| PUT | `/products/:id` | Update product |
| DELETE | `/products/:id` | Delete product |

**Product Model:**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "jobs": {
    "environment": "jenkins_job_name"
  }
}
```

### Deployments API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/jobs` | List available Jenkins jobs |
| POST | `/deployments` | Trigger a new deployment |
| GET | `/deployments/pending` | View pending jobs in queue |
| GET | `/deployments/queue` | Convert queue URL to build number |
| GET | `/deployments/job/:job/build/:build/status` | Get build status |
| GET | `/deployments/job/:job/build/:build/logs` | Stream build logs |

**Deployment Request:**
```json
{
  "product_id": "string",
  "environment": "string",
  "parameters": {
    "key": "value"
  }
}
```

**Alternative (Direct Job):**
```json
{
  "job_name": "string",
  "parameters": {
    "key": "value"
  }
}
```

### Tailscale API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/tailscale/devices` | List all devices in network |
| GET | `/tailscale/devices/:id` | Get device details |
| POST | `/tailscale/auth-keys` | Create authentication key |
| DELETE | `/tailscale/devices/:id` | Remove device from network |

**Auth Key Request:**
```json
{
  "description": "string",
  "reusable": false,
  "ephemeral": false,
  "preauthorized": true,
  "expiry_seconds": 3600,
  "tags": ["tag:value"]
}
```

---

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Application Configuration
ARK_PORT=5050

# Jenkins Configuration
JENKINS_BASE_URL=http://jenkins.example.com:8080
JENKINS_USER=admin
JENKINS_API_TOKEN=your_jenkins_api_token
JENKINS_JOB=default_job_name

# Tailscale Configuration
TAILSCALE_API_KEY=tskey-api-xxxxxxxxxxxxx
TAILSCALE_TAILNET=example.com
```

### Obtaining Credentials

**Jenkins API Token:**
1. Login to Jenkins
2. Click on your username (top right)
3. Click "Configure"
4. Under "API Token", click "Add new Token"
5. Copy the generated token

**Tailscale API Key:**
1. Visit https://login.tailscale.com/admin/settings/keys
2. Click "Generate API key"
3. Copy the generated key (format: `tskey-api-...`)

**Tailscale Tailnet:**
- Personal account: Use your email address
- Organization: Use your organization domain
- View at: https://login.tailscale.com/admin/settings/general

---

## Installation

### Prerequisites

- Go 1.26 or higher
- Docker (optional, for containerized deployment)
- Access to Jenkins server
- Tailscale account with API access

### Clone Repository

```bash
git clone <repository-url>
cd ARK_DEPLOY
```

### Install Dependencies

```bash
go mod download
```

---

## Running the Application

### Local Development

```bash
# Ensure .env file is configured
go run cmd/api/main.go
```

The server will start on `http://localhost:5050` (or the port specified in `ARK_PORT`).

### Verify Server

```bash
curl http://localhost:5050/health
```

Expected response:
```json
{"status":"ok"}
```

---

## Testing

### Unit Tests

Run all unit tests with mocked dependencies:

```bash
go test ./... -v
```

Run tests for specific package:

```bash
# Tailscale tests
go test ./internal/tailscale/... -v

# Products tests
go test ./internal/products/... -v

# Deployments tests
go test ./internal/deployments/... -v
```

### Test Coverage

```bash
go test ./internal/tailscale/... -cover
go test ./internal/products/... -cover
go test ./internal/deployments/... -cover
```

Generate HTML coverage report:

```bash
go test ./internal/tailscale/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests require the server to be running.

**Terminal 1** - Start server:
```bash
go run cmd/api/main.go
```

**Terminal 2** - Run integration tests:
```bash
go test ./cmd/test_api/... -v
```

---

## Docker Deployment

### Prerequisites

- Docker Engine 20.10+
- Docker Compose V2

### Quick Start (Recommended)

Start both backend and frontend services:

```bash
# Build and start all services
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build
```

Access the application:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:5050

### Individual Services

#### Backend Only

```bash
docker-compose up ark-deploy
```

#### Frontend Only

```bash
docker-compose up ark-frontend
```

### Manual Docker Build

#### Backend

```bash
docker build -t ark-deploy-backend:latest .

docker run -d \
  --name ark-deploy-backend \
  -p 5050:5050 \
  --env-file .env \
  ark-deploy-backend:latest
```

#### Frontend

```bash
cd frontend
docker build -t ark-deploy-frontend:latest .

docker run -d \
  --name ark-deploy-frontend \
  -p 3000:3000 \
  ark-deploy-frontend:latest
```

### Managing Services

View logs:
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f ark-deploy
docker-compose logs -f ark-frontend
```

Stop services:
```bash
docker-compose down
```

Restart services:
```bash
docker-compose restart
```

### Docker Image Details

**Backend Image:**
- Base: golang:1.26-alpine (builder) + alpine:latest (runtime)
- Size: ~25MB (optimized multi-stage build)
- Exposes: Port 5050
- Health check: `/health` endpoint

**Frontend Image:**
- Base: node:20-alpine (builder) + nginx:alpine (runtime)
- Size: ~25MB (optimized static files + nginx)
- Exposes: Port 3000
- Health check: HTTP GET on `/`
- Serves: React SPA with client-side routing
- Proxy: `/api/*` routes forwarded to backend

---

## Development

### Backend Development

#### Code Organization

The project follows a standard Go project layout:

- **`cmd/`**: Application entry points
- **`internal/`**: Private application logic (not importable by external projects)
- **`internal/config/`**: Configuration management
- **`internal/server/`**: HTTP server setup and routing
- **`internal/{module}/`**: Domain-specific logic organized by feature

#### Adding New Endpoints

1. Create handler in appropriate `internal/{module}/handler.go`
2. Register route in `internal/server/routes.go`
3. Add unit tests in `internal/{module}/handler_test.go`
4. Add integration tests in `cmd/test_api/`

### Code Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Write tests for all new features
- Document exported functions and types

### Dependency Management

```bash
# Add new dependency
go get github.com/package/name

# Update dependencies
go get -u ./...

# Tidy up
go mod tidy
```

### Frontend Development

#### Local Development Setup

```bash
cd frontend

# Install dependencies
npm install

# Run development server with hot reload
npm run dev
```

The development server will start on `http://localhost:3000` with Vite's HMR (Hot Module Replacement).

#### Frontend Code Organization

```
frontend/
├── src/
│   ├── App.jsx          # Main application component
│   ├── main.jsx         # React entry point
│   └── index.css        # Global styles (Tailwind)
├── index.html           # HTML template
├── vite.config.js       # Vite configuration & proxy
└── nginx.conf           # Production nginx config
```

#### Building for Production

```bash
cd frontend
npm run build
```

Builds are optimized and placed in `frontend/dist/`. The Dockerfile handles this automatically.

#### Adding New Features

1. Edit `src/App.jsx` to add new UI components
2. API calls go through `/api/*` prefix (proxied to backend)
3. Styles use Tailwind CSS utility classes
4. Icons use `lucide-react` library

---

## API Flow Diagrams

### Deployment Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant ProductStore
    participant JenkinsClient
    participant Jenkins

    Client->>API: POST /deployments
    API->>ProductStore: Get Product by ID
    ProductStore-->>API: Product with Jobs
    API->>API: Resolve Job from Environment
    API->>JenkinsClient: TriggerBuild(job, params)
    JenkinsClient->>Jenkins: POST /job/{name}/buildWithParameters
    Jenkins-->>JenkinsClient: Queue URL
    JenkinsClient-->>API: Queue URL
    API-->>Client: 201 Created with Queue URL
```

### Tailscale Device Management

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant TailscaleClient
    participant TailscaleAPI

    Client->>API: POST /tailscale/auth-keys
    API->>API: Sanitize Description
    API->>TailscaleClient: CreateAuthKey(sanitized)
    TailscaleClient->>TailscaleAPI: POST /api/v2/tailnet/{tailnet}/keys
    TailscaleAPI-->>TailscaleClient: Auth Key Response
    TailscaleClient-->>API: Auth Key
    API-->>Client: 201 Created with Key + Instructions
```

---

## Troubleshooting

### Server fails to start

**Error**: `missing required env vars`

**Solution**: Ensure all required environment variables are set in `.env` file.

### Jenkins connection fails

**Error**: `error triggering build`

**Solution**: Verify Jenkins URL, credentials, and network connectivity.

### Tailscale API errors

**Error**: `API error (status 400): keys: description had invalid characters`

**Solution**: This should be handled automatically by the sanitizer. Ensure you're using the latest version.

### Tests fail with connection errors

**Error**: `Error al conectar con el servidor`

**Solution**: Ensure the server is running before executing integration tests.

---

## Production Deployment (Ansible + Jenkins)

### Overview

ARK_DEPLOY incluye un sistema completo de deployment automatizado usando:
- **Ansible** para orquestación de infraestructura
- **Jenkins** para CI/CD pipeline
- **Docker Compose** para gestión de contenedores en producción

### Quick Start

```bash
# 1. Instalar dependencias de Ansible
pip3 install ansible>=2.9
ansible-galaxy collection install -r ansible-requirements.yml

# 2. Configurar inventario de servidores
nano inventory/production.ini

# 3. Configurar variables de entorno de producción
cp .env.prod.example .env.prod
nano .env.prod

# 4. Deployment manual
ansible-playbook \
  -i inventory/production.ini \
  deploy-playbook.yml \
  --extra-vars "repo_dir=$(pwd)" \
  --extra-vars "env_file=$(pwd)/.env.prod" \
  --extra-vars "compose_file=$(pwd)/docker-compose.prod.yml"
```

### Archivos de Deployment

| Archivo | Propósito |
|---------|-----------|
| `deploy-playbook.yml` | Playbook principal de Ansible |
| `ansible-requirements.yml` | Collections de Ansible necesarias |
| `docker-compose.prod.yml` | Configuración Docker para producción |
| `Jenkinsfile` | Pipeline CI/CD |
| `inventory/production.ini` | Inventario de servidores |
| `.env.prod.example` | Plantilla de variables de entorno |

### Pipeline de Jenkins

El `Jenkinsfile` automatiza:

1. 🔍 **Checkout** - Clona el repositorio
2. 🔧 **Setup Ansible** - Instala dependencias
3. ✅ **Validación** - Verifica sintaxis y conectividad
4. 🧪 **Tests** - Ejecuta tests unitarios Go
5. 🏗️ **Build** - Construye imágenes Docker
6. 🚀 **Deploy** - Ejecuta playbook Ansible
7. 🔬 **Health Check** - Verifica servicios
8. 🧹 **Limpieza** - Limpia recursos temporales

### Features del Playbook

- ✅ Validación pre-deploy (sintaxis, conectividad)
- ✅ Sincronización automática de archivos con rsync
- ✅ Gestión de red Docker personalizada
- ✅ Health checks post-deployment
- ✅ Logging estructurado de cada fase
- ✅ Rollback automático en caso de error
- ✅ Limpieza de recursos no utilizados

### Documentación Completa

Para información detallada sobre deployment, troubleshooting y rollback, consulta:

📖 **[DEPLOYMENT.md](DEPLOYMENT.md)** - Guía completa de deployment

---

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]

## Contact

[Add contact information here]

---

**Version**: 1.0.0  
**Last Updated**: February 2026
