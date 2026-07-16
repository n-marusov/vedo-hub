"""Tests for the LLM client (mock HTTP responses)."""

from __future__ import annotations

import pytest

from llm.client import LlmClient


@pytest.mark.asyncio
async def test_client_initialization() -> None:
    """Test that LlmClient initializes with settings defaults."""
    async with LlmClient() as client:
        assert client._provider in ("openai", "anthropic", "ollama")
        assert client._model is not None
        assert client._max_retries > 0
        assert client._timeout > 0


@pytest.mark.asyncio
async def test_client_fails_without_api_key() -> None:
    """Test that client raises RuntimeError when API key is missing (non-Ollama)."""
    async with LlmClient() as client:
        client._api_key = ""
        client._provider = "openai"
        with pytest.raises(RuntimeError):
            await client.extract(messages=[{"role": "user", "content": "test"}])


@pytest.mark.asyncio
async def test_extract_with_retry_handles_missing_api_key() -> None:
    """Test that extract_with_retry fails gracefully."""
    async with LlmClient() as client:
        client._api_key = ""
        client._provider = "openai"
        client._max_retries = 2
        with pytest.raises(RuntimeError, match="LLM call failed after"):
            await client.extract_with_retry(messages=[{"role": "user", "content": "test"}])
