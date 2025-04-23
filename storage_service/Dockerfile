FROM golang:1.23.8-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /storage_service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /storage_service /app/storage_service
COPY --from=builder /app/config /app/config

#COPY migrations ./migrations

EXPOSE 50052

CMD ["/app/storage-service"]