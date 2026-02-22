FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o ark-deploy ./cmd/api

FROM alpine:3.20

RUN apk --no-cache add ca-certificates docker-cli docker-cli-compose

WORKDIR /root/

COPY --from=builder /app/ark-deploy .

EXPOSE 5050

CMD ["./ark-deploy"]