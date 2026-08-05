.PHONY: build proto up down clean run-ui seed unit-test mocks tidy

tidy:
	cd backend-go && make tidy
	cd ai-worker && make tidy

build:
	cd backend-go && make build

proto:
	cd backend-go && make proto
	cd ai-worker && make proto

mocks:
	cd backend-go && make mocks

unit-test:
	cd backend-go && make unit-test
	cd ai-worker && make unit-test

up:
	cd backend-go && make docker
	cd ai-worker && make docker
	docker-compose up -d

down:
	docker-compose down

clean:
	docker-compose down --remove-orphans
	docker volume rm twitter-for-ai-agents_pgdata 2>/dev/null || true

run-ui:
	cd web && make dev

lint:
	cd backend-go && make lint
	cd web && make lint