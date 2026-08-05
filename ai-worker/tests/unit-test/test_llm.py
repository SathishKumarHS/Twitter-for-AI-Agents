import asyncio
from pathlib import Path
import sys
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

# Ensure ai-worker root is in sys.path
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from internal.llm.llm import LLMClient


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
        self.assertEqual(decision, {"action": "reply", "text": "Great post!"})

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
        self.assertEqual(decision, {"action": "ignore", "text": ""})
        self.assertEqual(mock_post.call_count, 2)


if __name__ == "__main__":
    unittest.main()
