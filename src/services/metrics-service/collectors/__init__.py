"""Event collectors for the metrics service.

Consumes domain events from RabbitMQ (commit, import, publish events)
and forwards them to the MetricsComputer for processing.
"""

import asyncio
import json
import logging
from collections.abc import Callable
from typing import Any

logger = logging.getLogger(__name__)


class EventCollector:
    """Collects domain events and dispatches them to handlers.

    For MVP, events are consumed via RabbitMQ. When RabbitMQ is unavailable,
    the collector falls back to an HTTP ingestion endpoint for testing.
    """

    def __init__(
        self,
        rabbitmq_url: str | None = None,
        event_handler: Callable[[dict[str, Any]], Any] | None = None,
    ) -> None:
        self._rabbitmq_url = rabbitmq_url
        self._event_handler = event_handler
        self._consumed: int = 0
        self._failed: int = 0
        self._running = False

    async def start(self) -> None:
        """Start consuming events from RabbitMQ.

        Falls back gracefully if RabbitMQ is unreachable.
        """
        if not self._rabbitmq_url:
            logger.info("No RabbitMQ URL configured — event collector running in API-only mode")
            return

        self._running = True
        logger.info("Starting RabbitMQ event consumer at %s", self._rabbitmq_url)

        try:
            import aio_pika

            connection = await aio_pika.connect_robust(self._rabbitmq_url)

            async with connection:
                channel = await connection.channel()
                await channel.set_qos(prefetch_count=10)

                # Declare exchange and queue for ontology events
                exchange = await channel.declare_exchange(
                    "vedo.ontology.events",
                    aio_pika.ExchangeType.TOPIC,
                    durable=True,
                )

                queue = await channel.declare_queue(
                    "metrics.ontology.events",
                    durable=True,
                )

                await queue.bind(exchange, routing_key="ontology.#")
                await queue.bind(exchange, routing_key="versioning.#")

                logger.info("RabbitMQ consumer ready, waiting for events...")

                async with queue.iterator() as queue_iter:
                    async for message in queue_iter:
                        if not self._running:
                            break
                        async with message.process():
                            await self._process_message(message)
        except Exception as exc:
            logger.warning(
                "RabbitMQ consumer failed: %s — running in API-only mode", exc
            )

    async def stop(self) -> None:
        """Stop the event collector."""
        self._running = False
        logger.info("Event collector stopped")

    async def ingest_event(self, event: dict[str, Any]) -> bool:
        """Ingest a single event (HTTP fallback path)."""
        logger.debug("Ingesting event via HTTP: %s", event.get("type", "unknown"))
        return await self._dispatch_event(event)

    async def _process_message(self, message: Any) -> None:
        """Process a single RabbitMQ message."""
        try:
            body = message.body.decode("utf-8")
            event = json.loads(body)
            logger.debug(
                "Received event: %s (routing_key: %s)",
                event.get("type", "unknown"),
                message.routing_key,
            )
            await self._dispatch_event(event)
        except json.JSONDecodeError as exc:
            self._failed += 1
            logger.warning("Failed to decode event message: %s", exc)
        except Exception as exc:
            self._failed += 1
            logger.warning("Failed to process event: %s", exc)

    async def _dispatch_event(self, event: dict[str, Any]) -> bool:
        """Dispatch an event to the configured handler."""
        if self._event_handler:
            try:
                result = self._event_handler(event)
                if asyncio.iscoroutine(result):
                    await result
                self._consumed += 1
                logger.debug("Event dispatched successfully (total: %d)", self._consumed)
                return True
            except Exception as exc:
                self._failed += 1
                logger.error("Event handler failed: %s", exc)
                return False
        return False

    @property
    def stats(self) -> dict[str, int]:
        """Return collector statistics."""
        return {
            "consumed": self._consumed,
            "failed": self._failed,
        }
