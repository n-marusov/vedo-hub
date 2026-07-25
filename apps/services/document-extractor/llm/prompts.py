"""Prompt templates for ontology extraction from documents.

Templates are designed to progressively guide the LLM to produce structured
JSON sequences for the ApplySequence RPC.
"""

from __future__ import annotations

# ─── System Prompt ─────────────────────────────────────────────────────────────

SYSTEM_PROMPT = """You are an ontology extraction AI. Your task is to analyze documents and extract
OWL ontology structures from them. You MUST output ONLY a valid JSON object
containing a "steps" array. No explanations, no markdown, no extra text.

Each step in the "steps" array represents one ontology operation. The available
operations and their required fields are:

1. **create_class** — Create a new OWL class.
   Required: entity_id (string), label (string)
   Optional: parent_id (string), annotations ([string])

2. **create_object_property** — Create an object property (links classes).
   Required: entity_id, label
   Optional: domain_id, range_id, annotations

3. **create_datatype_property** — Create a datatype property (links class to literal).
   Required: entity_id, label
   Optional: domain_id, range_id (xsd type), annotations

4. **create_individual** — Create an individual (instance).
   Required: entity_id, label
   Optional: parent_id (class type), annotations

5. **add_annotation** — Add annotation to an existing entity.
   Required: entity_id, annotations

6. **set_parent** — Set parent class (subclass relationship).
   Required: entity_id, parent_id

7. **set_domain** / **set_range** — Set property domain or range.
   Required: entity_id, domain_id or range_id

Rules:
- Use snake_case for entity_id values (e.g., "person", "has_name")
- Reuse entity_id values when referring to the same entity
- Create parent classes before child classes
- Create classes before properties that reference them
- Annotations are key=value strings (e.g., "description=Represents a person")
- Use xsd:string, xsd:integer, xsd:float, xsd:boolean for datatype ranges

Output format:
{"steps": [
  {"operation": "create_class", "entity_id": "...", "label": "...", "parent_id": "...", "annotations": ["description=..."]}
]}"""

# ─── Extraction Prompts ────────────────────────────────────────────────────────

EXTRACTION_PROMPT_V1 = """Analyze the following document and extract an ontology structure from it.

Document title: {title}
Document format: {format}

Document content:
{content}

Extract classes, properties, and individuals as a sequence of ontology operations.
Focus on the main entities and their relationships."""

EXTRACTION_PROMPT_V2 = """Extract a complete OWL ontology from the following document.

Document: {title} ({format})

Content:
{content}

Requirements:
1. Identify all main entities (classes) and their hierarchy
2. Identify object properties (relationships between classes)
3. Identify datatype properties (attributes of classes)
4. Identify any individuals/instances
5. Output as a JSON "steps" array

Focus on accuracy over completeness. Only extract what you are confident about."""

EXTRACTION_PROMPT_V3 = """Extract an ontology from the following document with detailed structural analysis.

Document: {title} ({format})

Full content:
{content}

Section analysis:
{section_analysis}

For each section, identify:
1. The main topic → Class
2. Sub-topics → Subclasses
3. Relationships mentioned → Object Properties
4. Attributes mentioned → Datatype Properties
5. Examples listed → Individuals

Output as a JSON "steps" array. Include:
- All entity_ids in snake_case
- Descriptive labels
- Parent/child relationships via parent_id
- Domain and range for properties
- Annotations for descriptions

IMPORTANT: Order the steps so parent classes appear before their children,
and classes appear before properties that reference them."""


# ─── Retry Prompts ─────────────────────────────────────────────────────────────

RETRY_PROMPT = """The previous extraction attempt had issues. Please fix the following:

{warnings}

Errors:
{errors}

Original document:
{content}

Please provide a corrected JSON "steps" array only."""

SIMPLE_RETRY_PROMPT = """Please try again. The previous response was not valid.

Requirements:
- Output ONLY valid JSON with a "steps" array
- Each step must have "operation" and "entity_id"
- Valid operations: {valid_operations}

Document:
{content}"""


# ─── Prompt Builder ────────────────────────────────────────────────────────────


def build_extraction_prompt(
    content: str,
    title: str,
    fmt: str = "text",
    sections: list[dict] | None = None,
    retry_count: int = 0,
    validation_errors: list[str] | None = None,
    validation_warnings: list[str] | None = None,
) -> list[dict]:
    """Build a list of message dicts for the LLM extraction call.

    Progressively more detailed prompts with each retry.
    """
    # Truncate content to avoid token limits (approx 8000 chars = ~2000 tokens)
    max_content_chars = 8000
    truncated = content[:max_content_chars]
    if len(content) > max_content_chars:
        truncated += f"\n\n[...content truncated from {len(content)} to {max_content_chars} chars]"

    # Build section analysis if sections provided
    section_analysis = ""
    if sections:
        analysis_lines = []
        for s in sections:
            heading = s.get("heading", "")
            level = s.get("level", 1)
            para_count = len(s.get("paragraphs", []))
            table_count = len(s.get("tables", []))
            analysis_lines.append(
                f"{'  ' * level}H{level}: {heading} ({para_count} paragraphs, {table_count} tables)"
            )
        section_analysis = "\n".join(analysis_lines)

    messages = [{"role": "system", "content": SYSTEM_PROMPT}]

    if retry_count == 0:
        # First attempt: use progressive prompt based on section detail
        if sections and len(sections) > 3:
            prompt = EXTRACTION_PROMPT_V3.format(
                title=title,
                format=fmt,
                content=truncated,
                section_analysis=section_analysis,
            )
        elif sections:
            prompt = EXTRACTION_PROMPT_V2.format(
                title=title,
                format=fmt,
                content=truncated,
            )
        else:
            prompt = EXTRACTION_PROMPT_V1.format(
                title=title,
                format=fmt,
                content=truncated,
            )
    elif retry_count == 1:
        prompt = RETRY_PROMPT.format(
            warnings="; ".join(validation_warnings or []),
            errors="; ".join(validation_errors or []),
            content=truncated,
        )
    else:
        prompt = SIMPLE_RETRY_PROMPT.format(
            valid_operations=", ".join(
                [
                    "create_class",
                    "create_object_property",
                    "create_datatype_property",
                    "create_individual",
                ]
            ),
            content=truncated,
        )

    messages.append({"role": "user", "content": prompt})
    return messages
