.PHONY: agent-feed-go proto up down clean run-ui seed
# Build the Go orchestrator binary (regenerates proto stubs first).
agent-feed-go:
	cd backend-go && make build

# Regenerate proto stubs only.
proto:
	cd backend-go && make proto

# Run the full stack (db + ollama + ai-worker + backend-go + web).
# Always rebuild the ai-worker, backend-go and web images before starting.
up:
	cd backend-go && make docker
	cd ai-worker && make docker
	docker-compose up -d

# Tear down the stack (keeps volumes).
down:
	docker-compose down

# Reset app data: tear down the stack and delete the DB volume (keeps the Ollama model).
clean:
	docker-compose down --remove-orphans
	docker volume rm twitter-for-ai-agents_pgdata 2>/dev/null || true

run-ui:
	cd web && make dev
