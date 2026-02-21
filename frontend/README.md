# ARK_DEPLOY Frontend

Frontend React + Vite para el sistema de gestión de despliegues ARK_DEPLOY.

## Tecnologías

- React 18 - Framework de UI
- Vite - Herramienta de construcción y servidor de desarrollo
- Lucide React - Iconos
- Tailwind CSS - Estilos
- Nginx - Servidor web de producción

## Estructura del Proyecto

```
frontend/
├── src/
│   ├── App.jsx          # Componente principal
│   ├── main.jsx         # Punto de entrada
│   └── index.css        # Estilos globales
├── Dockerfile           # Build multi-stage de Docker
├── nginx.conf           # Configuración Nginx
├── vite.config.js       # Configuración Vite
└── package.json         # Dependencias
```

## Desarrollo Local

```bash
# Instalar dependencias
npm install

# Servidor de desarrollo
npm run dev

# Build de producción
npm run build

# Vista previa del build
npm run preview
```

Accede en http://localhost:3000

## Docker

### Usando Docker Compose

Desde la raíz del proyecto:

```bash
docker-compose up ark-frontend
```

O construir todo:

```bash
docker-compose up --build
```

### Docker Standalone

```bash
# Construir
docker build -t ark-frontend .

# Ejecutar
docker run -p 3000:3000 ark-frontend
```

Accede en http://localhost:3000

## Integración con Backend

El frontend se comunica con la API del backend:

- **Producción (Docker)**: Proxy a través de `nginx.conf` (`/api/*` -> `http://ark-deploy:5050/`)
- **Desarrollo (Vite)**: Proxy a través de `vite.config.js`

El backend se ejecuta en http://localhost:5050

## Características

- Panel de control de gestión de productos
- Orquestación de despliegues
- Visualización de árbol de dispositivos Tailscale
- Logs en tiempo real
- Interfaz responsiva
- Build optimizado para producción

## Imagen Docker

Build multi-stage:

1. Builder: Node 20 Alpine - Instala dependencias, construye app
2. Producción: Nginx Alpine - Sirve archivos optimizados

Resultado: ~25MB de imagen final

## Estándares de Desarrollo

- Usa componentes funcionales
- Importa iconos desde `lucide-react`
- Estilos con clases Tailwind CSS
- Llamadas API a través del prefijo `/api/*`

---

**Versión**: 1.0.0 | **Última actualización**: Febrero 2026
