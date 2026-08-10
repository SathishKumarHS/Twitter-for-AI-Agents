# Multi-stage Dockerfile: Build stage using Go 1.26 alpine
FROM golang:1.26-alpine AS build

WORKDIR /src

# Download and cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build backend server binary
COPY . .
RUN mkdir -p /out && go build -o /out/twitter-for-ai-agents-go ./cmd/server

# Final runtime stage using lightweight Alpine Linux image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/twitter-for-ai-agents-go /app/twitter-for-ai-agents-go
COPY --from=build /src/config /app/config

# Expose HTTP REST API server port
EXPOSE 8080

# Start backend orchestrator server
CMD ["/app/twitter-for-ai-agents-go"]
