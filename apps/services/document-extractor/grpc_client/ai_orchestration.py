"""gRPC client for ai-orchestration-service CheckPolicy and LogLLMUsage RPCs.

Used by the document-extractor to enforce LLM access policies and log
LLM usage to the central audit system (per ADR Hybrid Model).
"""

from __future__ import annotations

import logging
from typing import Any

import grpc

from config import settings

logger = logging.getLogger(settings.SERVICE_NAME)


class AIOrchestrationGrpcClient:
    """gRPC client for ai-orchestration-service policy and audit operations.

    CheckPolicy is called BEFORE each LLM call to verify that the action
    is permitted for the given ontology and deployment mode.

    LogLLMUsage is called AFTER each LLM call to record usage metrics
    (fire-and-forget — failures are logged but not propagated).
    """

    def __init__(self, target: str | None = None) -> None:
        """Initialize the gRPC client.

        Args:
            target: gRPC target address (host:port).
                    Defaults to settings.AI_ORCHESTRATION_URL.
        """
        self._target = target or settings.AI_ORCHESTRATION_URL
        self._channel: grpc.aio.Channel | None = None
        self._stub: Any = None  # AIOrchestrationServiceStub — set in connect()

    async def connect(self) -> None:
        """Establish the gRPC channel and create the service stub."""
        logger.debug("Connecting ai-orchestration gRPC client to %s", self._target)
        self._channel = grpc.aio.insecure_channel(
            self._target,
            options=[
                ("grpc.keepalive_time_ms", 10000),
                ("grpc.keepalive_timeout_ms", 5000),
                ("grpc.keepalive_permit_without_calls", True),
            ],
        )

        # Lazy import generated stubs
        try:
            from grpc_client.proto.ai_orchestration.v1.ai_orchestration_pb2_grpc import (  # type: ignore[import-untyped]
                AIOrchestrationServiceStub,
            )

            self._stub = AIOrchestrationServiceStub(self._channel)
            logger.info("ai-orchestration gRPC client connected to %s", self._target)
        except ImportError:
            logger.warning(
                "ai-orchestration gRPC proto stubs not available; "
                "run proto-generate-ai-orchestration-python. "
                "Policy checks and usage logging will be skipped."
            )
            self._stub = None

    async def close(self) -> None:
        """Close the gRPC channel."""
        if self._channel:
            logger.debug("Closing ai-orchestration gRPC channel to %s", self._target)
            await self._channel.close()
            self._channel = None
            self._stub = None

    async def check_policy(
        self,
        ontology_id: str,
        action: str = "document_extraction",
        user_id: str = "",
        trace_id: str = "",
    ) -> dict:
        """Call CheckPolicy to verify LLM access is allowed.

        Args:
            ontology_id: Target ontology ID.
            action: Action identifier (e.g., "document_extraction").
            user_id: Authenticated user ID (from proxy headers).
            trace_id: OpenTelemetry trace ID for correlation.

        Returns:
            Dict with keys: allowed, provider, model, reason, require_consent.
            On failure (stub not available / RPC error), returns a default
            allow-with-warning so extraction is not blocked.
        """
        if self._stub is None:
            logger.warning(
                "CheckPolicy skipped: gRPC stub not available for %s/%s",
                ontology_id,
                action,
            )
            return {
                "allowed": True,
                "provider": "",
                "model": "",
                "reason": "Policy check skipped: gRPC stub not available",
                "require_consent": False,
            }

        try:
            from grpc_client.proto.ai_orchestration.v1.ai_orchestration_pb2 import (  # type: ignore[import-untyped]
                CheckPolicyRequest,
            )

            request = CheckPolicyRequest(
                ontology_id=ontology_id,
                action=action,
                user_id=user_id,
                trace_id=trace_id,
            )

            response = await self._stub.CheckPolicy(request)  # type: ignore[union-attr]

            result = {
                "allowed": response.allowed,
                "provider": response.provider,
                "model": response.model,
                "reason": response.reason,
                "require_consent": response.require_consent,
            }

            if result["allowed"]:
                logger.info(
                    "CheckPolicy allowed: ontology=%s action=%s reason=%s",
                    ontology_id,
                    action,
                    result["reason"],
                )
            else:
                logger.warning(
                    "CheckPolicy blocked: ontology=%s action=%s reason=%s",
                    ontology_id,
                    action,
                    result["reason"],
                )

            return result

        except ImportError:
            logger.warning(
                "CheckPolicy stubs not generated; skipping policy check for %s/%s",
                ontology_id,
                action,
            )
            return {
                "allowed": True,
                "provider": "",
                "model": "",
                "reason": "Policy check skipped: proto stubs not generated",
                "require_consent": False,
            }
        except grpc.RpcError as exc:
            logger.error(
                "CheckPolicy RPC failed for %s/%s: %s",
                ontology_id,
                action,
                exc.details() if hasattr(exc, "details") else str(exc),
            )
            # Fail-open during migration: allow extraction but log the issue
            return {
                "allowed": True,
                "provider": "",
                "model": "",
                "reason": f"Policy check failed: {exc}",
                "require_consent": False,
            }

    async def log_usage(
        self,
        ontology_id: str,
        action: str = "document_extraction",
        provider: str = "",
        model: str = "",
        tokens_in: int = 0,
        tokens_out: int = 0,
        cost: float = 0.0,
        trace_id: str = "",
        user_id: str = "",
        duration_ms: int = 0,
    ) -> None:
        """Call LogLLMUsage to record LLM usage for audit.

        Fire-and-forget: errors are logged but never propagated to the caller.

        Args:
            ontology_id: Target ontology ID.
            action: Action identifier.
            provider: LLM provider name.
            model: LLM model name.
            tokens_in: Input tokens consumed.
            tokens_out: Output tokens consumed.
            cost: Estimated cost in USD.
            trace_id: OpenTelemetry trace ID.
            user_id: Authenticated user ID.
            duration_ms: Request duration in milliseconds.
        """
        if self._stub is None:
            logger.debug("LogLLMUsage skipped: gRPC stub not available")
            return

        try:
            from grpc_client.proto.ai_orchestration.v1.ai_orchestration_pb2 import (  # type: ignore[import-untyped]
                LogLLMUsageRequest,
            )

            request = LogLLMUsageRequest(
                ontology_id=ontology_id,
                action=action,
                provider=provider,
                model=model,
                tokens_in=tokens_in,
                tokens_out=tokens_out,
                cost=cost,
                trace_id=trace_id,
                user_id=user_id,
                duration_ms=duration_ms,
            )

            await self._stub.LogLLMUsage(request)  # type: ignore[union-attr]

            logger.debug(
                "LogLLMUsage recorded: ontology=%s action=%s tokens_in=%d tokens_out=%d",
                ontology_id,
                action,
                tokens_in,
                tokens_out,
            )

        except ImportError:
            logger.debug("LogLLMUsage stubs not generated; skipping")
        except grpc.RpcError as exc:
            # Fire-and-forget: log the error but don't propagate
            logger.warning(
                "LogLLMUsage RPC failed (fire-and-forget): %s",
                exc.details() if hasattr(exc, "details") else str(exc),
            )


# Singleton instance
_client: AIOrchestrationGrpcClient | None = None


async def get_ai_orch_client() -> AIOrchestrationGrpcClient:
    """Get or create the singleton ai-orchestration gRPC client instance."""
    global _client
    if _client is None:
        _client = AIOrchestrationGrpcClient()
        await _client.connect()
    return _client


async def close_ai_orch_client() -> None:
    """Close the singleton ai-orchestration gRPC client if it exists."""
    global _client
    if _client:
        await _client.close()
        _client = None
