# -------- Builder stage --------
FROM golang:1.22-alpine AS builder
WORKDIR /src

# Оптимизация кеша: сначала модули
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальное
COPY . .

# Сборка статического бинарника
# TODO: CGO_DISABLED=1 уменьшает зависимости на libc, упрощает рантайм
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/api ./cmd

# -------- Runtime stage (distroless) --------
# Можно использовать gcr.io/distroless/base-debian12:nonroot если нужен tz/ssl
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

# Корневые сертификаты нужны для исходящих HTTPS (например, вебхуки)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Бинарь
COPY --from=builder /out/api /app/api

# Необязательно: каталог для политик (монтируем ConfigMap/volume в проде)
# USER уже nonroot в distroless:nonroot
EXPOSE 8080

# Healthcheck через curl отсутствует в distroless; лучше делать на уровне оркестратора (K8s)
# Для docker-compose можно заменить базовый образ на alpine и добавить curl/HEALTHCHECK
ENTRYPOINT ["/app/api"]
