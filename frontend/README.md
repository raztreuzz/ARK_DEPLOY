# ARK Deploy - Frontend

Frontend del sistema ARK Deploy construido con React + Vite.

## 🚀 Tecnologías

- **React 18** - Biblioteca de UI
- **Vite** - Build tool y dev server
- **Lucide React** - Iconos
- **Tailwind CSS** - Estilos (incluidos en el código)

## 📦 Estructura del Proyecto

```
frontend/
├── src/
│   ├── App.jsx          # Componente principal
│   ├── main.jsx         # Punto de entrada
│   └── index.css        # Estilos globales
├── Dockerfile           # Configuración Docker multi-stage
├── nginx.conf           # Configuración Nginx para producción
├── vite.config.js       # Configuración Vite
└── package.json         # Dependencias
```

## 🐳 Ejecutar con Docker

### Usando Docker Compose (Recomendado)

Desde la raíz del proyecto:

```bash
docker-compose up ark-frontend
```

O para construir y ejecutar todo el stack:

```bash
docker-compose up --build
```

### Docker standalone

```bash
# Construir imagen
docker build -t ark-frontend .

# Ejecutar contenedor
docker run -p 3000:3000 ark-frontend
```

## 💻 Desarrollo Local (sin Docker)

Si necesitas desarrollar localmente:

```bash
# Instalar dependencias
npm install

# Ejecutar en modo desarrollo
npm run dev

# Construir para producción
npm run build

# Preview de la build
npm run preview
```

## 🔗 Conexión con Backend

El frontend se comunica con el backend a través de:

- **Producción (Docker)**: Proxy configurado en `nginx.conf` (`/api/*` → `http://ark-deploy:5050/`)
- **Desarrollo (Vite)**: Proxy configurado en `vite.config.js`

## 📝 Características

- ✅ Panel de control de productos ARK
- ✅ Gestión de despliegues
- ✅ Visualización de nodos Tailscale
- ✅ Logs en tiempo real
- ✅ Interfaz responsive
- ✅ Build optimizado para producción

## 🌐 Acceso

Una vez ejecutado:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:5050

## 🏗️ Build Multi-Stage

El Dockerfile utiliza una build multi-stage:

1. **Builder**: Node 20 Alpine - Instala deps y construye la app
2. **Production**: Nginx Alpine - Sirve archivos estáticos optimizados

Resultado: Imagen final ~25MB (vs ~500MB con Node completo)
