FROM golang:1.23.8-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY config/ ./config/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o text-analyzer ./cmd/main.go

FROM alpine:3.18
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/text-analyzer .
COPY --from=builder /app/config ./config
CMD ["./text-analyzer"]