"""Sequence models for ontology extraction operations."""

from __future__ import annotations

from pydantic import BaseModel, Field


class SequenceStep(BaseModel):
    """A single operation in an ontology extraction sequence.

    Maps directly to the SequenceStep protobuf message for ApplySequence.
    """

    operation: str = Field(
        description="Operation type: create_class, create_object_property, "
        "create_datatype_property, create_individual, add_annotation, "
        "set_parent, set_domain, set_range"
    )
    entity_id: str = Field(default="", description="IRI or generated entity ID")
    label: str = Field(default="", description="rdfs:label")
    parent_id: str = Field(default="", description="Parent class ID (for classes)")
    domain_id: str = Field(default="", description="Domain class ID (for properties)")
    range_id: str = Field(default="", description="Range class ID (for properties)")
    annotations: list[str] = Field(default_factory=list, description="key=value pairs")
    source_file: str = Field(default="", description="Source file reference (batch)")
    skip_if_exists: bool = Field(default=False, description="Skip if entity exists")


class ApplySequenceRequest(BaseModel):
    """Request model for the apply sequence operation."""

    ontology_id: str = Field(..., description="Target ontology ID")
    branch_id: str = Field(default="main", description="Target branch ID")
    steps: list[SequenceStep] = Field(default_factory=list)
    commit_message: str = Field(default="AI-assisted ontology extraction")
    source_file: str | None = Field(default=None, description="Original file as base64 artifact")
    source_filename: str | None = Field(default=None)


class ApplySequenceResponse(BaseModel):
    """Response model for the apply sequence operation."""

    commit_id: str = ""
    steps_applied: int = 0
    steps_skipped: int = 0
    steps_failed: int = 0
    errors: list[str] = Field(default_factory=list)
    success: bool = True


class ExtractionResult(BaseModel):
    """Result of an extraction (parse + LLM + validation) pipeline."""

    filename: str = ""
    format: str = ""
    file_size_bytes: int = 0
    parsed_sections: int = 0
    raw_text_length: int = 0
    sequence: list[SequenceStep] = Field(default_factory=list)
    step_count: int = 0
    warnings: list[str] = Field(default_factory=list)
    error: str | None = None
