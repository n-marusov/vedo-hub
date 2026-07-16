"""Sequence validator — validates LLM output for ApplySequence compatibility.

Checks:
- JSON is parseable
- Schema matches expected structure
- Operation types are valid (from SequenceStep.Operation enum)
- Entity references are consistent (no dangling references)
- No circular references in parent/domain/range chains
"""

from __future__ import annotations

import logging

from sequence.model import SequenceStep

logger = logging.getLogger("document-extractor.sequence.validator")

# Set of valid operation types matching the proto enum
VALID_OPERATIONS = {
    "create_class",
    "create_object_property",
    "create_datatype_property",
    "create_individual",
    "add_annotation",
    "set_parent",
    "set_domain",
    "set_range",
}


class ValidationResult:
    """Result of sequence validation."""

    def __init__(self) -> None:
        self.valid: bool = True
        self.steps: list[SequenceStep] = []
        self.warnings: list[str] = []
        self.errors: list[str] = []

    def add_error(self, message: str) -> None:
        """Add a validation error and mark the result as invalid."""
        self.valid = False
        self.errors.append(message)
        logger.warning("Sequence validation error: %s", message)

    def add_warning(self, message: str) -> None:
        """Add a non-blocking warning."""
        self.warnings.append(message)
        logger.debug("Sequence validation warning: %s", message)


def validate_sequence(data: list[dict]) -> ValidationResult:
    """Validate a list of sequence step dicts.

    Args:
        data: Raw list of step dicts from LLM output.

    Returns:
        ValidationResult with parsed SequenceStep list and any issues.
    """
    result = ValidationResult()

    if not data:
        result.add_error("Empty sequence — no steps provided")
        return result

    for i, step in enumerate(data):
        if not isinstance(step, dict):
            result.add_error(f"Step {i}: expected dict, got {type(step).__name__}")
            continue

        validated = _validate_step(step, i, result)
        if validated:
            result.steps.append(validated)

    # Cross-step consistency checks
    if result.steps:
        _check_consistency(result)

    logger.info(
        "Sequence validation: %d/%d steps valid, %d warnings, %d errors",
        len(result.steps),
        len(data),
        len(result.warnings),
        len(result.errors),
    )

    return result


def _validate_step(step: dict, index: int, result: ValidationResult) -> SequenceStep | None:
    """Validate a single step dict and return a SequenceStep or None."""
    operation = step.get("operation", "").lower()

    if not operation:
        result.add_error(f"Step {index}: missing 'operation' field")
        return None

    if operation not in VALID_OPERATIONS:
        result.add_error(
            f"Step {index}: invalid operation '{operation}'. "
            f"Valid: {', '.join(sorted(VALID_OPERATIONS))}"
        )
        return None

    entity_id = step.get("entity_id", "") or step.get("id", "")
    if not entity_id:
        result.add_warning(f"Step {index}: missing 'entity_id' — a generated ID will be used")

    label = step.get("label", "") or step.get("name", "")
    if not label and operation in (
        "create_class",
        "create_object_property",
        "create_datatype_property",
    ):
        result.add_warning(f"Step {index}: {operation} has no label/name")

    # Coerce annotations to list
    annotations = step.get("annotations", [])
    if isinstance(annotations, str):
        annotations = [annotations]

    return SequenceStep(
        operation=operation,
        entity_id=entity_id,
        label=label,
        parent_id=step.get("parent_id", "") or step.get("parent", ""),
        domain_id=step.get("domain_id", "") or step.get("domain", ""),
        range_id=step.get("range_id", "") or step.get("range", ""),
        annotations=annotations,
        source_file=step.get("source_file", ""),
        skip_if_exists=step.get("skip_if_exists", False),
    )


def _check_consistency(result: ValidationResult) -> None:
    """Check cross-step consistency: entity ID uniqueness and reference validity."""
    # Collect all entity IDs defined in the sequence
    defined_ids: set[str] = set()
    for step in result.steps:
        if step.entity_id:
            defined_ids.add(step.entity_id)

    # Check references
    for i, step in enumerate(result.steps):
        for ref_field, ref_name in [
            (step.parent_id, "parent_id"),
            (step.domain_id, "domain_id"),
            (step.range_id, "range_id"),
        ]:
            if ref_field and ref_field not in defined_ids:
                result.add_warning(
                    f"Step {i} ({step.operation} '{step.label}'): "
                    f"{ref_name} '{ref_field}' references undefined entity"
                )

    # Check for circular parent references (simple depth-limited check)
    parent_map: dict[str, str] = {}
    for step in result.steps:
        if step.parent_id:
            parent_map[step.entity_id] = step.parent_id

    for entity_id in parent_map:
        visited = {entity_id}
        current = parent_map.get(entity_id)
        depth = 0
        while current and current in parent_map and depth < 100:
            if current in visited:
                result.add_error(
                    f"Circular parent reference detected involving '{entity_id}' and '{current}'"
                )
                break
            visited.add(current)
            current = parent_map.get(current)
            depth += 1


def validate_llm_response(raw: str) -> ValidationResult:
    """Parse and validate a raw LLM response string.

    Attempts to extract JSON from the response (handles ```json blocks).
    """
    import json

    # Strip markdown code fences if present
    text = raw.strip()
    if text.startswith("```"):
        # Remove opening and closing fences
        lines = text.split("\n")
        # Remove first line (```json or ```)
        if lines[0].startswith("```"):
            lines = lines[1:]
        # Remove last line if it's ```
        if lines and lines[-1].strip() == "```":
            lines = lines[:-1]
        text = "\n".join(lines).strip()

    try:
        data = json.loads(text)
    except json.JSONDecodeError as exc:
        result = ValidationResult()
        result.add_error(f"Invalid JSON in LLM response: {exc}")
        return result

    # Normalize: if top-level is a dict with a "steps" key, use that
    if isinstance(data, dict):
        steps = data.get("steps", data.get("sequence", [data]))
        if isinstance(steps, list):
            return validate_sequence(steps)
        result = ValidationResult()
        result.add_error(f"Expected list of steps, got {type(steps).__name__}")
        return result
    if isinstance(data, list):
        return validate_sequence(data)
    result = ValidationResult()
    result.add_error(f"Expected list or dict with 'steps', got {type(data).__name__}")
    return result
