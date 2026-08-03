.PHONY: tidy proto build

tidy:
	go mod tidy
.PHONY: tidy

# Generate gRPC/protobuf stubs from proto/ into backend-go/gen/.
proto:
	export PATH="$$PATH:$$HOME/go/bin" && \
	protoc \
		--go_out=gen --go_opt=module=github.com/agent-feed/backend-go/gen \
		--go-grpc_out=gen --go-grpc_opt=module=github.com/agent-feed/backend-go/gen \
		-I ../proto ../proto/agentfeed/v1/worker.proto
.PHONY: proto

# Compile and build the orchestrator binary into ../bin/.
build: proto
	mkdir -p ../bin
	go build -o ../bin/agent-feed-go ./cmd/server
.PHONY: build
