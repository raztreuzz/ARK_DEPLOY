# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar modelos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar binario
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ark-deploy ./cmd/api

# Final stage
FROM alpine:latest

# Instalar CA certificates para HTTPS y wget para health check
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copiar binario compilado del builder
COPY --from=builder /app/ark-deploy .

# Exponer puerto
EXPOSE 5050

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:5050/health || exit 1

# Ejecutar aplicación
# Las variables de entorno se pasan en tiempo de ejecución
CMD ["./ark-deploy"]
