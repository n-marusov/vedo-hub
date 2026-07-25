"""Tests for the sequence validator.

Validates: REQ-FUN.API.doc-extract-sequence-schema
Validates: REQ-FUN.CROSS.sequences
"""

from __future__ import annotations

from sequence.validator import validate_llm_response, validate_sequence


class TestValidateSequence:
    """Tests for validate_sequence()."""

    def test_valid_create_class(self) -> None:
        """A single valid create_class step."""
        steps = [{"operation": "create_class", "entity_id": "person", "label": "Person"}]
        result = validate_sequence(steps)
        assert result.valid is True
        assert len(result.steps) == 1
        assert result.steps[0].operation == "create_class"
        assert result.steps[0].label == "Person"

    def test_invalid_operation(self) -> None:
        """Step with invalid operation should fail."""
        steps = [{"operation": "delete_everything", "entity_id": "x"}]
        result = validate_sequence(steps)
        assert result.valid is False
        assert len(result.errors) > 0

    def test_missing_operation(self) -> None:
        """Step without operation should fail."""
        steps = [{"entity_id": "x", "label": "Test"}]
        result = validate_sequence(steps)
        assert result.valid is False
        assert any("missing 'operation'" in e for e in result.errors)

    def test_empty_sequence(self) -> None:
        """Empty sequence should fail."""
        result = validate_sequence([])
        assert result.valid is False
        assert any("Empty sequence" in e for e in result.errors)

    def test_multiple_steps(self) -> None:
        """Multiple valid steps should succeed."""
        steps = [
            {"operation": "create_class", "entity_id": "person", "label": "Person"},
            {
                "operation": "create_datatype_property",
                "entity_id": "has_name",
                "label": "has name",
                "domain_id": "person",
                "range_id": "xsd:string",
            },
        ]
        result = validate_sequence(steps)
        assert result.valid is True
        assert len(result.steps) == 2

    def test_entity_id_with_warning(self) -> None:
        """Missing entity_id should produce a warning but not fail."""
        steps = [{"operation": "create_class", "label": "No ID"}]
        result = validate_sequence(steps)
        # Warning about missing entity_id, but still valid
        assert result.valid is True
        assert len(result.warnings) > 0

    def test_all_operation_types(self) -> None:
        """All valid operation types should be accepted."""
        ops = [
            "create_class",
            "create_object_property",
            "create_datatype_property",
            "create_individual",
            "add_annotation",
            "set_parent",
            "set_domain",
            "set_range",
        ]
        for op in ops:
            result = validate_sequence([{"operation": op, "entity_id": "test"}])
            assert result.valid, f"Operation '{op}' should be valid"


class TestValidateLlmResponse:
    """Tests for validate_llm_response()."""

    def test_valid_json_steps_array(self) -> None:
        """Valid JSON with steps array should pass."""
        raw = '{"steps": [{"operation": "create_class", "entity_id": "person", "label": "Person"}]}'
        result = validate_llm_response(raw)
        assert result.valid is True
        assert len(result.steps) == 1

    def test_invalid_json(self) -> None:
        """Invalid JSON should fail."""
        result = validate_llm_response("{not valid json}")
        assert result.valid is False
        assert any("Invalid JSON" in e for e in result.errors)

    def test_json_with_code_fence(self) -> None:
        """JSON wrapped in markdown code fence should be parsed."""
        raw = '```json\n{"steps": [{"operation": "create_class", "entity_id": "person", "label": "Person"}]}\n```'
        result = validate_llm_response(raw)
        assert result.valid is True
        assert len(result.steps) == 1

    def test_json_with_sequence_key(self) -> None:
        """Response with 'sequence' key instead of 'steps' should work."""
        raw = '{"sequence": [{"operation": "create_class", "entity_id": "test"}]}'
        result = validate_llm_response(raw)
        assert result.valid is True

    def test_direct_array_no_wrapper(self) -> None:
        """Raw array without wrapper object should work."""
        raw = '[{"operation": "create_class", "entity_id": "test", "label": "Test"}]'
        result = validate_llm_response(raw)
        assert result.valid is True

    def test_response_with_warnings(self) -> None:
        """Steps with missing labels should produce warnings."""
        raw = '[{"operation": "create_class", "entity_id": "test"}]'
        result = validate_llm_response(raw)
        assert result.valid is True
        assert len(result.warnings) > 0

    def test_empty_steps_array(self) -> None:
        """Empty steps array should fail."""
        raw = '{"steps": []}'
        result = validate_llm_response(raw)
        assert result.valid is False
