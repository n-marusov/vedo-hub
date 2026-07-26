"""Ontology metric computation.

Computes and caches ontology metrics such as axiom count, class depth,
property density, and complexity trends. Metrics are stored in Redis
for dashboard queries and exposed via Prometheus.
"""

import json
import logging
from collections.abc import Callable
from dataclasses import dataclass, field, asdict
from typing import Any

import redis.asyncio as aioredis

logger = logging.getLogger(__name__)


@dataclass
class OntologyMetrics:
    """Computed metrics for a single ontology snapshot."""

    ontology_id: str = ""
    axiom_count: int = 0
    class_count: int = 0
    property_count: int = 0
    individual_count: int = 0
    max_class_depth: int = 0
    avg_class_depth: float = 0.0
    property_density: float = 0.0
    class_depth_scores: dict[str, int] = field(default_factory=dict)


class MetricsComputer:
    """Computes and caches ontology metrics.

    For MVP, metrics are derived from event payload data. In production,
    the computer queries the ontology-service for full graph analysis.
    """

    def __init__(self, redis_url: str | None = None) -> None:
        self._redis: aioredis.Redis | None = None
        self._redis_url = redis_url
        self._local_cache: dict[str, OntologyMetrics] = {}
        self._stored_metrics_count: int = 0

    async def connect(self) -> None:
        """Connect to Redis if a URL is configured."""
        if self._redis_url:
            try:
                self._redis = aioredis.from_url(
                    self._redis_url,
                    decode_responses=True,
                    socket_connect_timeout=5,
                )
                await self._redis.ping()
                logger.info("Connected to Redis at %s", self._redis_url)
            except Exception as exc:
                logger.warning("Redis connection failed: %s — running with local cache only", exc)
                self._redis = None
        else:
            logger.info("No Redis URL configured — running with local cache only")

    async def compute_from_event(self, event: dict[str, Any]) -> OntologyMetrics:
        """Compute metrics from an event payload.

        The event is expected to have fields like:
        - ontology_id, type (commit/import/publish)
        - class_count, property_count, individual_count
        - class_depth_scores (dict of class_id -> depth)
        - axiom_count
        """
        ontology_id = event.get("ontology_id", "unknown")
        logger.debug(
            "Computing metrics for ontology %s from event type %s",
            ontology_id,
            event.get("type", "unknown"),
        )

        class_count = event.get("class_count", 0)
        property_count = event.get("property_count", 0)
        individual_count = event.get("individual_count", 0)
        axiom_count = event.get("axiom_count", 0)

        depth_scores: dict[str, int] = event.get("class_depth_scores", {})
        max_depth = max(depth_scores.values()) if depth_scores else 0
        avg_depth = (
            sum(depth_scores.values()) / len(depth_scores) if depth_scores else 0.0
        )

        # Property density: properties per class
        total_entities = class_count + property_count + individual_count
        property_density = (
            round(property_count / class_count, 2) if class_count > 0 else 0.0
        )

        metrics = OntologyMetrics(
            ontology_id=ontology_id,
            axiom_count=axiom_count,
            class_count=class_count,
            property_count=property_count,
            individual_count=individual_count,
            max_class_depth=max_depth,
            avg_class_depth=round(avg_depth, 2),
            property_density=property_density,
            class_depth_scores=depth_scores,
        )

        # Cache locally
        self._local_cache[ontology_id] = metrics

        # Store to Redis if available
        if self._redis:
            await self._store_to_redis(metrics)

        logger.info(
            "Metrics computed for ontology %s: %d classes, %d properties, "
            "%d individuals, max depth %d",
            ontology_id,
            class_count,
            property_count,
            individual_count,
            max_depth,
        )

        return metrics

    async def get_metrics(self, ontology_id: str) -> OntologyMetrics | None:
        """Retrieve cached metrics for an ontology."""
        # Check local cache first
        if ontology_id in self._local_cache:
            return self._local_cache[ontology_id]

        # Fall back to Redis
        if self._redis:
            try:
                data = await self._redis.get(f"metrics:{ontology_id}")
                if data:
                    metrics_dict = json.loads(data)
                    metrics = OntologyMetrics(**metrics_dict)
                    self._local_cache[ontology_id] = metrics
                    return metrics
            except Exception as exc:
                logger.warning("Failed to read metrics from Redis: %s", exc)

        return None

    async def get_all_metrics(self) -> list[OntologyMetrics]:
        """Retrieve all cached metrics."""
        if self._redis:
            try:
                keys = await self._redis.keys("metrics:*")
                results: list[OntologyMetrics] = []
                for key in keys:
                    data = await self._redis.get(key)
                    if data:
                        metrics_dict = json.loads(data)
                        results.append(OntologyMetrics(**metrics_dict))
                return results
            except Exception as exc:
                logger.warning("Failed to read all metrics from Redis: %s", exc)

        return list(self._local_cache.values())

    async def _store_to_redis(self, metrics: OntologyMetrics) -> None:
        """Store metrics snapshot in Redis with TTL."""
        if not self._redis:
            return

        key = f"metrics:{metrics.ontology_id}"
        try:
            await self._redis.setex(
                key,
                86400,  # 24-hour TTL
                json.dumps(asdict(metrics), default=str),
            )
            self._stored_metrics_count += 1
            logger.debug(
                "Stored metrics for %s in Redis (total stored: %d)",
                metrics.ontology_id,
                self._stored_metrics_count,
            )
        except Exception as exc:
            logger.warning("Failed to store metrics in Redis: %s", exc)

    async def get_stored_count(self) -> int:
        """Return the number of metrics stored in Redis."""
        if self._redis:
            try:
                keys = await self._redis.keys("metrics:*")
                return len(keys)
            except Exception:
                pass
        return len(self._local_cache)

    async def close(self) -> None:
        """Close the Redis connection."""
        if self._redis:
            await self._redis.close()
            logger.info("Redis connection closed")
