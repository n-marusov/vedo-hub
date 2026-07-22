# VEDO Demos — Seed Data

This directory contains bootstrap seed data for the `VEDO Demos` group — 5 demo projects
that users can fork to get started with VEDO Core.

## Contents

| File | Demo Project | Key Classes |
|------|-------------|-------------|
| `organization.json` | Organization | Organization, Department, Employee, Position |
| `product.json` | Product | Product, Category, Manufacturer, Review |
| `process.json` | Process | Process, Step, Input, Output, Agent |
| `glossary.json` | Glossary | Term, Definition, RelatedTerm, Source |
| `event.json` | Event | Event, Participant, Location, DateTime |
| `bootstrap.sh` | Bootstrap script | Creates group + projects + applies sequences |

## Usage

```bash
export VEDO_ADMIN_TOKEN="<admin-jwt>"
./bootstrap.sh
```

The script is idempotent — safe to run multiple times.

## Architecture

Each demo project is a regular `Project` with a 1:1 paired `Ontology`. The bootstrap script:
1. Creates the `VEDO Demos` group
2. Creates 5 projects, each with a paired ontology
3. Applies the JSON sequence via `ontology.ApplySequence` gRPC (the same Published Language used by document-extractor)
4. Sets visibility to `public`

Per `ADR-DES.PROCESS.templates-via-forks`, demo projects replace the former `Ontology Template` concept.
There is no separate semver, catalog, metadata.json, or lifecycle review — those mechanisms were
premature automation for 5 projects. Users fork demo projects via `POST /api/v1/projects/{id}/fork`.
