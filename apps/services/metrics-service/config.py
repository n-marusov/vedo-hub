"""Metrics service configuration — pydantic-settings from environment."""

from __future__ import annotations

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    # ─── Service ────────────────────────────────────────────────────────────
    SERVICE_PORT: int = 8084
    """HTTP port for REST API and metrics endpoints."""

    SERVICE_NAME: str = "metrics-service"
    """Service name used in logging and metrics."""

    SERVICE_VERSION: str = "0.1.0"

    # ─── RabbitMQ ───────────────────────────────────────────────────────────
    RABBITMQ_URL: str = "amqp://guest:guest@rabbitmq:5672/"
    """RabbitMQ connection URL for consuming telemetry events."""

    # ─── Redis ──────────────────────────────────────────────────────────────
    REDIS_URL: str = "redis://redis:6379/0"
    """Redis connection URL for storing aggregated metrics."""


settings = Settings()
