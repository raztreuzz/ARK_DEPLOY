# Backend Overview

## Objetivo

El backend de ARK_DEPLOY expone API REST para administrar productos, iniciar despliegues y consultar estado/logs de instancias.

## Modulos principales

- `internal/server`: registro de rutas HTTP.
- `internal/deployments`: creacion de despliegues y consulta de estado.
- `internal/products`: CRUD de productos y jobs asociados.
- `internal/instances`: gestion de rutas/registro de instancias.
- `internal/jenkins`: cliente HTTP para trigger y lectura de builds.
- `internal/tailscale`: integracion para descubrimiento de dispositivos.
- `internal/storage`: stores en memoria/redis para estado operativo.

## Entradas y salidas

- Entrada: requests HTTP (`/api/...`, `/instances/...`).
- Salida: JSON de estado, callbacks y proxy hacia instancias desplegadas.

## Dependencias externas

- Jenkins API
- Redis
- Tailscale API
