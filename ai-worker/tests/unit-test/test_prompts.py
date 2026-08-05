from pathlib import Path
import sys
import unittest

# Ensure ai-worker root is in sys.path
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from gen.twitterforaiworker.v1 import worker_pb2
from internal.prompts import prompts


class TestPrompts(unittest.TestCase):
    """Unit tests for system and user prompt generation in prompts.py."""

    def setUp(self) -> None:
        self.agent = worker_pb2.Agent(
            id=1,
            handle="test_bot",
            name="Test Bot",
            description="A test AI agent",
            instructions="Be witty and helpful",
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

    def test_build_system_without_memories(self) -> None:
        """Test build_system formats agent metadata and default memories."""
        sys_prompt = prompts.build_system(self.agent, memories=None)
        self.assertIn("Test Bot (@test_bot)", sys_prompt)
        self.assertIn("Description: A test AI agent", sys_prompt)
        self.assertIn("Goals: Engage users, Share tech tips", sys_prompt)
        self.assertIn("Policies (must follow): No spam, Be polite", sys_prompt)
        self.assertIn("Interests: AI, Go, Python", sys_prompt)
        self.assertIn("(none yet)", sys_prompt)

    def test_build_system_with_memories(self) -> None:
        """Test build_system formats provided agent memories."""
        memories = ["posted: Hello world", "replied to post #5"]
        sys_prompt = prompts.build_system(self.agent, memories=memories)
        self.assertIn("- posted: Hello world", sys_prompt)
        self.assertIn("- replied to post #5", sys_prompt)

    def test_build_user_proactive_mode(self) -> None:
        """Test build_user formats proactive new post prompt."""
        user_prompt = prompts.build_user(
            agent=self.agent, post=worker_pb2.Post(), recent=[], mode="proactive"
        )
        self.assertIn("authoring a NEW main post for @test_bot", user_prompt)
        self.assertIn("Start a new thread!", user_prompt)

    def test_build_user_respond_mode_with_recent_context(self) -> None:
        """Test build_user formats respond prompt with recent context."""
        recent = [
            worker_pb2.Post(
                id=1,
                agent_id=3,
                body="Original thread",
                kind="post",
                author_handle="alice",
                author_name="Alice",
            ),
            worker_pb2.Post(
                id=2,
                agent_id=2,
                body="A reply in thread",
                kind="reply",
                parent_id=1,
                author_handle="other_agent",
                author_name="Other Agent",
            ),
        ]

        user_prompt = prompts.build_user(
            agent=self.agent, post=self.post, recent=recent, mode="respond"
        )
        self.assertIn("Recent conversation context:", user_prompt)
        self.assertIn("#1 [POST] @alice (Alice): Original thread", user_prompt)
        self.assertIn("#2 [REPLY to #1] @other_agent (Other Agent): A reply in thread", user_prompt)
        self.assertIn("Decide how @test_bot should react to this post from @other_agent:", user_prompt)


if __name__ == "__main__":
    unittest.main()
