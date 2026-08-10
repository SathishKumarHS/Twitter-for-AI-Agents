import asyncio
from pathlib import Path
import sys
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

# Ensure ai-worker root is in sys.path
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from internal.llm.llm import ALLOWED_TOOLS, LLMClient


class TestLLMClient(unittest.IsolatedAsyncioTestCase):
    """Unit tests for LLMClient decision making, sanitization, and error handling."""

    def setUp(self) -> None:
        self.client = LLMClient(base_url="http://test-ollama:11434", model="test-model")
        self.client.max_retries = 2
        self.client.backoff = 0.01

    def test_clean_text_artifact_stripping(self) -> None:
        """Test that _clean_text strips leading markers and trailing emojis."""
        raw = "... Hello world! 🧠"
        cleaned = LLMClient._clean_text(raw)
        self.assertEqual(cleaned, "Hello world!")

    def test_clean_text_character_truncation(self) -> None:
        """Test that _clean_text truncates text longer than 280 characters."""
        long_text = "word " * 100  # 500 characters
        cleaned = LLMClient._clean_text(long_text)
        self.assertLessEqual(len(cleaned), 280)

    def test_sanitize_action(self) -> None:
        """Test that _sanitize normalizes valid actions and defaults invalid actions to ignore."""
        self.assertEqual(
            self.client._sanitize({"action": " POST ", "text": "Hi"}),
            {"action": "post", "text": "Hi"},
        )
        self.assertEqual(
            self.client._sanitize({"action": "invalid_action", "text": "Hi"}),
            {"action": "ignore", "text": "Hi"},
        )

    @patch("httpx.AsyncClient.post")
    async def test_decide_success(self, mock_post: MagicMock) -> None:
        """Test successful decide call with mocked httpx post response."""
        mock_response = MagicMock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = {
            "message": {
                "content": '{"action": "reply", "text": "Great post!"}'
            }
        }
        mock_post.return_value = mock_response

        decision = await self.client.decide(system="sys prompt", user="user prompt")
        self.assertEqual(
            decision,
            {"action": "reply", "text": "Great post!"},
        )

    @patch("logging.Logger.error")
    @patch("logging.Logger.warning")
    @patch("httpx.AsyncClient.post")
    @patch("asyncio.sleep", new_callable=AsyncMock)
    async def test_decide_failure_retry_fallback(
        self,
        mock_sleep: AsyncMock,
        mock_post: MagicMock,
        mock_log_warn: MagicMock,
        mock_log_err: MagicMock,
    ) -> None:
        """Test that decide retries on error and falls back to ignore when retries exhaust."""
        mock_post.side_effect = Exception("Ollama Connection Refused")

        decision = await self.client.decide(system="sys prompt", user="user prompt")
        self.assertEqual(
            decision,
            {"action": "ignore", "text": ""},
        )
        self.assertEqual(mock_post.call_count, 2)

    @patch("httpx.AsyncClient.post")
    async def test_decide_native_tool_calls(self, mock_post: MagicMock) -> None:
        """Native Ollama tool_calls are returned as a tool_calls decision."""
        mock_response = MagicMock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = {
            "message": {
                "role": "assistant",
                "content": "",
                "tool_calls": [
                    {
                        "function": {
                            "name": "get_feed",
                            "arguments": '{"limit": 10}',
                        }
                    }
                ],
            }
        }
        mock_post.return_value = mock_response

        decision = await self.client.decide(
            system="sys", user="user", tools=[{"type": "function", "function": {"name": "get_feed"}}]
        )
        self.assertIn("tool_calls", decision)
        self.assertEqual(decision["tool_calls"][0]["name"], "get_feed")
        self.assertEqual(decision["tool_calls"][0]["arguments"], {"limit": 10})

    @patch("httpx.AsyncClient.post")
    async def test_decide_json_tool_call_fallback(self, mock_post: MagicMock) -> None:
        """JSON-mode fallback: content with a tool_call object is honored."""
        mock_response = MagicMock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = {
            "message": {"role": "assistant", "content": '{"tool_call": {"name": "get_thread", "arguments": {"root_id": 5}}}'}
        }
        mock_post.return_value = mock_response

        decision = await self.client.decide(system="sys", user="user")
        self.assertIn("tool_calls", decision)
        self.assertEqual(decision["tool_calls"][0]["name"], "get_thread")
        self.assertEqual(decision["tool_calls"][0]["arguments"], {"root_id": 5})

    @patch("httpx.AsyncClient.post")
    async def test_decide_tool_whitelist(self, mock_post: MagicMock) -> None:
        """Disallowed tool names are dropped by the sanitizer."""
        mock_response = MagicMock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = {
            "message": {
                "role": "assistant",
                "content": "",
                "tool_calls": [{"function": {"name": "drop_database", "arguments": "{}"}}],
            }
        }
        mock_post.return_value = mock_response

        decision = await self.client.decide(system="sys", user="user")
        self.assertEqual(
            decision,
            {"action": "ignore", "text": ""},
        )

    @patch("httpx.AsyncClient.post")
    async def test_decide_rebuilds_transcript(self, mock_post: MagicMock) -> None:
        """Previous tool calls + results are replayed into the message history."""
        mock_response = MagicMock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = {
            "message": {"role": "assistant", "content": '{"action": "reply", "text": "ok"}'}
        }
        mock_post.return_value = mock_response

        calls = [{"name": "get_feed", "arguments": {"limit": 5}}]
        results = [{"name": "get_feed", "output": '[{"id":1}]', "error": False, "acted": False}]
        await self.client.decide(
            system="sys",
            user="user",
            tools=[{"type": "function", "function": {"name": "get_feed"}}],
            tool_calls=calls,
            tool_results=results,
        )

        request_body = mock_post.call_args.kwargs["json"]
        roles = [m["role"] for m in request_body["messages"]]
        self.assertEqual(roles, ["system", "user", "assistant", "tool"])
        self.assertEqual(request_body["messages"][2]["tool_calls"][0]["function"]["name"], "get_feed")
        self.assertEqual(request_body["messages"][2]["tool_calls"][0]["function"]["arguments"], {"limit": 5})

    @patch("httpx.AsyncClient.post")
    async def test_decide_rebuilds_failed_tool_as_error(self, mock_post: MagicMock) -> None:
        """Failed tool results are surfaced to the model as errors."""
        mock_response = MagicMock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = {
            "message": {"role": "assistant", "content": '{"action": "reply", "text": "ok"}'}
        }
        mock_post.return_value = mock_response

        calls = [{"name": "get_feed", "arguments": {}}]
        results = [{"name": "get_feed", "output": "boom", "error": True, "acted": False}]
        await self.client.decide(
            system="sys",
            user="user",
            tool_calls=calls,
            tool_results=results,
        )

        request_body = mock_post.call_args.kwargs["json"]
        self.assertEqual(request_body["messages"][3]["content"], "error: boom")

    def test_allowed_tools_match_build_tools(self) -> None:
        """The whitelist must cover every tool descriptor advertised by prompts."""
        from internal.prompts import prompts

        names = {t["function"]["name"] for t in prompts.build_tools()}
        self.assertEqual(names, set(ALLOWED_TOOLS))


if __name__ == "__main__":
    unittest.main()
