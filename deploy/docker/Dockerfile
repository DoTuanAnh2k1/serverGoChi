# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto

WORKDIR /build

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o app ./cmd/main/main.go

# ── Final stage ───────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -S app \
    && adduser  -S app -G app

WORKDIR /app

COPY --from=builder /build/app     ./app
COPY --from=builder /build/api.yaml ./api.yaml

RUN mkdir -p /data/subscribers /data/csv \
    && chown -R app:app /app /data
USER app

EXPOSE 3000

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -sf http://127.0.0.1:${SERVER_PORT:-3000}/health || exit 1

ENTRYPOINT ["./app"]
