.PHONY: proto run docker

# Regenerate Python gRPC stubs from the shared proto into gen/.
#
# grpc_tools derives the generated import path from the proto file's location
# relative to -I. Staging the proto under a "gen/" prefix produces the package
# gen.agentfeed.v1, mirroring backend-go's gen/agentfeedpb layout.
# Requires the venv to have grpcio-tools installed:
#   .venv/bin/python -m pip install grpcio-tools
proto:
	rm -rf /tmp/agentfeed-proto && mkdir -p /tmp/agentfeed-proto/gen/agentfeed/v1
	cp ../proto/agentfeed/v1/worker.proto /tmp/agentfeed-proto/gen/agentfeed/v1/
	.venv/bin/python -m grpc_tools.protoc \
		-I /tmp/agentfeed-proto \
		--python_out=. --grpc_python_out=. \
		/tmp/agentfeed-proto/gen/agentfeed/v1/worker.proto
.PHONY: proto

# Run the worker locally (Ollama optional; mock fallback otherwise).
run:
	.venv/bin/python -m cmd.worker.main
.PHONY: run

# Build the ai-worker image from the repo-root compose file.
docker:
	docker-compose -f ../docker-compose.yml build ai-worker
.PHONY: docker