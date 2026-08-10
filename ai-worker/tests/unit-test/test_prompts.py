from pathlib import Path
import sys
import unittest

# Ensure ai-worker root is in sys.path
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from gen.twitterforaiworker.v1 import worker_pb2
from internal.config import config
from internal.modes import MODE_PROACTIVE, MODE_RESPOND
from internal.prompts import prompts


class TestPrompts(unittest.TestCase):
    """Unit tests for system and user prompt generation in prompts.py."""

    def setUp(self) -> None:
        self.agent = worker_pb2.Agent(
            id=1,
            handle="test_bot",
            name="Test Bot",
            persona="A test AI agent",
            goals=["Engage users", "Share tech tips"],
            policies=["No spam", "Be polite"],
            interests=["AI", "Go", "Python"],
        )

        self.post = worker_pb2.Post(
            id=10,
            agent_id=2,
            body="Hello world post!",
            kind="post",
            author_handle="other_agent",
            author_name="Other Agent",
        )

    def test_build_system(self) -> None:
        """Test build_system formats agent metadata and instructions."""
        sys_prompt = prompts.build_system(self.agent)
        self.assertIn("Test Bot (@test_bot)", sys_prompt)
        self.assertIn("Persona: A test AI agent", sys_prompt)
        self.assertIn("Goals: Engage users, Share tech tips", sys_prompt)
        self.assertIn("Policies: No spam, Be polite", sys_prompt)
        self.assertIn("Interests: AI, Go, Python", sys_prompt)
    def test_build_user_proactive_mode(self) -> None:
        """Test build_user formats proactive new post prompt."""
        user_prompt = prompts.build_user(
            agent=self.agent, post=worker_pb2.Post(), mode=MODE_PROACTIVE
        )
        self.assertIn("authoring a NEW main post for @test_bot", user_prompt)
        self.assertIn("get_feed", user_prompt)
        self.assertIn("Start a new thread!", user_prompt)

    def test_build_user_respond_mode(self) -> None:
        """Test build_user formats respond prompt and points at the tools."""
        user_prompt = prompts.build_user(
            agent=self.agent, post=self.post, mode=MODE_RESPOND
        )
        self.assertIn("Decide how @test_bot should react to this post from @other_agent:", user_prompt)
        self.assertIn("#10 (@other_agent): Hello world post!", user_prompt)
        self.assertIn("get_thread", user_prompt)
        self.assertIn("get_memories", user_prompt)

    def test_build_tools_shape(self) -> None:
        """build_tools returns Ollama-schema descriptors with whitelisted names."""
        tools = prompts.build_tools()
        names = {t["function"]["name"] for t in tools}
        self.assertEqual(names, {"get_feed", "get_thread", "get_memories", "post", "reply"})
        for t in tools:
            self.assertEqual(t["type"], "function")
            self.assertIn("parameters", t["function"])

    def test_build_tools_section(self) -> None:
        """build_tools_section describes the tools in prose for JSON-mode fallback."""
        section = prompts.build_tools_section()
        self.assertIn("get_feed", section)
        self.assertIn("post(text)", section)
        self.assertIn("reply(parent_id, text)", section)

    def test_build_tools_feed_limits_from_config(self) -> None:
        """The get_feed tool advertises the feed limits from config, not literals."""
        tools = prompts.build_tools()
        feed = next(t for t in tools if t["function"]["name"] == "get_feed")
        desc = feed["function"]["parameters"]["properties"]["limit"]["description"]
        self.assertEqual(
            desc,
            f"Number of posts to return (default {config.feed_default_limit()}, "
            f"max {config.feed_max_limit()}).",
        )

    def test_build_system_tool_budget_from_config(self) -> None:
        """The tool budget is advertised to the agent from config, not a literal."""
        sys_prompt = prompts.build_system(self.agent)
        self.assertIn(f"at most {config.max_tool_calls_per_turn()} tools this turn", sys_prompt)


if __name__ == "__main__":
    unittest.main()
