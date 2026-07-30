// Validates: REQ-FUN.API.graphql-sparql
// Validates: ADR-DES.API.graphql-sparql-split-strategy
// Validates: ADR-DES.API.rest-graphql-mutation-boundary
//
// GraphQL schema boundary tests — verify that the GraphQL endpoint exposes
// ONLY read navigation, not SPARQL execution or entity write mutations.
//
// These tests run against the deployed docker-compose.test.yml stack via
// playwright.api.config.ts. They use introspection to assert the schema
// shape mandated by the ADRs:
//
//   - `sparqlQuery` is NOT a field on the Query root.
//   - `createClass`, `createProperty`, `createIndividual`, `updateClass`,
//     `updateProperty`, `updateIndividual`, `deleteClass`, `deleteProperty`,
//     `deleteIndividual`, `createOntology`, `updateOntology`,
//     `deleteOntology` are NOT fields on the Mutation root.
//   - Navigation queries (`ontology`, `classes`, `class`, `properties`,
//     `individuals`, `commits`, `branches`) ARE present on the Query root.
//
// Per ADR-DES.API.rest-graphql-mutation-boundary, the long-term target is
// `EmptyMutation` (no Mutation type at all). The current schema still
// exposes deprecated placeholder mutations (`updateDraft`,
// `updateMemberRole`, `removeMember`) pending frontend REST migration —
// these tests do NOT assert their absence, only the absence of the
// forbidden entity CRUD mutations and SPARQL.
import { test, expect } from '@playwright/test'

import { OWNER_JWT } from '../../jwt-tokens'

const BASE = 'http://localhost:3000/api/v1'
const AUTH = { Authorization: `Bearer ${OWNER_JWT}`, 'Content-Type': 'application/json' }

// Standard GraphQL introspection query — fetches the schema's type system.
// We only need __schema.queryType and __schema.mutationType fields, plus
// the full list of types with their fields to assert presence/absence.
const INTROSPECTION_QUERY = `query BoundaryIntrospection {
  __schema {
    queryType { name fields { name } }
    mutationType { name fields { name } }
  }
}`

interface IntrospectionField {
  name: string
}

interface IntrospectionType {
  name: string
  fields: IntrospectionField[] | null
}

interface IntrospectionSchema {
  queryType: IntrospectionType
  mutationType: IntrospectionType | null
}

interface GraphQLResponse {
  data?: { __schema: IntrospectionSchema }
  errors?: Array<{ message: string }>
}

async function fetchGraphQL(
  request: import('@playwright/test').APIRequestContext,
  query: string,
  variables: Record<string, unknown> = {},
): Promise<GraphQLResponse> {
  const params = new URLSearchParams({ query, variables: JSON.stringify(variables) })
  const res = await request.get(`${BASE}/graphql?${params}`, {
    headers: AUTH,
  })
  expect(res.status()).toBeLessThan(500)
  return (await res.json()) as GraphQLResponse
}

test.describe('GraphQL schema boundary — ADR-DES.API.rest-graphql-mutation-boundary', () => {
  let schema: IntrospectionSchema | null = null

  test.beforeAll(async ({ request }) => {
    const resp = await fetchGraphQL(request, INTROSPECTION_QUERY)
    if (resp.errors && resp.errors.length > 0) {
      throw new Error(`Introspection failed: ${JSON.stringify(resp.errors)}`)
    }
    if (!resp.data || !resp.data.__schema) {
      throw new Error('Introspection returned no __schema')
    }
    schema = resp.data.__schema
  })

  test('Query root exposes navigation fields', () => {
    expect(schema).not.toBeNull()
    const queryFields = schema!.queryType.fields?.map((f) => f.name) ?? []
    // Required navigation fields per ADR-DES.API.graphql-sparql-split-strategy.
    const required = [
      'classes',
      'class',
      'properties',
      'individuals',
      'commits',
      'branches',
    ]
    for (const field of required) {
      expect(queryFields, `Query root should expose ${field}`).toContain(field)
    }
  })

  test('Query root does NOT expose sparqlQuery (REST-only per ADR)', () => {
    expect(schema).not.toBeNull()
    const queryFields = schema!.queryType.fields?.map((f) => f.name) ?? []
    expect(queryFields, 'sparqlQuery must not appear on Query root').not.toContain('sparqlQuery')
    expect(queryFields, 'cypherQuery must not appear on Query root').not.toContain('cypherQuery')
  })

  test('Mutation root does NOT expose entity CRUD mutations', () => {
    expect(schema).not.toBeNull()
    // Per ADR, the long-term target is EmptyMutation. The current schema may
    // still expose deprecated placeholders (updateDraft, updateMemberRole,
    // removeMember); we only forbid entity CRUD here.
    const mutationFields = schema!.mutationType?.fields?.map((f) => f.name) ?? []

    const forbidden = [
      'createClass',
      'createProperty',
      'createIndividual',
      'updateClass',
      'updateProperty',
      'updateIndividual',
      'deleteClass',
      'deleteProperty',
      'deleteIndividual',
      'createOntology',
      'updateOntology',
      'deleteOntology',
      'sparqlQuery',
      'cypherQuery',
    ]

    for (const field of forbidden) {
      expect(
        mutationFields,
        `${field} must not appear on Mutation root (use REST per ADR)`,
      ).not.toContain(field)
    }
  })

  test('Mutation root is absent or contains only deprecated placeholders', () => {
    expect(schema).not.toBeNull()
    // The target end-state is no Mutation type at all (EmptyMutation). Until
    // the frontend fully migrates, only three deprecated placeholder
    // mutations are tolerated. If a new mutation appears, that is an
    // architectural violation.
    const allowedDeprecatedPlaceholders = new Set([
      'updateDraft',
      'updateMemberRole',
      'removeMember',
    ])
    const mutationFields = schema!.mutationType?.fields?.map((f) => f.name) ?? []

    if (schema!.mutationType === null) {
      // Ideal end-state — no Mutation type at all.
      return
    }

    for (const field of mutationFields) {
      if (!allowedDeprecatedPlaceholders.has(field)) {
        throw new Error(
          `Unexpected Mutation field "${field}" — only ${[...allowedDeprecatedPlaceholders].join(', ')} ` +
            'are tolerated as deprecated placeholders. New mutations violate ' +
            'ADR-DES.API.rest-graphql-mutation-boundary.',
        )
      }
    }
  })

  test('Navigation query execution works (classes)', async ({ request }) => {
    // Execute a real navigation query to confirm the read path still works.
    const resp = await fetchGraphQL(
      request,
      `query ClassesNav($ontologyId: String!) {
        classes(ontologyId: $ontologyId, page: 0, perPage: 5) {
          items { id label }
          total
        }
      }`,
      { ontologyId: '00000000-0000-0000-0000-000000000001' },
    )
    // The query must not return top-level errors (data may be empty if the
    // upstream has no data, but the schema must resolve).
    expect(resp.errors ?? []).toEqual([])
    expect(resp.data?.classes).toBeTruthy()
    expect(Array.isArray(resp.data!.classes.items)).toBeTruthy()
  })
})
