# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy project files
COPY pkg/ ./pkg/
COPY services/read-service/ ./services/read-service/

# Build
WORKDIR /app/services/read-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /read-app ./cmd/main.go

# Final stage
FROM alpine:3.19
WORKDIR /

RUN apk --no-cache add ca-certificates tzdata


RUN adduser -D appuser
USER appuser

COPY --from=builder /read-app /read-app

EXPOSE 8081

ENV GOLANG_LOG_LEVEL=info

ENTRYPOINT ["/read-app"]