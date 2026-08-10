# Python binary selector: uses virtualenv python if .venv exists, else system python3
PYTHON := $(shell if [ -d ".venv" ]; then echo ".venv/bin/python"; else echo "python3"; fi)

.PHONY: tidy proto run docker unit-test

# Install and upgrade Python dependencies from requirements.txt
tidy:
	$(PYTHON) -m pip install --upgrade pip
	$(PYTHON) -m pip install -r requirements.txt

# Generate gRPC and Protobuf Python stubs from ../proto/
proto:
	rm -rf /tmp/twitterforaiworker-proto && mkdir -p /tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1
	cp ../proto/twitterforaiworker/v1/worker.proto /tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1/
	$(PYTHON) -m grpc_tools.protoc \
		-I /tmp/twitterforaiworker-proto \
		--python_out=. --grpc_python_out=. \
		/tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1/worker.proto

# Run the AI worker service locally
run:
	$(PYTHON) -m cmd.worker.main

# Build the ai-worker container image using docker-compose
docker:
	docker-compose -f ../docker-compose.yml build ai-worker

# Discover and run unit tests under tests/unit-test
unit-test:
	$(PYTHON) -m unittest discover -s tests/unit-test -p "test_*.py"