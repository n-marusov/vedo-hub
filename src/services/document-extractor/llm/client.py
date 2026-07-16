"""LLM client — runtime-configured provider abstraction using httpx.

Supports OpenAI-compatible APIs (OpenAI, local proxies, etc.) and Anthropic.
Configuration via environment variables (see config.py).
"""

from __future__ import annotations

import json
import logging
import time
from typing import Any

import httpx

from config import settings

logger = logging.getLogger(settings.SERVICE_NAME)


class LlmClient:
    """LLM client for ontology extraction.

    Supports OpenAI-compatible and Anthropic API formats.
    Configuration loaded from environment via ``settings``.
    """

    def __init__(self) -> None:
        self._provider = settings.LLM_PROVIDER.lower()
        self._api_key = settings.LLM_API_KEY
        self._base_url = settings.LLM_BASE_URL.rstrip("/") if settings.LLM_BASE_URL else ""
        self._model = settings.LLM_MODEL
        self._max_retries = settings.LLM_MAX_RETRIES
        self._timeout = settings.LLM_TIMEOUT_SECONDS
        self._http_client: httpx.AsyncClient | None = None

    async def __aenter__(self) -> LlmClient:
        self._http_client = httpx.AsyncClient(timeout=self._timeout)
        return self

    async def __aexit__(self, *args: Any) -> None:
        if self._http_client:
            await self._http_client.aclose()
            self._http_client = None

    async def extract(
        self,
        messages: list[dict],
        temperature: float = 0.3,
        max_tokens: int = 4096,
    ) -> str:
        """Send messages to the LLM and return the response text.

        Args:
            messages: List of message dicts (role/content).
            temperature: Sampling temperature (lower = more deterministic).
            max_tokens: Maximum tokens in the response.

        Returns:
            The response content text.

        Raises:
            RuntimeError: If API key is not configured.
            httpx.HTTPError: On HTTP transport failures.
        """
        if not self._api_key and self._provider != "ollama":
            raise RuntimeError(f"LLM_API_KEY not configured for provider '{self._provider}'")

        if self._http_client is None:
            self._http_client = httpx.AsyncClient(timeout=self._timeout)

        if self._provider == "anthropic":
            return await self._call_anthropic(messages, temperature, max_tokens)
        return await self._call_openai(messages, temperature, max_tokens)

    async def extract_with_retry(
        self,
        messages: list[dict],
        temperature: float = 0.3,
        max_tokens: int = 4096,
    ) -> tuple[str, int]:
        """Call LLM with retry logic.

        Retries up to ``max_retries`` times with exponential backoff.

        Returns:
            Tuple of (response_text, attempts_made).
        """
        last_error: Exception | None = None

        for attempt in range(1, self._max_retries + 1):
            try:
                response = await self.extract(messages, temperature, max_tokens)
                if attempt > 1:
                    logger.info("LLM call succeeded on retry %d", attempt)
                return response, attempt
            except (httpx.HTTPError, json.JSONDecodeError, RuntimeError) as exc:
                last_error = exc
                if attempt < self._max_retries:
                    delay = 2**attempt  # exponential backoff
                    logger.warning(
                        "LLM call attempt %d/%d failed: %s. Retrying in %ds...",
                        attempt,
                        self._max_retries,
                        exc,
                        delay,
                    )
                    time.sleep(delay)
                else:
                    logger.error(
                        "LLM call failed after %d attempts: %s",
                        self._max_retries,
                        exc,
                    )

        raise RuntimeError(
            f"LLM call failed after {self._max_retries} attempts. Last error: {last_error}"
        ) from last_error

    # ─── Provider Implementations ───────────────────────────────────────────

    async def _call_openai(
        self,
        messages: list[dict],
        temperature: float,
        max_tokens: int,
    ) -> str:
        """Call an OpenAI-compatible API."""
        url = (
            f"{self._base_url}/chat/completions"
            if self._base_url
            else "https://api.openai.com/v1/chat/completions"
        )

        payload = {
            "model": self._model,
            "messages": messages,
            "temperature": temperature,
            "max_tokens": max_tokens,
        }

        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self._api_key}",
        }

        logger.debug(
            "LLM request: provider=openai model=%s messages=%d",
            self._model,
            len(messages),
        )

        response = await self._http_client.post(url, json=payload, headers=headers)
        response.raise_for_status()
        data = response.json()

        content = data["choices"][0]["message"]["content"]
        self._log_response(content, data)

        return content

    async def _call_anthropic(
        self,
        messages: list[dict],
        temperature: float,
        max_tokens: int,
    ) -> str:
        """Call the Anthropic API."""
        url = self._base_url or "https://api.anthropic.com/v1/messages"

        # Anthropic uses system separate from messages
        system_msg = ""
        clean_messages = []
        for m in messages:
            if m["role"] == "system":
                system_msg += m["content"] + "\n"
            else:
                clean_messages.append(m)

        payload = {
            "model": self._model,
            "messages": clean_messages,
            "max_tokens": max_tokens,
            "temperature": temperature,
        }
        if system_msg:
            payload["system"] = system_msg.strip()

        headers = {
            "Content-Type": "application/json",
            "x-api-key": self._api_key,
            "anthropic-version": "2023-06-01",
        }

        logger.debug(
            "LLM request: provider=anthropic model=%s messages=%d",
            self._model,
            len(clean_messages),
        )

        response = await self._http_client.post(url, json=payload, headers=headers)
        response.raise_for_status()
        data = response.json()

        content = data["content"][0]["text"]
        self._log_response(content, data)

        return content

    def _log_response(self, content: str, raw_data: dict) -> None:
        """Log LLM response metadata (truncate content if >1KB)."""
        usage = raw_data.get("usage", {})
        log_content = content[:1024] + "..." if len(content) > 1024 else content
        logger.debug(
            "LLM response: chars=%d tokens_in=%d tokens_out=%d",
            len(content),
            usage.get("prompt_tokens", 0),
            usage.get("completion_tokens", 0),
        )
