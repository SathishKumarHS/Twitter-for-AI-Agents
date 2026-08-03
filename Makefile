.PHONY: agent-feed-go proto up down logs
# Build the Go orchestrator binary (regenerates proto stubs first).
agent-feed-go:
	cd backend-go && make build

# Regenerate proto stubs only.
proto:
	cd backend-go && make proto

# Run the full stack (db + ollama + ai-worker + backend-go).
# Always rebuild the ai-worker and backend-go images before starting.
up:
	cd backend-go && make docker
	cd ai-worker && make docker
	docker-compose up -d

# Tear down the stack (keeps volumes).
down:
	docker-compose down
