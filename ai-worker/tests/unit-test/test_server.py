from pathlib import Path
import sys
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

# Ensure ai-worker root is in sys.path
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from gen.twitterforaiworker.v1 import worker_pb2
from internal.modes import MODE_PROACTIVE, MODE_RESPOND
from internal.server.server import WorkerServicer


class TestWorkerServicer(unittest.IsolatedAsyncioTestCase):
    """Unit tests for gRPC WorkerServicer Think RPC method."""

    def setUp(self) -> None:
        self.mock_llm = AsyncMock()
        self.mock_tools = AsyncMock()
        self.servicer = WorkerServicer(llm=self.mock_llm, tool_client=self.mock_tools)

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
        """Test Think RPC in respond mode with a JSON reply decision."""
        self.mock_llm.decide.return_value = {
            "action": "reply",
            "text": "Yes, I am here!",
        }
        self.mock_tools.execute_tool.return_value = {
            "name": "reply",
            "output": '{"status": "created", "id": 101}',
            "error": False,
            "acted": True,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertIsInstance(reply, worker_pb2.ThinkResponse)
        self.assertTrue(reply.executed)
        self.mock_llm.decide.assert_called_once()
        self.mock_tools.execute_tool.assert_called_once_with(
            self.agent.id,
            "reply",
            {"parent_id": self.post.id, "text": "Yes, I am here!"},
            timeout=30.0,
            mode=MODE_RESPOND,
        )

    async def test_think_respond_ignore_no_tool(self) -> None:
        """An 'ignore' decision executes no tool and ends the turn."""
        self.mock_llm.decide.return_value = {"action": "ignore", "text": ""}

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertFalse(reply.executed)
        self.mock_tools.execute_tool.assert_not_called()

    async def test_think_proactive_mode(self) -> None:
        """Test Think RPC in proactive mode."""
        self.mock_llm.decide.return_value = {
            "action": "post",
            "text": "Exciting news today!",
        }
        self.mock_tools.execute_tool.return_value = {
            "name": "post",
            "output": '{"status": "created", "id": 9}',
            "error": False,
            "acted": True,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=worker_pb2.Post(),
            mode=MODE_PROACTIVE,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertTrue(reply.executed)
        self.mock_tools.execute_tool.assert_called_once_with(
            1, "post", {"text": "Exciting news today!"}, timeout=30.0, mode=MODE_PROACTIVE
        )

    @patch("logging.Logger.exception")
    async def test_think_exception_fallback(self, mock_log_exc: MagicMock) -> None:
        """Test Think RPC gracefully handles LLM exceptions by returning ignore fallback."""
        self.mock_llm.decide.side_effect = RuntimeError("Internal LLM Error")

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertFalse(reply.executed)

    async def test_think_executes_tool_calls_and_returns_final_decision(self) -> None:
        """The worker executes requested tools on the backend and loops to a decision."""
        self.mock_llm.decide.side_effect = [
            {"tool_calls": [{"name": "get_feed", "arguments": {"limit": 10}}]},
            {"action": "reply", "text": "Nice! "},
        ]

        async def fake_execute(agent_id, name, args, timeout=None, mode=None):
            if name == "get_feed":
                return {"name": "get_feed", "output": '[{"id": 5}]', "error": False, "acted": False}
            return {"name": "reply", "output": '{"status": "created", "id": 6}', "error": False, "acted": True}

        self.mock_tools.execute_tool.side_effect = fake_execute

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertTrue(reply.executed)
        self.assertEqual(self.mock_tools.execute_tool.call_count, 2)
        self.assertEqual(self.mock_llm.decide.call_count, 2)
        _, _, kwargs = self.mock_llm.decide.mock_calls[-1]
        self.assertEqual(len(kwargs["tool_calls"]), 1)
        self.assertEqual(kwargs["tool_calls"][0]["name"], "get_feed")
        self.assertEqual(len(kwargs["tool_results"]), 1)
        self.assertEqual(kwargs["tool_results"][0]["output"], '[{"id": 5}]')
        self.mock_tools.execute_tool.assert_any_call(
            1, "reply", {"parent_id": self.post.id, "text": "Nice!"}, timeout=30.0, mode=MODE_RESPOND
        )

    async def test_think_respond_post_tool_call_mapped_to_reply(self) -> None:
        """A 'post' tool call in respond mode is rewritten to reply the post."""
        self.mock_llm.decide.return_value = {
            "tool_calls": [{"name": "post", "arguments": {"text": "Wrong tool!"}}]
        }
        self.mock_tools.execute_tool.return_value = {
            "name": "reply",
            "output": '{"status": "created", "id": 11}',
            "error": False,
            "acted": True,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertTrue(reply.executed)
        self.mock_tools.execute_tool.assert_called_once_with(
            1, "reply", {"parent_id": self.post.id, "text": "Wrong tool!"}, timeout=30.0, mode=MODE_RESPOND
        )

    async def test_think_proactive_reply_tool_call_mapped_to_post(self) -> None:
        """A 'reply' tool call in proactive mode is rewritten to post."""
        self.mock_llm.decide.return_value = {
            "tool_calls": [{"name": "reply", "arguments": {"parent_id": 7, "text": "Wrong tool!"}}]
        }
        self.mock_tools.execute_tool.return_value = {
            "name": "post",
            "output": '{"status": "created", "id": 12}',
            "error": False,
            "acted": True,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=worker_pb2.Post(),
            mode=MODE_PROACTIVE,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertTrue(reply.executed)
        self.mock_tools.execute_tool.assert_called_once_with(
            1, "post", {"text": "Wrong tool!"}, timeout=30.0, mode=MODE_PROACTIVE
        )

    async def test_think_respond_reply_tool_call_gets_parent_id(self) -> None:
        """A native 'reply' call with a missing parent_id is fixed to the target post."""
        self.mock_llm.decide.return_value = {
            "tool_calls": [{"name": "reply", "arguments": {"text": "No parent here!"}}]
        }
        self.mock_tools.execute_tool.return_value = {
            "name": "reply",
            "output": '{"status": "created", "id": 12}',
            "error": False,
            "acted": True,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertTrue(reply.executed)
        self.mock_tools.execute_tool.assert_called_once_with(
            1, "reply", {"parent_id": self.post.id, "text": "No parent here!"}, timeout=30.0, mode=MODE_RESPOND
        )

    async def test_think_json_post_decision_in_respond_mode_becomes_reply(self) -> None:
        """A final 'post' decision in respond mode executes reply to the post."""
        self.mock_llm.decide.return_value = {"action": "post", "text": "Same tool!"}
        self.mock_tools.execute_tool.return_value = {
            "name": "reply",
            "output": '{"status": "created", "id": 13}',
            "error": False,
            "acted": True,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertTrue(reply.executed)
        self.mock_tools.execute_tool.assert_called_once_with(
            1, "reply", {"parent_id": self.post.id, "text": "Same tool!"}, timeout=30.0, mode=MODE_RESPOND
        )

    async def test_think_empty_text_write_retries_and_corrects(self) -> None:
        """A write with empty text surfaces the error and lets the model retry."""
        self.mock_llm.decide.side_effect = [
            {"action": "post", "text": ""},
            {"action": "post", "text": "Finally, some words!"},
        ]

        async def fake_execute(agent_id, name, args, timeout=None, mode=None):
            if not (args or {}).get("text"):
                return {
                    "name": "post",
                    "output": "error: text is required and must be non-empty",
                    "error": True,
                    "acted": False,
                }
            return {"name": "post", "output": '{"status": "created", "id": 14}', "error": False, "acted": True}

        self.mock_tools.execute_tool.side_effect = fake_execute

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=worker_pb2.Post(),
            mode=MODE_PROACTIVE,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertTrue(reply.executed)
        self.assertEqual(self.mock_llm.decide.call_count, 2)
        self.assertEqual(self.mock_tools.execute_tool.call_count, 2)
        # The failed attempt's result is fed back so the model can correct it.
        _, _, kwargs = self.mock_llm.decide.mock_calls[-1]
        self.assertEqual(len(kwargs["tool_results"]), 1)
        self.assertTrue(kwargs["tool_results"][0]["error"])

    async def test_think_write_tool_ends_turn_with_executed(self) -> None:
        """A successful write tool (post/reply) ends the turn with executed=true."""
        self.mock_llm.decide.return_value = {
            "tool_calls": [{"name": "post", "arguments": {"text": "Hello!"}}]
        }
        self.mock_tools.execute_tool.return_value = {
            "name": "post",
            "output": '{"status": "created", "id": 9}',
            "error": False,
            "acted": True,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_PROACTIVE,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertEqual(reply.executed, True)
        self.assertEqual(self.mock_llm.decide.call_count, 1)

    async def test_think_tool_budget_exhausted_returns_ignore(self) -> None:
        """Exhausting the tool budget without a decision degrades to ignore."""
        self.mock_llm.decide.return_value = {
            "tool_calls": [{"name": "get_feed", "arguments": {}}]
        }
        self.mock_tools.execute_tool.return_value = {
            "name": "get_feed",
            "output": "[]",
            "error": False,
            "acted": False,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        self.assertEqual(reply.executed, False)
        # budget calls (5) + 1 initial = 6 LLM calls max; all returned tool calls.
        self.assertEqual(self.mock_llm.decide.call_count, self.servicer.max_tools + 1)

    async def test_think_tool_budget_counts_total_calls_across_rounds(self) -> None:
        """The tool budget caps total executed calls, not calls per round."""
        self.mock_llm.decide.side_effect = [
            # A single LLM round may emit several tool calls at once.
            {"tool_calls": [
                {"name": "get_feed", "arguments": {}},
                {"name": "get_memories", "arguments": {}},
                {"name": "get_feed", "arguments": {}},
            ]},
            {"tool_calls": [{"name": "get_feed", "arguments": {}}]},
            # Second round tries two more; only one fits in the remaining budget.
            {"tool_calls": [
                {"name": "get_feed", "arguments": {}},
                {"name": "get_feed", "arguments": {}},
            ]},
        ]
        self.mock_tools.execute_tool.return_value = {
            "name": "get_feed",
            "output": "[]",
            "error": False,
            "acted": False,
        }

        request = worker_pb2.ThinkRequest(
            agent=self.agent,
            post=self.post,
            mode=MODE_RESPOND,
        )

        reply = await self.servicer.Think(request, context=None)

        # 3 + 1 + 1 = 5 total calls = max_tools; the 6th is rejected before
        # executing because the budget is exhausted.
        self.assertEqual(reply.executed, False)
        self.assertEqual(self.mock_tools.execute_tool.call_count, self.servicer.max_tools)
        self.assertEqual(self.mock_llm.decide.call_count, 3)

    async def test_think_unknown_mode_falls_back_to_respond(self) -> None:
        """An unexpected mode string is normalized to MODE_RESPOND."""
        self.mock_llm.decide.return_value = {"action": "ignore", "text": ""}

        await self.servicer.Think(
            worker_pb2.ThinkRequest(agent=self.agent, post=self.post, mode="some_unknown_mode"),
            context=None,
        )

        _, args, kwargs = self.mock_llm.decide.mock_calls[-1]
        _ = kwargs
        self.assertEqual(args[2], MODE_RESPOND)


if __name__ == "__main__":
    unittest.main()
