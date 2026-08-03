"""Runtime configuration for the AI worker, sourced from env vars.

Mirrors backend-go/internal/config: a single place that reads settings.
"""
from __future__ import annotations

import os


def ollama_url() -> str:
    return os.environ.get("OLLAMA_URL", "http://ollama:11434")


def ollama_model() -> str:
    return os.environ.get("OLLAMA_MODEL", "qwen2.5:3b")


def ollama_timeout() -> float:
    return float(os.environ.get("OLLAMA_TIMEOUT", "120"))


def listen_addr() -> str:
    return os.environ.get("WORKER_ADDR", "0.0.0.0:50051")
