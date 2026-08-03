import json
import logging

import httpx

from internal.config import config

log = logging.getLogger("llm")


class LLMClient:
    """Thin client for Ollama's chat API with JSON-mode decision output.

    Falls back to a deterministic mock when Ollama is unreachable so the
    demo loop keeps running without the model.
    """

    def __init__(self, base_url: str | None = None, model: str | None = None):
        self.base_url = (base_url or config.ollama_url()).rstrip("/")
        self.model = model or config.ollama_model()
        self._available: bool | None = None

    async def decide(self, system: str, user: str, mode: str = "respond") -> dict:
        """Return a decision dict: {"action": "post|reply|ignore", "text": "..."}."""
        if not await self._ping():
            log.warning("Ollama unavailable; using mock decision")
            return self._mock_decide(user, mode)
        try:
            async with httpx.AsyncClient(timeout=config.ollama_timeout()) as client:
                resp = await client.post(
                    f"{self.base_url}/api/chat",
                    json={
                        "model": self.model,
                        "messages": [
                            {"role": "system", "content": system},
                            {"role": "user", "content": user},
                        ],
                        "stream": False,
                        "format": "json",
                    },
                )
                resp.raise_for_status()
                content = resp.json()["message"]["content"]
                decision = json.loads(content)
                return self._sanitize(decision)
        except Exception as exc:  # noqa: BLE001
            log.error("Ollama chat failed: %s", exc)
            return self._mock_decide(user, mode)

    async def _ping(self) -> bool:
        if self._available is not None:
            return self._available
        try:
            async with httpx.AsyncClient(timeout=5) as client:
                resp = await client.get(f"{self.base_url}/api/tags")
                self._available = resp.status_code == 200
        except Exception:  # noqa: BLE001
            self._available = False
        return self._available

    def _sanitize(self, decision: dict) -> dict:
        action = str(decision.get("action", "ignore")).strip().lower()
        if action not in {"post", "reply", "ignore"}:
            action = "ignore"
        text = str(decision.get("text", "")).strip()
        return {"action": action, "text": text}

    def _mock_decide(self, user: str, mode: str = "respond") -> dict:
        """Deterministic fallback: reply when asked something, else ignore."""
        if mode == "proactive":
            return {
                "action": "post",
                "text": "Just finished a fresh brew of my favorite topic today. "
                "Anyone else on the feed following along?",
            }
        lower = user.lower()
        if "mention" in lower or "@" in user or "?" in user:
            return {"action": "reply", "text": "Interesting point! Tell me more."}
        if "hello" in lower or "hi " in lower or "hey" in lower:
            return {"action": "reply", "text": "Hey there! Good to see you on the feed."}
        return {"action": "ignore", "text": ""}
