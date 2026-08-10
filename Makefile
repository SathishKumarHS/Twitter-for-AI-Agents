.PHONY: build proto up down clean run-ui unit-test mocks tidy integration integration-up integration-test integration-down

# Run tidy across backend-go and ai-worker modules
tidy:
	cd backend-go && make tidy
	cd ai-worker && make tidy

# Build the backend-go binary
build:
	cd backend-go && make build

# Generate Protobuf and gRPC stubs for backend-go and ai-worker
proto:
	cd backend-go && make proto
	cd ai-worker && make proto

# Generate mock interfaces for unit testing backend-go
mocks:
	cd backend-go && make mocks

# Run unit test suites for backend-go and ai-worker
unit-test:
	cd backend-go && make unit-test
	cd ai-worker && make unit-test

# Build Docker images and start all containerized services via Docker Compose
up:
	cd backend-go && make docker
	cd ai-worker && make docker
	docker-compose up -d

# Stop and remove running Docker Compose containers
down:
	docker-compose down

# Stop containers and remove persistent database volumes
clean:
	docker-compose down --remove-orphans
	docker volume rm twitter-for-ai-agents_pgdata 2>/dev/null || true

# Run the frontend web application in development mode
run-ui:
	cd web && make dev

# Run static linters across backend-go and web frontend
lint:
	cd backend-go && make lint
	cd web && make lint

# Start services in test environment mode for integration testing
integration-up:
	cd backend-go && make docker
	cd ai-worker && make docker
	ACTIVE_ENV=test docker-compose up -d

# Run backend integration tests
integration-test:
	cd backend-go && go test -tags=integration -count=1 -p 1 -v ./integration/...

# Stop test environment containers
integration-down:
	docker-compose down --remove-orphans

# Run full integration test lifecycle (start test environment, execute tests, tear down)
integration: integration-up
	$(MAKE) integration-test
	$(MAKE) integration-down