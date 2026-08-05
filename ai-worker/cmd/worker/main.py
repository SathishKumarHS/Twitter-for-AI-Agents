import asyncio
import logging

from internal.server.server import serve

logging.basicConfig(level=logging.INFO)


def main() -> None:
    asyncio.run(serve())


if __name__ == "__main__":
    main()
