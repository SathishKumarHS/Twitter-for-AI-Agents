from __future__ import annotations

import os
from pathlib import Path

import yaml


# Internal helper: Loads and parses the active config.yaml file based on ACTIVE_ENV
def _load_file() -> dict:
    env = os.getenv("ACTIVE_ENV", "local")
    path = Path(__file__).resolve().parents[2] / "config" / env / "config.yaml"
    with open(path, "r", encoding="utf-8") as fh:
        data = yaml.safe_load(fh)
    if not isinstance(data, dict):
        raise RuntimeError("config.yaml: missing top-level mapping")
    return data


# Global configuration dictionary cached on module load
_CFG = _load_file()


# Internal helper: Retrieves a required configuration key path or raises RuntimeError
def _require(path: list[str]):
    cur = _CFG
    keys = []
    for key in path:
        keys.append(key)
        if not isinstance(cur, dict) or key not in cur:
            raise RuntimeError(f"config.yaml: missing value at '{'.'.join(keys)}'")
        cur = cur[key]
    return cur


# Returns the configured Ollama base URL
def ollama_url() -> str:
    return str(_require(["ollama", "url"]))


# Returns the configured Ollama model name
def ollama_model() -> str:
    return str(_require(["ollama", "model"]))


# Returns the timeout in seconds for Ollama HTTP API calls
def ollama_timeout() -> float:
    return float(_require(["ollama", "timeout"]))


# Returns the maximum number of retry attempts for Ollama calls
def ollama_retries() -> int:
    return int(_require(["ollama", "retries"]))


# Returns the backoff multiplier in seconds between Ollama retry attempts
def ollama_backoff() -> float:
    return float(_require(["ollama", "backoff"]))


# Returns the gRPC server listen address for the AI worker
def listen_addr() -> str:
    return str(_require(["server", "address"]))


# Returns the Go backend gRPC tool-executor server address
def backend_tool_address() -> str:
    return str(_require(["backend", "tool-address"]))


# Returns the maximum number of tool calls permitted per turn
def max_tool_calls_per_turn() -> int:
    return int(_require(["prompts", "max-tool-calls-per-turn"]))


# Returns the maximum character limit for posts and replies (280 chars)
def max_character_limit() -> int:
    return int(_require(["prompts", "max-character-limit"]))


# Returns the maximum number of replies before a thread is considered full
def max_thread_replies() -> int:
    return int(_require(["prompts", "max-thread-replies"]))


# Returns the default number of posts returned by get_feed
def feed_default_limit() -> int:
    return int(_require(["prompts", "feed-default-limit"]))


# Returns the maximum cap on posts returned by get_feed
def feed_max_limit() -> int:
    return int(_require(["prompts", "feed-max-limit"]))
