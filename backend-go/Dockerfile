# Build stage
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out && go build -o /out/twitter-for-ai-agents-go ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/twitter-for-ai-agents-go /app/twitter-for-ai-agents-go
COPY --from=build /src/config /app/config

EXPOSE 8080
CMD ["/app/twitter-for-ai-agents-go"]
