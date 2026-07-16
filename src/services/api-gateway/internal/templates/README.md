# Ontology Domain Templates

Pre-built ontology templates for common domains. These templates are loaded
at API Gateway startup and can be applied to any ontology via the REST API.

## Available Templates

| ID | Name | Domain | Classes | Properties |
|----|------|--------|---------|------------|
| `person` | Person & Organization | General | Person, Organization, Event, Location, Document, ContactPoint | 11 |
| `product` | Product & Catalog | E-commerce | Product, Category, Manufacturer, Review, Price, Inventory | 12 |
| `software` | Software & System Architecture | IT & Software | SoftwareComponent, APIService, Database, Dependency, Version, DeploymentEnvironment | 10 |
| `medical` | Healthcare & Clinical | Healthcare | Patient, Diagnosis, Treatment, Medication, MedicalRecord, Symptom | 11 |

## API Endpoints

- `GET /api/v1/templates/ontologies` — list available templates
- `POST /api/v1/ontologies/:id/apply-template` — apply a template to an ontology

## Creating New Templates

1. Create a new JSON file in `internal/templates/ontologies/`
2. Follow the existing schema:
   - `id` — unique identifier (kebab-case)
   - `name` — human-readable name
   - `description` — brief description
   - `domain` — domain category
   - `steps` — array of SequenceStep objects with:
     - `operation`: CREATE_CLASS, CREATE_DATATYPE_PROPERTY, or CREATE_OBJECT_PROPERTY
     - `entity_id`: unique identifier (PascalCase for classes, camelCase for properties)
     - `label`: human-readable label
     - `parent_id`: parent class (default: "Thing")
     - `domain`: domain class for properties
     - `range`: range for properties (xsd type or class name)
     - `annotations`: optional annotations map

## Constraints

- Templates are read-only (cannot be modified via API in M2)
- Templates are loaded at gateway startup
- Maximum 50 steps per template
