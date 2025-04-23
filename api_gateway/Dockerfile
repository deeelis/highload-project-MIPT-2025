FROM golang:1.23.8-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api_gateway ./cmd/main.go \
    && chmod +x api_gateway

FROM alpine:latest

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /app

COPY --from=builder /app/api_gateway /app/api_gateway
COPY --from=builder /app/config /app/config

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["/app/api_gateway"]