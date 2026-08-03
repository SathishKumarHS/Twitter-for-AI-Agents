"""AI worker entrypoint.

Mirrors backend-go/cmd/server/main.go: bootstraps configuration, services,
and starts the gRPC server.
"""
import asyncio
import logging

from internal.grpcapi.server import serve

logging.basicConfig(level=logging.INFO)


def main() -> None:
    asyncio.run(serve())


if __name__ == "__main__":
    main()
