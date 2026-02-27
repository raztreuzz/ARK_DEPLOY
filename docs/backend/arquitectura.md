# Arquitectura Backend

## Explicación

ARK_DEPLOY desacopla el despliegue (Jenkins/Ansible) del acceso web a instancias (API + proxy).  
El backend no ejecuta directamente contenedores en todos los nodos; coordina jobs, recibe callback y mantiene el estado/ruteo de cada instancia.

Componentes clave:

- API (Go/Gin): validación, orquestación y endpoints operativos.
- Jenkins: ejecución del pipeline de deploy por instancia.
- Redis/Storage: registro de rutas y estado de instancias.
- Nginx: puerta de entrada HTTP para frontend y rutas de instancia.

## Ruteo dinámico

Objetivo: resolver `instance_id -> target_host:target_port` en tiempo de request.

Secuencia:

1. Se crea una instancia con `instance_id` único.
2. Jenkins despliega en el cliente y descubre puerto publicado.
3. Jenkins envía callback al backend con host/puerto final.
4. El backend persiste el mapping.
5. Solicitudes a `/instances/<instance_id>/...` se enrutan al destino resuelto.

Ventaja: la URL pública permanece estable aunque cambie el puerto interno.

## Callback

El callback cierra el flujo asíncrono de despliegue.

Responsabilidades:

- Confirmar resultado de despliegue.
- Registrar `target_host`, `target_port`, `container_name`, URLs de acceso.
- Cambiar estado de la instancia de `provisioning` a estado operativo.

Sin callback, la API conoce que el job fue lanzado, pero no tendría certeza del destino final para enrutar tráfico.

## Nginx

Nginx cumple rol de entrypoint HTTP:

- Sirve frontend.
- Reenvía rutas backend.
- Mantiene estable la entrada `/instances/...`.

Principio de diseño:

- Nginx no “sabe” cada cliente/producto.
- El backend decide el destino dinámicamente con el mapping de instancia.

Esto evita configuraciones estáticas por cliente y reduce mantenimiento manual.

## Multi-instancia

El modelo soporta múltiples instancias concurrentes:

- `instance_id` único por despliegue.
- Aislamiento por `product_id` y `environment`.
- Mapping independiente por instancia (host/puerto propios).

Resultado: varias instancias pueden coexistir en distintos clientes sin colisión de rutas públicas.

## Seguridad

Controles implementados/recomendados:

- Secretos fuera de código (credenciales Jenkins/env vars).
- Validación estricta de entradas (`target_host`, `environment`, ids, job names).
- Restricción de nombres para evitar payloads peligrosos en comandos/jobs.
- Segmentación de acceso vía Tailscale y políticas de red.
- Uso de callback controlado por backend para registrar estado final.

Buenas prácticas operativas:

- Rotar tokens de Jenkins/Tailscale.
- Usar usuarios SSH con privilegios mínimos.
- Restringir origen de callback por red/ACL.
