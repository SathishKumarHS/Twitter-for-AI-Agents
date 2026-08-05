.PHONY: tidy proto build mocks

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

# Generate Go mock files for all test interfaces.
mocks:
	@export PATH="$$PATH:$$HOME/go/bin" && \
	if ! command -v mockgen > /dev/null 2>&1; then \
		echo "Installing mockgen (go.uber.org/mock/mockgen)..."; \
		go install go.uber.org/mock/mockgen@latest; \
	fi && \
	mkdir -p mocks && \
	mockgen -source=internal/db/agents/agents_interface.go -destination=mocks/mock_agents.go -package=mocks && \
	mockgen -source=internal/db/posts/post_interface.go -destination=mocks/mock_posts.go -package=mocks && \
	mockgen -source=internal/db/subscriptions/subscribe_interface.go -destination=mocks/mock_subscriptions.go -package=mocks && \
	mockgen -source=internal/db/last_seen/last_seen_interface.go -destination=mocks/mock_last_seen.go -package=mocks && \
	mockgen -source=internal/db/memories/memories_interface.go -destination=mocks/mock_memories.go -package=mocks && \
	mockgen -source=internal/dispatch/dispatch_interface.go -destination=mocks/mock_dispatcher.go -package=mocks && \
	mockgen -source=internal/events/events_interface.go -destination=mocks/mock_events.go -package=mocks && \
	mockgen -source=internal/redis/redis_interface.go -destination=mocks/mock_redis.go -package=mocks && \
	mockgen -source=internal/scheduler/scheduler_interface.go -destination=mocks/mock_scheduler.go -package=mocks && \
	mockgen -source=gen/twitterforaiworkerpb/worker_grpc.pb.go -destination=mocks/mock_worker_client.go -package=mocks && \
	echo "Mocks generated in mocks/"
.PHONY: mocks

# Compile and build the orchestrator binary into ../bin/.
build: proto
	mkdir -p ../bin
	go build -o ../bin/twitter-for-ai-agents-go ./cmd/server
.PHONY: build

# Build the backend-go image from the repo-root compose file.
docker:
	docker-compose -f ../docker-compose.yml build backend-go
.PHONY: docker

unit-test:
	go test ./...
.PHONY: unit-test