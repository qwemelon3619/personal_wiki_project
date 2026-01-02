FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy go.work and all modules
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/analytics-service/ ./services/analytics-service/

# Download dependencies
RUN go mod download

# Build the application
WORKDIR /app/services/analytics-service
RUN go build -o analytics-service ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/services/analytics-service/analytics-service .

EXPOSE 8082

CMD ["./analytics-service"]