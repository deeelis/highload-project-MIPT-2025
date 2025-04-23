FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrator ./cmd/migrator/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /migrator /app/migrator
COPY --from=builder /app/migrations /app/migrations

CMD ["/app/migrator", \
     "-dsn=postgres://admin:password@storage_db:5432/storage?sslmode=disable", \
     "-migrations-path=/app/migrations", \
     "-action=up"]