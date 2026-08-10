import asyncio
import logging

import grpc

from gen.twitterforaiworker.v1 import worker_pb2, worker_pb2_grpc
from internal.config import config
from internal.llm.llm import LLMClient
from internal.modes import MODE_PROACTIVE, MODE_RESPOND, VALID_MODES
from internal.prompts.prompts import build_system, build_tools, build_tools_section, build_user
from internal.toolclient.client import ToolClient

log = logging.getLogger("worker")


# gRPC service implementation that handles agent thinking and tool execution turns
class WorkerServicer(worker_pb2_grpc.WorkerServicer):
    def __init__(self, llm: LLMClient, tool_client: ToolClient):
        # Store LLM client, gRPC tool client, tool schemas, and max tool budget
        self.llm = llm
        self.tool_client = tool_client
        self.tools = build_tools()
        self.max_tools = config.max_tool_calls_per_turn()

    # Entry point for the Think gRPC RPC called by the Go backend orchestrator
    async def Think(self, request, context):  # noqa: N802
        try:
            # Validate turn mode (proactive vs respond)
            mode = request.mode or MODE_RESPOND
            if mode not in VALID_MODES:
                mode = MODE_RESPOND

            # Build system and user prompts for the agent
            system = f"{build_system(request.agent)}\n\n{build_tools_section()}"
            user = build_user(
                request.agent,
                request.post,
                mode,
            )

            # Run the agent deliberation loop with tools
            return await self._decide_with_tools(request, system, user, mode, context)
        except Exception as exc:  # noqa: BLE001
            log.exception("Think failed: %s", exc)
            return worker_pb2.ThinkResponse()

    # Core loop: queries LLM, executes requested tools, and applies final decisions
    async def _decide_with_tools(self, request, system: str, user: str, mode: str, context) -> worker_pb2.ThinkResponse:
        tool_calls: list[dict] = []
        tool_results: list[dict] = []
        executed = 0

        # Loop up to max_tools + 1 times (allowing 1 final turn after budget is used)
        for _ in range(self.max_tools + 1):
            # Check if gRPC deadline/timeout has expired
            if context is not None and context.done():
                log.warning("turn deadline expired before a final decision; returning ignore")
                return worker_pb2.ThinkResponse()

            # Query the LLM with system prompt, user prompt, and past tool results
            decision = await self.llm.decide(
                system,
                user,
                mode,
                tools=self.tools,
                tool_calls=tool_calls,
                tool_results=tool_results,
            )
            calls = decision.get("tool_calls")

            # Handle final JSON decision when LLM makes no more tool calls
            if not calls:
                response, used = await self._apply_final_decision(
                    request, decision, mode, context, tool_calls, tool_results
                )
                executed += used
                if response is not None:
                    return response
                if executed >= self.max_tools:
                    log.warning(
                        "write retries exhausted (%d attempts); returning ignore",
                        self.max_tools,
                    )
                    return worker_pb2.ThinkResponse()
                continue

            # Execute tool calls requested by the LLM
            for call in calls:
                # Ensure tool call matches current mode (rewrite post->reply or reply->post if needed)
                call = self._normalize_write_call(request, mode, call)
                if executed >= self.max_tools:
                    log.warning(
                        "tool budget exhausted (%d total calls); returning ignore",
                        self.max_tools,
                    )
                    return worker_pb2.ThinkResponse()

                # Call the backend gRPC tool executor
                result = await self.tool_client.execute_tool(
                    request.agent.id,
                    call["name"],
                    call.get("arguments") or {},
                    timeout=self._tool_timeout(context),
                    mode=mode,
                )
                executed += 1
                tool_calls.append(call)
                tool_results.append(result)
                log.info(
                    "think: agent=%s mode=%s post=%d tool=%s (error=%s, acted=%s)%s",
                    request.agent.handle,
                    mode,
                    request.post.id,
                    call["name"],
                    result.get("error"),
                    result.get("acted"),
                    self._error_suffix(result),
                )
                # If a post or reply was successfully created, end turn
                if result.get("acted"):
                    return worker_pb2.ThinkResponse(
                        executed=True,
                    )

        log.warning(
            "tool budget exhausted (%d calls) without a final decision; returning ignore",
            self.max_tools,
        )
        return worker_pb2.ThinkResponse()

    # Executes a final JSON action (post/reply) on the backend if no native tool call was used
    async def _apply_final_decision(
        self, request, decision: dict, mode: str, context, tool_calls: list, tool_results: list
    ) -> tuple[worker_pb2.ThinkResponse | None, int]:
        action = decision.get("action")
        if action not in {"post", "reply"}:
            return worker_pb2.ThinkResponse(), 0
        if mode == MODE_RESPOND and request.post.id <= 0:
            log.warning("write decision has no target post in respond mode; ignoring")
            return worker_pb2.ThinkResponse(), 0

        text = (decision.get("text") or "").strip()
        tool, args = self._write_tool_for_mode(request, mode, text)

        result = await self.tool_client.execute_tool(
            request.agent.id,
            tool,
            args,
            timeout=self._tool_timeout(context),
            mode=mode,
        )
        log.info(
            "think: agent=%s mode=%s post=%d tool=%s (error=%s, acted=%s)%s",
            request.agent.handle,
            mode,
            request.post.id,
            tool,
            result.get("error"),
            result.get("acted"),
            self._error_suffix(result),
        )
        if result.get("acted"):
            return worker_pb2.ThinkResponse(executed=True), 1

        tool_calls.append({"name": tool, "arguments": args})
        tool_results.append(result)
        return None, 1

    # Map action decision to legal tool based on mode (respond -> reply, proactive -> post)
    @staticmethod
    def _write_tool_for_mode(request, mode: str, text: str) -> tuple[str, dict]:
        if mode == MODE_RESPOND:
            return "reply", {"parent_id": request.post.id, "text": text}
        return "post", {"text": text}

    # Normalize tool call to enforce valid action for mode (e.g. force parent_id in respond mode)
    @staticmethod
    def _normalize_write_call(request, mode: str, call: dict) -> dict:
        name = call.get("name")
        args = dict(call.get("arguments") or {})
        text = args.get("text", "")
        if mode == MODE_RESPOND and name in {"post", "reply"} and request.post.id > 0:
            return {"name": "reply", "arguments": {"parent_id": request.post.id, "text": text}}
        if mode == MODE_PROACTIVE and name == "reply":
            return {"name": "post", "arguments": {"text": text}}
        return call

    # Helper: format error suffix for log messages
    @staticmethod
    def _error_suffix(result: dict) -> str:
        if not result.get("error"):
            return ""
        output = result.get("output", "")
        if not output:
            return ""
        return f' error_msg="{output}"'

    # Helper: calculate remaining deadline timeout for tool calls
    @staticmethod
    def _tool_timeout(context, default: float = 30.0) -> float:
        if context is None:
            return default
        remaining = context.time_remaining()
        if remaining is None:
            return default
        return max(0.5, min(default, remaining - 1))


# Starts the gRPC server listening for incoming Think requests from the orchestrator
async def serve() -> None:
    server = grpc.aio.server()
    worker_pb2_grpc.add_WorkerServicer_to_server(
        WorkerServicer(LLMClient(), ToolClient()), server
    )
    addr = config.listen_addr()
    server.add_insecure_port(addr)
    log.info("ai-worker gRPC listening on %s", addr)
    await server.start()
    await server.wait_for_termination()
