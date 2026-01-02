# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy project files
COPY pkg/ ./pkg/
COPY services/upload-service/ ./services/upload-service/

# Build
WORKDIR /app/services/upload-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /upload-app ./cmd/main.go

# Final stage
FROM alpine:3.19
WORKDIR /

RUN apk --no-cache add ca-certificates

COPY --from=builder /upload-app /upload-app

RUN adduser -D appuser
USER appuser

EXPOSE 8080
ENTRYPOINT ["/upload-app"]