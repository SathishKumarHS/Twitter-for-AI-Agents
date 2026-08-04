"""Runtime configuration for the AI worker.

Mirrors backend-go/internal/config: settings live in a YAML config file
(ai-worker/config/local/config.yaml). No environment-variable overrides and
no hardcoded defaults — the config file is the single source of truth.
"""
from __future__ import annotations

from pathlib import Path

import yaml


def _load_file() -> dict:
    path = Path(__file__).resolve().parents[2] / "config" / "local" / "config.yaml"
    with open(path, "r", encoding="utf-8") as fh:
        data = yaml.safe_load(fh)
    if not isinstance(data, dict):
        raise RuntimeError("config.yaml: missing top-level mapping")
    return data


_CFG = _load_file()


def _require(path: list[str]):
    cur = _CFG
    keys = []
    for key in path:
        keys.append(key)
        if not isinstance(cur, dict) or key not in cur:
            raise RuntimeError(f"config.yaml: missing value at '{'.'.join(keys)}'")
        cur = cur[key]
    return cur


def ollama_url() -> str:
    return str(_require(["ollama", "url"]))


def ollama_model() -> str:
    return str(_require(["ollama", "model"]))


def ollama_timeout() -> float:
    return float(_require(["ollama", "timeout"]))


def ollama_retries() -> int:
    return int(_require(["ollama", "retries"]))


def ollama_backoff() -> float:
    return float(_require(["ollama", "backoff"]))


def listen_addr() -> str:
    return str(_require(["server", "address"]))


def max_character_limit() -> int:
    return int(_require(["prompts", "max-character-limit"]))


def max_thread_replies() -> int:
    return int(_require(["prompts", "max-thread-replies"]))
