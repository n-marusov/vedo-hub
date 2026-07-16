"""gRPC client wrapper for the ontology-service ApplySequence RPC."""

from __future__ import annotations

import logging
from typing import Any

import grpc

from config import settings

logger = logging.getLogger(settings.SERVICE_NAME)


class OntologyGrpcClient:
    """gRPC client for ontology-service ApplySequence operations.

    Connects to the ontology-service gRPC endpoint and sends validated
    ontology sequences for atomic batch application.
    """

    def __init__(self, target: str | None = None) -> None:
        """Initialize the gRPC client.

        Args:
            target: gRPC target address (host:port).
                    Defaults to settings.ONTOLOGY_SERVICE_URL.
        """
        self._target = target or settings.ONTOLOGY_SERVICE_URL
        self._channel: grpc.aio.Channel | None = None
        self._stub: Any = None  # OntologyServiceStub — set in connect()

    async def connect(self) -> None:
        """Establish the gRPC channel and create the service stub."""
        logger.debug("Connecting gRPC client to %s", self._target)
        self._channel = grpc.aio.insecure_channel(
            self._target,
            options=[
                ("grpc.keepalive_time_ms", 10000),
                ("grpc.keepalive_timeout_ms", 5000),
                ("grpc.keepalive_permit_without_calls", True),
            ],
        )

        # Lazy import generated stubs to avoid import-order issues
        try:
            from grpc_client.proto.ontology.v1.ontology_pb2_grpc import (  # type: ignore[import-untyped]
                OntologyServiceStub,
            )

            self._stub = OntologyServiceStub(self._channel)
            logger.info("gRPC client connected to %s", self._target)
        except ImportError:
            logger.warning(
                "gRPC proto stubs not available; "
                "run proto generation or ensure grpc_client/proto/ exists. "
                "gRPC calls will fail at runtime."
            )
            self._stub = None

    async def close(self) -> None:
        """Close the gRPC channel."""
        if self._channel:
            logger.debug("Closing gRPC channel to %s", self._target)
            await self._channel.close()
            self._channel = None
            self._stub = None

    async def apply_sequence(
        self,
        ontology_id: str,
        branch_id: str,
        steps: list[dict],
        commit_message: str = "AI-assisted ontology extraction",
        source_file: bytes | None = None,
        source_filename: str | None = None,
    ) -> dict:
        """Send an ApplySequence request to the ontology-service.

        Args:
            ontology_id: Target ontology ID.
            branch_id: Target branch ID.
            steps: List of sequence step dicts (see SequenceStep proto).
            commit_message: Commit message for the versioning operation.
            source_file: Optional raw source file bytes to attach as artifact.
            source_filename: Optional source filename for the artifact.

        Returns:
            ApplySequence response dict with keys:
                commit_id, steps_applied, steps_skipped, steps_failed, errors

        Raises:
            RuntimeError: If gRPC stub not available.
            grpc.RpcError: On gRPC communication failures.
        """
        if self._stub is None:
            raise RuntimeError("gRPC stub not available — proto stubs may not be generated")

        from grpc_client.proto.ontology.v1.ontology_pb2 import (  # type: ignore[import-untyped]
            ApplySequenceRequest,
        )

        # Build the request proto
        request = ApplySequenceRequest(
            ontology_id=ontology_id,
            branch_id=branch_id,
            commit_message=commit_message,
        )

        # Add steps
        from grpc_client.proto.ontology.v1.ontology_pb2 import (  # type: ignore[import-untyped]
            SequenceStep,
        )

        step_map = {
            "create_class": SequenceStep.OPERATION_CREATE_CLASS,
            "create_object_property": SequenceStep.OPERATION_CREATE_OBJECT_PROPERTY,
            "create_datatype_property": SequenceStep.OPERATION_CREATE_DATATYPE_PROPERTY,
            "create_individual": SequenceStep.OPERATION_CREATE_INDIVIDUAL,
            "add_annotation": SequenceStep.OPERATION_ADD_ANNOTATION,
            "set_parent": SequenceStep.OPERATION_SET_PARENT,
            "set_domain": SequenceStep.OPERATION_SET_DOMAIN,
            "set_range": SequenceStep.OPERATION_SET_RANGE,
        }

        for step_data in steps:
            op_name = step_data.get("operation", "").lower()
            op_value = step_map.get(op_name, SequenceStep.OPERATION_UNSPECIFIED)

            proto_step = SequenceStep(
                operation=op_value,
                entity_id=step_data.get("entity_id", ""),
                label=step_data.get("label", ""),
                parent_id=step_data.get("parent_id", ""),
                domain_id=step_data.get("domain_id", ""),
                range_id=step_data.get("range_id", ""),
                annotations=step_data.get("annotations", []),
                source_file=step_data.get("source_file", ""),
                skip_if_exists=step_data.get("skip_if_exists", False),
            )
            request.steps.append(proto_step)  # type: ignore[attr-defined]

        # Attach source file artifact if provided
        if source_file:
            request.source_file = source_file  # type: ignore[attr-defined]
        if source_filename:
            request.source_filename = source_filename  # type: ignore[attr-defined]

        logger.debug(
            "Sending ApplySequence: ontology=%s branch=%s steps=%d",
            ontology_id,
            branch_id,
            len(steps),
        )

        try:
            response = await self._stub.ApplySequence(request)  # type: ignore[union-attr]
            result = {
                "commit_id": response.commit_id,
                "steps_applied": response.steps_applied,
                "steps_skipped": response.steps_skipped,
                "steps_failed": response.steps_failed,
                "errors": list(response.errors),
            }
            logger.info(
                "ApplySequence result: commit=%s applied=%d skipped=%d failed=%d",
                result["commit_id"],
                result["steps_applied"],
                result["steps_skipped"],
                result["steps_failed"],
            )
            return result
        except grpc.RpcError as exc:
            logger.error(
                "gRPC ApplySequence failed: %s (%s)",
                exc.details() if hasattr(exc, "details") else str(exc),
                exc.code() if hasattr(exc, "code") else "unknown",
            )
            raise


# Singleton instance
_client: OntologyGrpcClient | None = None


async def get_client() -> OntologyGrpcClient:
    """Get or create the singleton gRPC client instance.

    The client is lazily initialized on first access and connected.
    """
    global _client
    if _client is None:
        _client = OntologyGrpcClient()
        await _client.connect()
    return _client


async def close_client() -> None:
    """Close the singleton gRPC client if it exists."""
    global _client
    if _client:
        await _client.close()
        _client = None
