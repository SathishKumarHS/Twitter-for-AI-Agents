from pathlib import Path
import sys
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

# Ensure ai-worker root is in sys.path
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from gen.twitterforaiworker.v1 import worker_pb2
from internal.server.server import WorkerServicer


class TestWorkerServicer(unittest.IsolatedAsyncioTestCase):
    """Unit tests for gRPC WorkerServicer Think RPC method."""

    def setUp(self) -> None:
        self.mock_llm = AsyncMock()
        self.servicer = WorkerServicer(llm=self.mock_llm)

        self.agent = worker_pb2.Agent(
            id=1,
            handle="brand_ai",
            name="Brand AI",
            persona="Friendly assistant",
        )
        self.post = worker_pb2.Post(
            id=100,
            agent_id=2,
            body="Is anyone available?",
            author_handle="user1",
        )

    async def test_think_respond_mode(self) -> None:
        """Test Think RPC in respond mode with mocked LLM response."""
        self.mock_llm.decide.return_value = {
            "action": "reply",
            "text": "Yes, I am here!",
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            recent=[],
            mode="respond",
            memories=["learned: respond quickly"],
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertIsInstance(reply, worker_pb2.ThinkReply)
        self.assertEqual(reply.action, "reply")
        self.assertEqual(reply.text, "Yes, I am here!")
        self.mock_llm.decide.assert_called_once()

    async def test_think_proactive_mode(self) -> None:
        """Test Think RPC in proactive mode."""
        self.mock_llm.decide.return_value = {
            "action": "post",
            "text": "Exciting news today!",
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=worker_pb2.Post(),
            recent=[],
            mode="proactive",
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertEqual(reply.action, "post")
        self.assertEqual(reply.text, "Exciting news today!")

    @patch("logging.Logger.exception")
    async def test_think_exception_fallback(self, mock_log_exc: MagicMock) -> None:
        """Test Think RPC gracefully handles LLM exceptions by returning ignore fallback."""
        self.mock_llm.decide.side_effect = RuntimeError("Internal LLM Error")

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode="respond",
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertEqual(reply.action, "ignore")
        self.assertEqual(reply.text, "")


if __name__ == "__main__":
    unittest.main()
