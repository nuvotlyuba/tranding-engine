FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /build

# Копируем зависимости отдельно от кода — используем layer caching
# Если код изменился но go.mod не менялся — зависимости не перекачиваются
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
# CGO_ENABLED=0 — статическая сборка, не зависит от libc
# -trimpath — убирает пути к файлам из бинарника (безопасность)
# -ldflags:
#   -w — убирает debug info (меньше размер)
#   -s — убирает symbol table (меньше размер)
#   -extldflags=-static — полностью статический бинарник
RUN CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-w -s -extldflags=-static" \
    -o /build/tranding-engine \
    ./cmd/engine

# =============================================================================
# Этап 2 — финальный образ
# scratch — абсолютно пустой образ, только наш бинарник
# Никакого shell, никаких утилит — минимальная поверхность атаки
# =============================================================================
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /build/tranding-engine /tranding-engine


# Запускаем от непривилегированного пользователя
# В scratch нельзя создать пользователя через useradd
# Используем числовой UID — соглашение: 65534 = nobody
USER 65534:65534

# Healthcheck встроенный в образ
# В scratch нет curl/wget — используем встроенный healthcheck Go приложения

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/trading-engine", "-health-check"]

ENTRYPOINT ["/trading-engine"]
