"""gRPC server exposing Worker.Think, the worker's external API surface.

Mirrors backend-go/internal/httpapi: owns the network layer and the handler
(servicer) that translates requests into domain work.
"""
import asyncio
import logging

import grpc

from gen.agentfeed.v1 import worker_pb2, worker_pb2_grpc
from internal.config import config
from internal.llm.llm import LLMClient
from internal.prompts.prompts import build_system, build_user

log = logging.getLogger("worker")


class WorkerServicer(worker_pb2_grpc.WorkerServicer):
    """Implements the Think RPC: agent + post (+ recent context) -> decision."""

    def __init__(self, llm: LLMClient):
        self.llm = llm

    async def Think(self, request, context):  # noqa: N802
        try:
            memories = list(request.memories)
            system = build_system(request.agent, memories)
            user = build_user(
                request.agent,
                request.post,
                list(request.recent),
                request.mode or "respond",
            )
            decision = await self.llm.decide(system, user, request.mode or "respond")
            log.info(
                "think: agent=%s mode=%s post=%d -> action=%s",
                request.agent.handle,
                request.mode or "respond",
                request.post.id,
                decision["action"],
            )
            return worker_pb2.ThinkReply(action=decision["action"], text=decision["text"])
        except Exception as exc:  # noqa: BLE001
            log.exception("Think failed: %s", exc)
            return worker_pb2.ThinkReply(action="ignore", text="")


async def serve() -> None:
    server = grpc.aio.server()
    worker_pb2_grpc.add_WorkerServicer_to_server(WorkerServicer(LLMClient()), server)
    addr = config.listen_addr()
    server.add_insecure_port(addr)
    log.info("ai-worker gRPC listening on %s", addr)
    await server.start()
    await server.wait_for_termination()


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(serve())
