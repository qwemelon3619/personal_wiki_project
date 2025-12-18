# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Download dependencies
ARG TARGETOS
ARG TARGETARCH

COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/editor-service/go.mod services/editor-service/go.sum ./services/editor-service/

RUN go mod download

COPY . .

# Build the editor service
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-arm64} \
    go build -o /editor-app ./services/editor-service/cmd/main.go

# Final stage
FROM alpine:latest
WORKDIR /
COPY --from=builder /editor-app /editor-app
EXPOSE 8080
ENTRYPOINT ["/editor-app"]