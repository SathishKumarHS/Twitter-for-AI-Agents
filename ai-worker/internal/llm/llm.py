import asyncio
import json
import logging
import re

import httpx

from internal.config import config
from internal.modes import MODE_RESPOND

log = logging.getLogger("llm")

# Single source of truth for permitted tools, enforced on both AI Worker and Go Backend
ALLOWED_TOOLS = frozenset({"get_feed", "get_thread", "get_memories", "post", "reply"})


# Client for interacting with Ollama's HTTP chat API, supporting tool calling and retry backoff
class LLMClient:
    def __init__(self, base_url: str | None = None, model: str | None = None):
        # Initialize Ollama URL, model, max retries, and backoff delay settings
        self.base_url = (base_url or config.ollama_url()).rstrip("/")
        self.model = model or config.ollama_model()
        self.max_retries = config.ollama_retries()
        self.backoff = config.ollama_backoff()

    # Queries Ollama chat with exponential backoff retries; falls back to 'ignore' on persistent failure
    async def decide(
        self,
        system: str,
        user: str,
        mode: str = MODE_RESPOND,
        tools: list[dict] | None = None,
        tool_calls: list[dict] | None = None,
        tool_results: list[dict] | None = None,
    ) -> dict:
        last_err: Exception | None = None
        # Attempt LLM call with exponential backoff retries
        for attempt in range(1, self.max_retries + 1):
            try:
                return await self._post_once(
                    system, user, tools, tool_calls, tool_results
                )
            except Exception as exc:  # noqa: BLE001
                last_err = exc
                log.warning(
                    "Ollama chat failed (attempt %d/%d): %s",
                    attempt,
                    self.max_retries,
                    exc,
                )
                if attempt < self.max_retries:
                    await asyncio.sleep(self.backoff * (2 ** (attempt - 1)))
        log.error(
            "Ollama chat failed after %d attempts (model=%s): %s",
            self.max_retries,
            self.model,
            last_err,
        )
        return {"action": "ignore", "text": ""}

    # Rebuilds conversation transcript and sends a single HTTP POST request to Ollama /api/chat
    async def _post_once(
        self,
        system: str,
        user: str,
        tools: list[dict] | None,
        tool_calls: list[dict] | None,
        tool_results: list[dict] | None,
    ) -> dict:
        messages = [
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ]
        # Rebuild full conversation history with past tool calls and results
        calls = list(tool_calls or [])
        results = list(tool_results or [])
        for i, call in enumerate(calls):
            args = call.get("arguments", {})
            if isinstance(args, str):
                try:
                    args = json.loads(args)
                except (json.JSONDecodeError, TypeError):
                    args = {}
            messages.append(
                {
                    "role": "assistant",
                    "content": "",
                    "tool_calls": [
                        {"function": {"name": call.get("name", ""), "arguments": args}}
                    ],
                }
            )
            if i < len(results):
                result = results[i]
                output = result.get("output", "")
                if result.get("error"):
                    output = f"error: {output}"
                messages.append(
                    {"role": "tool", "name": call.get("name", ""), "content": output}
                )

        body: dict = {
            "model": self.model,
            "messages": messages,
            "stream": False,
            "options": {
                "temperature": 0.8,
                "top_p": 0.9,
                "frequency_penalty": 0.5,
                "presence_penalty": 0.5,
            },
        }
        if tools:
            body["tools"] = tools
        else:
            body["format"] = "json"

        # Make HTTP POST request to Ollama chat API
        async with httpx.AsyncClient(timeout=config.ollama_timeout()) as client:
            resp = await client.post(f"{self.base_url}/api/chat", json=body)
            resp.raise_for_status()
            return self._parse_message(resp.json()["message"])

    # Parses the Ollama assistant response into either tool_calls or a final decision dictionary
    def _parse_message(self, msg: dict) -> dict:
        native_calls = msg.get("tool_calls") or []
        if native_calls:
            parsed = []
            for tc in native_calls:
                fn = tc.get("function", {}) or {}
                sanitized = self._sanitize_tool_call(fn)
                if sanitized["name"]:
                    parsed.append(sanitized)
            if parsed:
                return {"tool_calls": parsed}

        content = (msg.get("content") or "").strip()
        if not content:
            return self._sanitize({})
        try:
            decision = json.loads(content)
        except (json.JSONDecodeError, TypeError):
            return self._sanitize({})

        tc = decision.get("tool_call")
        if isinstance(tc, dict):
            sanitized = self._sanitize_tool_call(tc)
            if sanitized["name"]:
                return {"tool_calls": [sanitized]}
        return self._sanitize(decision)

    # Validates tool name against ALLOWED_TOOLS whitelist and ensures arguments are a dictionary
    def _sanitize_tool_call(self, fn: dict) -> dict:
        name = str(fn.get("name", "")).strip().lower()
        if name not in ALLOWED_TOOLS:
            return {"name": "", "arguments": {}}
        args = fn.get("arguments") or {}
        if isinstance(args, str):
            try:
                args = json.loads(args)
            except (json.JSONDecodeError, TypeError):
                args = {}
        if not isinstance(args, dict):
            args = {}
        return {"name": name, "arguments": args}

    # Normalizes final action (post/reply/ignore) and cleans up output text
    def _sanitize(self, decision: dict) -> dict:
        action = str(decision.get("action", "ignore")).strip().lower()
        if action not in {"post", "reply", "ignore"}:
            action = "ignore"
        text = str(decision.get("text", "")).strip()
        text = self._clean_text(text)
        return {
            "action": action,
            "text": text,
        }

    # Strips leading model artifacts, trailing noise emojis, and caps length at 280 characters
    @staticmethod
    def _clean_text(text: str) -> str:
        text = text.strip()
        for marker in ("...", "…", "---"):
            text = text.removeprefix(marker).strip()
        text = re.sub(r"[\s]*[🧠❗👀🚨]+$", "", text).strip()
        if len(text) > 280:
            text = text[:280].rsplit(" ", 1)[0] if " " in text[:280] else text[:280]
        return text.strip()
