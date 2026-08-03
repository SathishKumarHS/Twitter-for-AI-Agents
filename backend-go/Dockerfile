# Build stage
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out && go build -o /out/agent-feed-go ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/agent-feed-go /app/agent-feed-go

EXPOSE 8080
CMD ["/app/agent-feed-go"]
