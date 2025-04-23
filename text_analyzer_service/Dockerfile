FROM golang:1.23.8-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
#COPY config/ ./config/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /text_analyzer_service ./cmd/main.go

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /text_analyzer_service /app/text_analyzer_service
COPY --from=builder /app/config /app/config
CMD ["/app/text_analyzer_service"]