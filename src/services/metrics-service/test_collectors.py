"""Tests for EventCollector — domain event collection and dispatch."""

from __future__ import annotations

from collectors import EventCollector


class TestEventCollector:
    """EventCollector unit tests — no RabbitMQ, mocked handlers."""

    async def test_ingest_event_calls_handler(self) -> None:
        """ingest_event calls the configured event handler."""
        received: list[dict] = []

        async def handler(event: dict) -> None:
            received.append(event)

        collector = EventCollector(event_handler=handler)
        result = await collector.ingest_event({"type": "commit", "ontology_id": "onto-1"})
        assert result is True
        assert len(received) == 1
        assert received[0]["type"] == "commit"

    async def test_ingest_event_no_handler(self) -> None:
        """ingest_event returns False when no handler is configured."""
        collector = EventCollector()
        result = await collector.ingest_event({"type": "commit"})
        assert result is False

    async def test_ingest_event_handler_failure(self) -> None:
        """ingest_event returns False when handler raises."""

        async def failing_handler(event: dict) -> None:
            raise ValueError("handler error")

        collector = EventCollector(event_handler=failing_handler)
        result = await collector.ingest_event({"type": "commit"})
        assert result is False

    async def test_stats_initial(self) -> None:
        """stats starts at zero for consumed and failed."""
        collector = EventCollector()
        stats = collector.stats
        assert stats["consumed"] == 0
        assert stats["failed"] == 0

    async def test_stats_after_ingest(self) -> None:
        """stats reflects consumed events."""

        async def handler(event: dict) -> None:
            pass

        collector = EventCollector(event_handler=handler)
        await collector.ingest_event({"type": "commit"})
        stats = collector.stats
        assert stats["consumed"] == 1
        assert stats["failed"] == 0

    async def test_stats_after_failure(self) -> None:
        """stats reflects failed events."""

        async def failing_handler(event: dict) -> None:
            raise RuntimeError("fail")

        collector = EventCollector(event_handler=failing_handler)
        await collector.ingest_event({"type": "commit"})
        stats = collector.stats
        assert stats["consumed"] == 0
        assert stats["failed"] == 1

    async def test_start_without_rabbitmq(self) -> None:
        """start without RabbitMQ URL logs and returns gracefully."""
        collector = EventCollector()
        await collector.start()  # should not block or raise
        assert collector._running is False

    async def test_stop_without_start(self) -> None:
        """stop without start is a no-op."""
        collector = EventCollector()
        await collector.stop()
        assert collector._running is False

    async def test_sync_handler(self) -> None:
        """EventCollector supports synchronous handlers."""
        received: list[dict] = []

        def sync_handler(event: dict) -> None:
            received.append(event)

        collector = EventCollector(event_handler=sync_handler)
        result = await collector.ingest_event({"type": "publish", "ontology_id": "onto-sync"})
        assert result is True
        assert len(received) == 1
