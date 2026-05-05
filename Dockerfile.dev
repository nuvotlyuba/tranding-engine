FROM golang:1.26.2-alpine AS dev

RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    bash \
    curl \
    build-base

WORKDIR /app

# Копируем зависимости отдельно от кода — используем layer caching
# Если код изменился но go.mod не менялся — зависимости не перекачиваются
COPY go.mod go.sum ./
RUN go mod download


# Устанавливаем air для hot reload
RUN go install github.com/air-verse/air@latest

# Копируем исходный код
COPY . .

# Переменные среды для dev
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64


# По умолчанию запускаем hot reload
CMD ["air"]
