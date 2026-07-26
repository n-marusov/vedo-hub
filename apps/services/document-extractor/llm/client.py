"""LLM client — runtime-configured provider abstraction using httpx.

Supports OpenAI-compatible APIs (OpenAI, local proxies, etc.) and Anthropic.
Configuration via environment variables (see config.py).

Integrates with ai-orchestration-service for LLM policy checks (CheckPolicy)
and centralized usage auditing (LogLLMUsage) per ADR Hybrid Model.
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
        ontology_id: str = "",
        trace_id: str = "",
        user_id: str = "",
    ) -> tuple[str, int]:
        """Call LLM with retry logic and ai-orchestration integration.

        Before the LLM call, performs a CheckPolicy gRPC call to verify
        that LLM access is permitted for the given ontology. After a
        successful call, logs usage to the central audit system.

        Retries up to ``max_retries`` times with exponential backoff.

        Args:
            messages: List of message dicts (role/content).
            temperature: Sampling temperature.
            max_tokens: Maximum tokens in the response.
            ontology_id: Target ontology ID for policy check and usage logging.
            trace_id: OpenTelemetry trace ID for correlation.
            user_id: Authenticated user ID.

        Returns:
            Tuple of (response_text, attempts_made).

        Raises:
            RuntimeError: On permanent failure after exhausting retries.
            PermissionError: If CheckPolicy blocks the action.
        """
        # Check LLM policy before extraction if ontology_id is provided
        if ontology_id:
            await self._check_policy(ontology_id, "document_extraction", user_id, trace_id)

        last_error: Exception | None = None
        llm_start = time.time()

        for attempt in range(1, self._max_retries + 1):
            try:
                response = await self.extract(messages, temperature, max_tokens)
                if attempt > 1:
                    logger.info("LLM call succeeded on retry %d", attempt)

                # Log LLM usage after successful extraction
                if ontology_id:
                    duration_ms = int((time.time() - llm_start) * 1000)
                    await self._log_usage(
                        ontology_id=ontology_id,
                        action="document_extraction",
                        tokens_in=0,  # Estimated from messages if needed
                        tokens_out=0,
                        trace_id=trace_id,
                        user_id=user_id,
                        duration_ms=duration_ms,
                    )

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
                    logger.exception(
                        "LLM call failed after %d attempts: %s",
                        self._max_retries,
                        exc,
                    )

        raise RuntimeError(
            f"LLM call failed after {self._max_retries} attempts. Last error: {last_error}"
        ) from last_error

    # ─── Policy Check & Usage Logging ─────────────────────────────────────────

    async def _check_policy(
        self,
        ontology_id: str,
        action: str,
        user_id: str = "",
        trace_id: str = "",
    ) -> None:
        """Call CheckPolicy and raise PermissionError if blocked."""
        try:
            from grpc_client.ai_orchestration import get_ai_orch_client

            client = await get_ai_orch_client()
            result = await client.check_policy(
                ontology_id=ontology_id,
                action=action,
                user_id=user_id,
                trace_id=trace_id,
            )

            if not result.get("allowed", True):
                reason = result.get("reason", "Policy check failed")
                logger.error(
                    "LLM extraction blocked by policy: ontology=%s reason=%s",
                    ontology_id,
                    reason,
                )
                raise PermissionError(f"LLM access blocked: {reason}")  # noqa: TRY301

            logger.debug(
                "CheckPolicy passed: ontology=%s action=%s",
                ontology_id,
                action,
            )

        except ImportError:
            logger.debug("CheckPolicy skipped: ai-orchestration gRPC client not available")
        except PermissionError:
            raise
        except Exception as exc:
            # Fail-open during migration
            logger.warning(
                "CheckPolicy error (fail-open): %s",
                exc,
            )

    async def _log_usage(
        self,
        ontology_id: str = "",
        action: str = "document_extraction",
        tokens_in: int = 0,
        tokens_out: int = 0,
        trace_id: str = "",
        user_id: str = "",
        duration_ms: int = 0,
    ) -> None:
        """Log LLM usage to the central audit system (fire-and-forget)."""
        try:
            from grpc_client.ai_orchestration import get_ai_orch_client

            client = await get_ai_orch_client()
            await client.log_usage(
                ontology_id=ontology_id,
                action=action,
                provider=self._provider,
                model=self._model,
                tokens_in=tokens_in,
                tokens_out=tokens_out,
                trace_id=trace_id,
                user_id=user_id,
                duration_ms=duration_ms,
            )
        except Exception as exc:
            # Fire-and-forget: errors are logged but not propagated
            logger.warning(
                "LogLLMUsage failed (fire-and-forget): %s",
                exc,
            )

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
        content[:1024] + "..." if len(content) > 1024 else content
        logger.debug(
            "LLM response: chars=%d tokens_in=%d tokens_out=%d",
            len(content),
            usage.get("prompt_tokens", 0),
            usage.get("completion_tokens", 0),
        )
