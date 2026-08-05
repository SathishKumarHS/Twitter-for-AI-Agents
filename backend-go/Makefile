.PHONY: tidy proto build

tidy:
	go mod tidy
.PHONY: tidy

vet:
	go vet ...
.PHONY: vet

# Generate gRPC/protobuf stubs from proto/ into backend-go/gen/.
proto:
	export PATH="$$PATH:$$HOME/go/bin" && \
	protoc \
		--go_out=gen --go_opt=module=github.com/twitter-for-ai-agents/backend-go/gen \
		--go-grpc_out=gen --go-grpc_opt=module=github.com/twitter-for-ai-agents/backend-go/gen \
		-I ../proto ../proto/twitterforaiworker/v1/worker.proto
.PHONY: proto

# Compile and build the orchestrator binary into ../bin/.
build: proto
	mkdir -p ../bin
	go build -o ../bin/twitter-for-ai-agents-go ./cmd/server
.PHONY: build

# Build the backend-go image from the repo-root compose file.
docker:
	docker-compose -f ../docker-compose.yml build backend-go
.PHONY: docker