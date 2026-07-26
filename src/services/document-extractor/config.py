"""Document extractor configuration — pydantic-settings from environment."""

from __future__ import annotations

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # ─── Service ────────────────────────────────────────────────────────────
    SERVICE_PORT: int = 8092
    """HTTP port for health/metrics endpoints."""

    SERVICE_NAME: str = "document-extractor"
    """Service name used in logging and metrics."""

    SERVICE_VERSION: str = "0.1.0"

    MAX_FILE_SIZE_MB: int = 20
    """Maximum file size (MB) for synchronous parsing. Files larger than this
    trigger an async/suggested-async response."""

    # ─── gRPC ───────────────────────────────────────────────────────────────
    ONTOLOGY_SERVICE_URL: str = "ontology-service:9001"
    """gRPC target for the ontology-service ApplySequence RPC."""

    AI_ORCHESTRATION_URL: str = "ai-orchestration-service:9014"
    """gRPC target for the ai-orchestration-service CheckPolicy and LogLLMUsage RPCs."""

    # ─── LLM ─────────────────────────────────────────────────────────────────
    LLM_PROVIDER: str = "openai"
    """LLM provider name: openai, anthropic, or custom-compatible API."""

    LLM_API_KEY: str = ""
    """API key for the LLM provider."""

    LLM_BASE_URL: str = ""
    """Optional base URL for OpenAI-compatible API (e.g., local LLM proxy)."""

    LLM_MODEL: str = "gpt-4o-mini"
    """Model identifier to use for ontology extraction."""

    LLM_MAX_RETRIES: int = 3
    """Maximum LLM call retries with exponential backoff."""

    LLM_TIMEOUT_SECONDS: int = 120
    """Timeout for individual LLM API calls."""

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")


settings = Settings()
"""Application-wide singleton settings instance."""
