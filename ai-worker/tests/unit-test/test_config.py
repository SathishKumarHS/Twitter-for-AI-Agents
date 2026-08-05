import sys
from pathlib import Path
import unittest

# Ensure ai-worker root is in sys.path
_ROOT = Path(__file__).resolve().parents[2]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

from internal.config import config


class TestConfig(unittest.TestCase):
    """Unit tests for internal.config.config loading config.yaml."""

    def test_config_values_loaded(self) -> None:
        """Verify that all values from config.yaml are loaded correctly."""
        self.assertEqual(config.listen_addr(), "0.0.0.0:50051")
        self.assertEqual(config.ollama_url(), "http://ollama:11434")
        self.assertEqual(config.ollama_model(), "llama3.2:3b")
        self.assertEqual(config.ollama_timeout(), 120.0)
        self.assertEqual(config.ollama_retries(), 3)
        self.assertEqual(config.ollama_backoff(), 1.0)
        self.assertEqual(config.max_character_limit(), 280)
        self.assertEqual(config.max_thread_replies(), 4)

    def test_require_missing_key_raises_runtime_error(self) -> None:
        """Verify that requiring a non-existent key raises RuntimeError."""
        with self.assertRaises(RuntimeError) as ctx:
            config._require(["non_existent_section", "foo"])
        self.assertIn("config.yaml: missing value", str(ctx.exception))


if __name__ == "__main__":
    unittest.main()
