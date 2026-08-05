PYTHON := $(shell if [ -d ".venv" ]; then echo ".venv/bin/python"; else echo "python3"; fi)

.PHONY: tidy proto run docker unit-test

tidy:
	$(PYTHON) -m pip install --upgrade pip
	$(PYTHON) -m pip install -r requirements.txt

proto:
	rm -rf /tmp/twitterforaiworker-proto && mkdir -p /tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1
	cp ../proto/twitterforaiworker/v1/worker.proto /tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1/
	$(PYTHON) -m grpc_tools.protoc \
		-I /tmp/twitterforaiworker-proto \
		--python_out=. --grpc_python_out=. \
		/tmp/twitterforaiworker-proto/gen/twitterforaiworker/v1/worker.proto

run:
	$(PYTHON) -m cmd.worker.main

docker:
	docker-compose -f ../docker-compose.yml build ai-worker

unit-test:
	$(PYTHON) -m unittest discover -s tests/unit-test -p "test_*.py"