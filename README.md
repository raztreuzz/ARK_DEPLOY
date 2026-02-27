# ARK_DEPLOY

ARK_DEPLOY es una plataforma para desplegar productos en nodos cliente de forma automatizada, usando una API backend, Jenkins y Docker Compose.

## Problema que resuelve

- Evita despliegues manuales por SSH en múltiples clientes.
- Estandariza creación de instancias, trazabilidad y callback de resultado.
- Expone rutas web por instancia de forma consistente.

## Tecnologías

- Backend: Go (Gin), Redis
- Orquestación: Jenkins + Ansible
- Runtime: Docker / Docker Compose
- Red: Tailscale
- Frontend: React + Vite + Nginx

## Diagrama

![Diagrama de Eventos](docs/Diagramas/Diagrama%20de%20eventos.png)

## Cómo correrlo

### Opción 1: Docker Compose
```bash
docker-compose up --build
```

- Backend: `http://localhost:5050`
- Frontend: `http://localhost:3000`

### Opción 2: Backend local
```bash
cp .env.example .env
go run cmd/api/main.go
```

## Documentación técnica

- Backend general: [docs/backend/backend.md](docs/backend/backend.md)
- Arquitectura backend (ruteo, callback, seguridad): [docs/backend/arquitectura.md](docs/backend/arquitectura.md)
- Flujo de despliegue backend: [docs/backend/deployment-flow.md](docs/backend/deployment-flow.md)
- Frontend: [frontend/README.md](frontend/README.md)
