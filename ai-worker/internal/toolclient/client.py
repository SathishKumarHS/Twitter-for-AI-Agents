import json
import logging

import grpc

from gen.twitterforaiworker.v1 import worker_pb2, worker_pb2_grpc
from internal.config import config

log = logging.getLogger("toolclient")


# gRPC client for calling the backend Go service to execute agent tools
class ToolClient:
    def __init__(self, address: str | None = None):
        # Open an async gRPC channel to the backend ToolExecutor gRPC address
        self.address = address or config.backend_tool_address()
        self._channel = grpc.aio.insecure_channel(self.address)
        self._stub = worker_pb2_grpc.ToolExecutorStub(self._channel)

    # Sends a ToolCall request to the backend and returns the dictionary result
    async def execute_tool(
        self,
        agent_id: int,
        name: str,
        arguments: dict,
        timeout: float = 30.0,
        mode: str = "",
    ) -> dict:
        # Build the gRPC protobuf tool call request
        req = worker_pb2.ToolCall(
            agent_id=agent_id,
            name=name,
            arguments=json.dumps(arguments or {}),
            mode=mode,
        )
        try:
            # Execute tool on backend via gRPC ExecuteTool RPC
            res = await self._stub.ExecuteTool(req, timeout=timeout)
            return {
                "name": res.name,
                "output": res.output,
                "error": res.error,
                "acted": res.acted,
            }
        except Exception as exc:  # noqa: BLE001
            # Log error and return fallback result if backend call fails
            log.warning("ExecuteTool(%s) failed: %s", name, exc)
            return {
                "name": name,
                "output": f"error: {exc}",
                "error": True,
                "acted": False,
            }
