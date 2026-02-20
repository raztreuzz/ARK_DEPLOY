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

### Base URL

#### Production
- **Frontend**: `http://100.103.47.3:3000`
- **Backend API**: `http://127.0.0.1:5050` (localhost only)
- **Internal**: `http://ark-deploy:5050` (container network)

#### Development
- **Frontend**: `http://localhost:3000`
- **Backend API**: `http://localhost:5050`

---

### Health Check

#### GET /health

Returns server health status.

**Response:**
```json
{
  "status": "ok"
}
```

**Status Code:** 200 OK

---

### Products API

#### GET /products

List all registered products with their Jenkins job mappings.

**Response:**
```json
{
  "total": 2,
  "products": [
    {
      "id": "ark-game",
      "name": "ARK Game Server",
      "description": "Main game server instance",
      "jobs": {
        "prod": "deploy-ark-prod",
        "staging": "deploy-ark-staging"
      }
    }
  ]
}
```

**Status Code:** 200 OK

**Frontend Integration:** Fully implemented - displays in product catalog with search/filter functionality.

---

#### GET /products/:id

Retrieve details of a specific product by ID.

**Response:**
```json
{
  "id": "ark-game",
  "name": "ARK Game Server",
  "description": "Main game server instance",
  "jobs": {
    "prod": "deploy-ark-prod",
    "staging": "deploy-ark-staging"
  }
}
```

**Status Codes:**
- 200 OK - Product found
- 404 Not Found - Product does not exist

**Frontend Integration:** Not implemented. Backend ready.

---

#### POST /products

Create a new product with Jenkins job mappings.

**Request Body:**
```json
{
  "id": "my-product",
  "name": "My Product",
  "description": "Product description",
  "jobs": {
    "prod": "jenkins-job-prod",
    "dev": "jenkins-job-dev"
  }
}
```

**Response:**
```json
{
  "id": "my-product",
  "name": "My Product",
  "description": "Product description",
  "jobs": {
    "prod": "jenkins-job-prod",
    "dev": "jenkins-job-dev"
  }
}
```

**Status Codes:**
- 201 Created - Product created successfully
- 400 Bad Request - Invalid request body
- 409 Conflict - Product ID already exists

**Frontend Integration:** Not implemented. Requires modal form UI.

---

#### PUT /products/:id

Update an existing product.

**Request Body:**
```json
{
  "name": "Updated Product Name",
  "description": "Updated description",
  "jobs": {
    "prod": "new-jenkins-job"
  }
}
```

**Response:**
```json
{
  "id": "my-product",
  "name": "Updated Product Name",
  "description": "Updated description",
  "jobs": {
    "prod": "new-jenkins-job"
  }
}
```

**Status Codes:**
- 200 OK - Product updated
- 400 Bad Request - Invalid request body
- 404 Not Found - Product does not exist

**Frontend Integration:** Not implemented. Requires edit modal UI.

---

#### DELETE /products/:id

Delete a product from the system.

**Response:**
```json
{
  "message": "product deleted"
}
```

**Status Codes:**
- 200 OK - Product deleted
- 404 Not Found - Product does not exist

**Frontend Integration:** Not implemented. Requires delete button with confirmation.

---

### Deployments API

#### GET /jobs

List all available Jenkins jobs.

**Response:**
```json
{
  "jobs": [
    {
      "name": "deploy-ark-prod",
      "url": "http://jenkins.example.com/job/deploy-ark-prod/"
    },
    {
      "name": "build-backend",
      "url": "http://jenkins.example.com/job/build-backend/"
    }
  ]
}
```

**Status Code:** 200 OK

**Frontend Integration:** Not displayed in UI. Backend ready.

---

#### POST /deployments

Trigger a new deployment via Jenkins.

**Request Body (Product-based):**
```json
{
  "product_id": "ark-game",
  "environment": "prod",
  "parameters": {
    "VERSION": "1.2.3",
    "TARGET_HOST": "100.82.15.42"
  }
}
```

**Request Body (Direct Job):**
```json
{
  "job_name": "deploy-ark-prod",
  "parameters": {
    "VERSION": "1.2.3",
    "TARGET_HOST": "100.82.15.42"
  }
}
```

**Response:**
```json
{
  "job_name": "deploy-ark-prod",
  "queue_url": "http://jenkins.example.com/queue/item/123/",
  "message": "Deployment triggered successfully"
}
```

**Status Codes:**
- 201 Created - Deployment triggered
- 400 Bad Request - Invalid request or missing product/job
- 404 Not Found - Product not found
- 500 Internal Server Error - Jenkins connection failed

**Frontend Integration:** Fully implemented via deployment modal. Supports product selection, target host selection, and real-time status updates.

---

#### GET /deployments/pending

View pending jobs in Jenkins queue waiting for execution.

**Response:**
```json
{
  "pending_jobs": [
    {
      "id": 123,
      "task": {
        "name": "deploy-ark-prod"
      },
      "inQueueSince": 1708473600000
    }
  ]
}
```

**Status Code:** 200 OK

**Frontend Integration:** Not implemented. Backend ready.

---

#### GET /deployments/queue

Convert Jenkins queue URL to build number once job starts executing.

**Query Parameters:**
- `queue_url` (required) - Jenkins queue item URL

**Example:**
```
GET /deployments/queue?queue_url=http://jenkins.example.com/queue/item/123/
```

**Response:**
```json
{
  "build_number": 42
}
```

**Status Codes:**
- 200 OK - Build number retrieved
- 400 Bad Request - Missing queue_url parameter
- 404 Not Found - Queue item not found or not started yet

**Frontend Integration:** Not implemented. Backend ready.

---

#### GET /deployments/job/:job/build/:build/status

Get current status of a specific build.

**Example:**
```
GET /deployments/job/deploy-ark-prod/build/42/status
```

**Response:**
```json
{
  "building": false,
  "result": "SUCCESS",
  "duration": 45000,
  "timestamp": 1708473600000,
  "number": 42,
  "url": "http://jenkins.example.com/job/deploy-ark-prod/42/"
}
```

**Possible `result` values:**
- `SUCCESS` - Build completed successfully
- `FAILURE` - Build failed
- `UNSTABLE` - Build completed with test failures
- `ABORTED` - Build was manually stopped
- `null` - Build still in progress

**Status Code:** 200 OK

**Frontend Integration:** Partially implemented. Shows "provisioning" and "running" states. Does not poll for completion status.

---

#### GET /deployments/job/:job/build/:build/logs

Retrieve complete console output logs from a build.

**Example:**
```
GET /deployments/job/deploy-ark-prod/build/42/logs
```

**Response:**
Plain text with Jenkins console output:
```
Started by user admin
Building remotely on agent-01
[Pipeline] Start of Pipeline
[Pipeline] node
Running on agent-01
[Pipeline] {
...
[Pipeline] End of Pipeline
Finished: SUCCESS
```

**Status Code:** 200 OK

**Frontend Integration:** Not implemented. Backend ready for streaming. Requires SSE or polling implementation in frontend.

---

### Tailscale API

#### GET /tailscale/devices

List all devices connected to the Tailscale network.

**Response:**
```json
{
  "devices": [
    {
      "id": "123456",
      "name": "ark-prod-server",
      "hostname": "ark-prod-server",
      "addresses": ["100.82.15.42"],
      "online": true,
      "isExitNode": true,
      "os": "linux",
      "clientVersion": "1.58.2",
      "lastSeen": "2026-02-20T12:30:45Z"
    }
  ],
  "count": 3
}
```

**Status Code:** 200 OK

**Frontend Integration:** Fully implemented. Displays as interactive tree view organized by region with online/offline status indicators.

---

#### GET /tailscale/devices/:id

Get detailed information about a specific device.

**Response:**
```json
{
  "id": "123456",
  "name": "ark-prod-server",
  "hostname": "ark-prod-server",
  "addresses": ["100.82.15.42", "fd7a:115c:a1e0::1"],
  "online": true,
  "isExitNode": true,
  "os": "linux",
  "clientVersion": "1.58.2",
  "lastSeen": "2026-02-20T12:30:45Z",
  "created": "2026-01-15T08:20:00Z",
  "user": "admin@example.com"
}
```

**Status Codes:**
- 200 OK - Device found
- 500 Internal Server Error - Tailscale API error

**Frontend Integration:** Not implemented. Backend ready.

---

#### POST /tailscale/auth-keys

Generate an authentication key for registering new devices to the Tailscale network.

**Request Body:**
```json
{
  "description": "Key for new production server",
  "reusable": false,
  "ephemeral": false,
  "preauthorized": true,
  "expiry_seconds": 3600,
  "tags": ["tag:prod", "tag:server"]
}
```

**Field Descriptions:**
- `description` - Human-readable description (sanitized automatically)
- `reusable` - Whether key can be used multiple times
- `ephemeral` - Device is removed when disconnected
- `preauthorized` - Skip authorization approval step
- `expiry_seconds` - Key expiration time (default: 3600)
- `tags` - ACL tags to apply to device (optional)

**Response:**
```json
{
  "auth_key": "tskey-auth-kxxxxxxxxxxxxxxx-yyyyyyyyyyyyyyyyy",
  "id": "key-123",
  "created": "2026-02-20T12:30:45Z",
  "expires": "2026-02-20T13:30:45Z",
  "instructions": "Install Tailscale on the device and run: tailscale up --authkey=tskey-auth-kxxxxxxxxxxxxxxx-yyyyyyyyyyyyyyyyy"
}
```

**Status Codes:**
- 201 Created - Auth key generated successfully
- 400 Bad Request - Invalid request body
- 500 Internal Server Error - Tailscale API error

**Security Note:** Description field is automatically sanitized to remove special characters that could cause Tailscale API errors.

**Frontend Integration:** Fully implemented via "New Device" modal. Generates key and displays instructions. Auto-refreshes device list after 3 seconds.

---

#### DELETE /tailscale/devices/:id

Remove a device from the Tailscale network.

**Response:**
```json
{
  "message": "Device deleted successfully",
  "device_id": "123456"
}
```

**Status Codes:**
- 200 OK - Device removed
- 400 Bad Request - Missing device ID
- 500 Internal Server Error - Tailscale API error

**Frontend Integration:** Not implemented. Requires delete button with confirmation dialog.

---

### Frontend Integration Status

| Endpoint | HTTP Method | Frontend Status | Notes |
|----------|-------------|-----------------|-------|
| `/products` | GET | **Implemented** | Product catalog with search |
| `/products/:id` | GET | Not implemented | Backend ready |
| `/products` | POST | Not implemented | Requires create modal |
| `/products/:id` | PUT | Not implemented | Requires edit modal |
| `/products/:id` | DELETE | Not implemented | Requires delete button |
| `/deployments` | POST | **Implemented** | Deployment modal functional |
| `/deployments/job/.../status` | GET | Partial | Shows basic status only |
| `/deployments/job/.../logs` | GET | Not implemented | Needs streaming UI |
| `/deployments/pending` | GET | Not implemented | Backend ready |
| `/tailscale/devices` | GET | **Implemented** | Tree view with regions |
| `/tailscale/devices/:id` | GET | Not implemented | Backend ready |
| `/tailscale/auth-keys` | POST | **Implemented** | New device modal functional |
| `/tailscale/devices/:id` | DELETE | Not implemented | Requires delete button |

**Summary:** Core read and deployment operations are functional. CRUD operations for products and advanced monitoring features pending UI implementation.

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
