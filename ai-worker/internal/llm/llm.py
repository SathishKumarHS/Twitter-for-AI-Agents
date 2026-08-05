import asyncio
import json
import logging

import httpx

from internal.config import config

log = logging.getLogger("llm")


class LLMClient:
    """Thin client for Ollama's chat API with JSON-mode decision output."""

    def __init__(self, base_url: str | None = None, model: str | None = None):
        self.base_url = (base_url or config.ollama_url()).rstrip("/")
        self.model = model or config.ollama_model()
        self.max_retries = config.ollama_retries()
        self.backoff = config.ollama_backoff()

    async def decide(self, system: str, user: str, mode: str = "respond") -> dict:
        """Return a decision dict: {"action": "post|reply|ignore", "text": "..."}.

        Transient Ollama failures are retried with exponential backoff; after
        retries are exhausted a clear error is logged (the platform's only
        signal that the LLM is unavailable) and the call degrades to ignore.
        """
        last_err: Exception | None = None
        for attempt in range(1, self.max_retries + 1):
            try:
                return await self._post_once(system, user)
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

    async def _post_once(self, system: str, user: str) -> dict:
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
                    "options": {
                        "temperature": 0.8,
                        "top_p": 0.9,
                        "frequency_penalty": 0.5,
                        "presence_penalty": 0.5,
                    },
                },
            )
            resp.raise_for_status()
            content = resp.json()["message"]["content"]
            decision = json.loads(content)
            return self._sanitize(decision)

    def _sanitize(self, decision: dict) -> dict:
        action = str(decision.get("action", "ignore")).strip().lower()
        if action not in {"post", "reply", "ignore"}:
            action = "ignore"
        text = str(decision.get("text", "")).strip()
        text = self._clean_text(text)
        return {"action": action, "text": text}

    @staticmethod
    def _clean_text(text: str) -> str:
        """Strip stray model artifacts and enforce the 280-char limit."""
        text = text.strip()
        for marker in ("...", "…", "---"):
            text = text.removeprefix(marker).strip()
        # Drop a trailing run of stray emoji artifacts (e.g. the model appending 🧠).
        import re  # noqa: PLC0415

        text = re.sub(r"[\s]*[🧠❗👀🚨]+$", "", text).strip()
        if len(text) > 280:
            text = text[:280].rsplit(" ", 1)[0] if " " in text[:280] else text[:280]
        return text.strip()
