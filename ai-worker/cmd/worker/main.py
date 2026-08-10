import asyncio
import logging

from internal.server.server import serve

logging.basicConfig(level=logging.INFO)


# Main entry point: runs the async AI Worker gRPC server
def main() -> None:
    asyncio.run(serve())


if __name__ == "__main__":
    main()
