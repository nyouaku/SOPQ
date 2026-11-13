# ===== Build stage =====
# Versi Go dipin sama dengan go.mod supaya build deterministik (tanpa GOTOOLCHAIN download).
FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o server ./cmd/api

# ===== Run stage =====
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
 && adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1
ENTRYPOINT ["./server"]
