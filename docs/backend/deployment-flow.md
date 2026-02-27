# Deployment Flow

## Flujo resumido

1. Cliente solicita despliegue a `POST /api/deployments`. ([aqui](../../internal/deployments/handler.go#L55))
2. Backend valida request y resuelve `job_name` + `ssh_user`.
3. Backend dispara Jenkins con parametros de instancia (`INSTANCE_ID`, `PRODUCT_ID`, `ENV`, `TARGET_HOST`, etc.).
4. Jenkins ejecuta Ansible y Docker Compose en nodo cliente.
5. Jenkins resuelve puerto publicado y envia callback a ARK.
6. Backend registra ruta/estado de instancia.
7. Trafico a `/instances/<instance_id>/...` se resuelve dinamicamente al host/puerto final.

## Estados tipicos

- `queued`
- `provisioning`
- `ready`
- `failed`

## Fallas comunes

- Credenciales Jenkins invalidas.
- Host cliente no alcanzable por Tailscale.
- Error al levantar contenedor o resolver puerto web.
- Callback no entregado al backend.
