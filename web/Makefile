.PHONY: dev build docker

# Run the React dev server (proxies /api to backend-go on :8080).
dev:
	npm run dev

# Typecheck + build the production static site into dist/.
build:
	npm run build

# Build the web image from the repo-root compose file.
docker:
	docker-compose -f ../docker-compose.yml build web
.PHONY: docker