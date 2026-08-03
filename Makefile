.PHONY: agent-feed-go proto

# Build the Go orchestrator binary (regenerates proto stubs first).
agent-feed-go:
	cd backend-go && make build

# Regenerate proto stubs only.
proto:
	cd backend-go && make proto