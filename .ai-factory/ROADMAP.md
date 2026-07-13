# VEDO Hub Roadmap

> GitHub + Hugging Face for ontologies: social hub, API infrastructure, and LLM generation in one product. Ontologies become living assets — forked, starred, reviewed, cited via DOI, and consumed through API.

## Milestones

- [x] **M0: Foundation & Infrastructure** — Microservices architecture, Docker Compose orchestration, CI/CD pipeline, basic observability (OTEL/Grafana/Prometheus/Loki/Tempo), Keycloak auth, CLI tool (`vedo-cli` with ticket management, secret rotation, guardrail controls)
- [x] **M1: Ontology Core Engine** — Graph browser and navigation (tree of classes, property lists, basic 2D visualization), ontology editor TBox (CRUD for classes, ObjectProperty, DatatypeProperty, annotations), individual management ABox (class instances, property values), Git-like versioning (commits, history, branches, rollback), import/export (Turtle, RDF/XML), REST API for read/write operations, SPARQL endpoint
- [ ] **M2: Collaboration & Social Hub** — Comments and discussions on entities, social hub features (profiles, stars, forks, issues), LLM generation of ontologies from natural language (NL → OWL), visual SPARQL query builder (drag-n-drop), flexible graph queries (SPARQL/CYPHER editors with autocomplete, NL mode, MCP server) with unified response format and saved query library
- [ ] **M3: CLI Administration & Operations** — `vedo-cli` production operations: scheduled backup/restore with verify, Neo4j/PostgreSQL migrations with rollback, ontology export/import between environments, `diff` of two ontology versions, incident diagnosis by `trace_id` with LLM-assisted recommendations, portal support (ticket intake, in-app feedback widget with metadata, NPS collection), knowledge base (Antora docs), public status page
- [ ] **M4: API as Infrastructure Layer** — Semantic search across public ontologies, SPARQL endpoint as managed service, webhooks for change notifications, CRUD ontology operations via API, CI/CD webhook automation, ontology template library, Excel import → OWL via LLM, saved SPARQL query sharing
- [ ] **M5: Advanced Ontology Features** — Complex OWL constructs (class equivalence, intersection, union, complement), property characteristics (transitivity, symmetry, functionality, inverse), visual graph editor with drag-n-drop relationship editing, Pull Requests with semantic diff and review (GitHub-style), DOI registration for published ontologies, ontology templates, community forum, public roadmap with voting, usage telemetry (opt-in)
- [ ] **M6: Enterprise & Marketplace** — OWL consistency validation via API (semantic linter at PR), entity resolution via API, LLM generation via API (NL → OWL programmatically), fork/versioning via API, partner commissions for popular ontologies, author sponsoring (GitHub Sponsors-style), webhooks for CI/CD automation
- [ ] **M7: Scale & Intelligence** — Large-graph visualization (WebGL, virtualized), advanced ontology analytics (hierarchy depth, branching factor, data coverage), Enterprise tier (dedicated API endpoints, on-prem API layer, SLA), GitHub/GitLab sync plugins, AI assistant for ontology completion and relationship hints

## Completed

| Milestone | Date |
|-----------|------|
| M0: Foundation & Infrastructure | 2026-07-12 |
| M1: Ontology Core Engine | 2026-07-14 |

## Roadmap Notes

- **M1–M3** correspond to the MVP scope defined in the vision specification (section 2.5). These cover the 16 functional domains (F1–F16) at their basic level.
- **M5** corresponds to Stage 2 of the vision document (social hub and LLM generation, section 2.6).
- **M6** corresponds to Stage 3 (infrastructure layer and monetization).
- **M7** corresponds to Stage 4 (Enterprise and deepening).
- **M4** (API as infrastructure) runs as a horizontal layer across M1–M7, gradually deepening with each milestone.
- Milestones are ordered by dependency: each builds on the previous one.
- NFRs (performance p95 < 120ms, security gates, observability) apply continuously across all milestones, not as standalone deliverables.
