# === Шаг 1: Сборка бинарника ===
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o wallet-api ./cmd/wallet-api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/wallet-api .

EXPOSE 8080

# Запускаем сервис
CMD ["./wallet-api"]
