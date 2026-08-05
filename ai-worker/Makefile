.PHONY: proto run docker

proto:
	rm -rf /tmp/twitterforaiworker-proto && mkdir -p /tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1
	cp ../proto/twitterforaiworker/v1/worker.proto /tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1/
	.venv/bin/python -m grpc_tools.protoc \
		-I /tmp/twitterforaiworker-proto \
		--python_out=. --grpc_python_out=. \
		/tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1/worker.proto

# Run the worker locally (Ollama optional; mock fallback otherwise).
run:
	.venv/bin/python -m cmd.worker.main
.PHONY: run

# Build the ai-worker image from the repo-root compose file.
docker:
	docker-compose -f ../docker-compose.yml build ai-worker
.PHONY: docker