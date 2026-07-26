"""Tests for the JSON parser."""

from __future__ import annotations

import json
import os
import tempfile

import pytest

from parsers.json_parser import JsonParser


@pytest.fixture
def parser() -> JsonParser:
    return JsonParser()


@pytest.mark.asyncio
async def test_parse_simple_json_object(parser: JsonParser) -> None:
    """Parse a simple JSON object with nested structure."""
    data = {
        "name": "John",
        "age": 30,
        "address": {
            "street": "123 Main St",
            "city": "NYC",
        },
        "hobbies": ["reading", "coding"],
    }
    with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
        json.dump(data, f)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "json"
        assert len(doc.sections) > 0
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_json_array(parser: JsonParser) -> None:
    """Parse a JSON array of objects."""
    data = [
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"},
        {"id": 3, "name": "Charlie"},
    ]
    with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
        json.dump(data, f)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "json"
        assert doc.metadata.get("root_is_array") is True
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_invalid_json(parser: JsonParser) -> None:
    """Attempt to parse invalid JSON."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
        f.write("{invalid json}")
        path = f.name
    try:
        with pytest.raises(Exception):
            await parser.parse(path)
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_empty_json(parser: JsonParser) -> None:
    """Parse empty JSON object."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
        json.dump({}, f)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "json"
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_json_with_nested_objects(parser: JsonParser) -> None:
    """Parse JSON with several levels of nesting."""
    data = {
        "company": {
            "name": "TechCorp",
            "department": {
                "name": "Engineering",
                "team": {
                    "name": "Platform",
                    "members": ["John", "Jane"],
                },
            },
        }
    }
    with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
        json.dump(data, f)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "json"
    finally:
        os.unlink(path)
