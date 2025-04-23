FROM golang:1.23.8-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth_service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /auth_service /app/auth_service
COPY --from=builder /app/config /app/config

EXPOSE 50051

CMD ["/app/auth_service"]