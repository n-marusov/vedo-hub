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
//   - Navigation queries (`class`, `classes`, `classTree`, `classAncestors`,
//     `classDescendants`, `graphNeighborhood`, `autocompleteClasses`,
//     `property`, `properties`, `individual`, `individuals` — 11 resolvers)
//     are present on the Query root.
//   - Non-graph queries (`ontology`, `commits`, `branches`) are NOT present.
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

interface GraphQLResponse<TData = { __schema: IntrospectionSchema }> {
  data?: TData
  errors?: Array<{ message: string }>
}

async function fetchGraphQL<TData = { __schema: IntrospectionSchema }>(
  request: import('@playwright/test').APIRequestContext,
  query: string,
  variables: Record<string, unknown> = {},
): Promise<GraphQLResponse<TData>> {
  const res = await request.post(`${BASE}/graphql`, {
    headers: AUTH,
    data: { query, variables },
  })
  expect(res.status()).toBeLessThan(500)
  return (await res.json()) as GraphQLResponse<TData>
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
    // Required navigation fields per revised ADR-DES.API.graphql-sparql-split-strategy
    // — 11 read-only graph navigation resolvers.
    const required = [
      'class',
      'classes',
      'classTree',
      'classAncestors',
      'classDescendants',
      'graphNeighborhood',
      'autocompleteClasses',
      'property',
      'properties',
      'individual',
      'individuals',
    ]
    for (const field of required) {
      expect(queryFields, `Query root should expose ${field}`).toContain(field)
    }
  })

  test('Query root does NOT expose non-graph fields (ontology, commits, branches)', () => {
    expect(schema).not.toBeNull()
    const queryFields = schema!.queryType.fields?.map((f) => f.name) ?? []
    // Non-graph queries are forbidden per ADR-DES.API.graphql-sparql-split-strategy §5.
    // Ontology metadata, versioning data, and org model go through REST only.
    const forbidden = ['ontology', 'commits', 'branches', 'branch', 'groups', 'projects', 'members']
    for (const field of forbidden) {
      expect(
        queryFields,
        `${field} must NOT appear on Query root (REST-only per ADR)`,
      ).not.toContain(field)
    }
  })

  test('Query root does NOT expose sparqlQuery (REST-only per ADR)', () => {
    expect(schema).not.toBeNull()
    const queryFields = schema!.queryType.fields?.map((f) => f.name) ?? []
    expect(queryFields, 'sparqlQuery must not appear on Query root').not.toContain('sparqlQuery')
    expect(queryFields, 'cypherQuery must not appear on Query root').not.toContain('cypherQuery')
  })

  test('Entity interface exists and Class implements it', async ({ request }) => {
    // Per ADR-DES.API.graphql-sparql-split-strategy: polymorphism via the
    // Entity interface implemented by Class, Property, Individual.
    const resp = await fetchGraphQL(
      request,
      `query EntityContract {
        entityType: __type(name: "Entity") { kind }
        classType: __type(name: "Class") { interfaces { name } fields { name } }
        propertyTypeEnum: __type(name: "PropertyType") { enumValues { name } }
      }`,
    )
    expect(resp.errors ?? []).toEqual([])
    const data = resp.data as unknown as {
      entityType: { kind: string }
      classType: {
        interfaces: Array<{ name: string }>
        fields: Array<{ name: string }>
      }
      propertyTypeEnum: { enumValues: Array<{ name: string }> }
    }
    expect(data.entityType.kind).toBe('INTERFACE')
    const ifaces = data.classType.interfaces.map((i) => i.name)
    expect(ifaces).toContain('Entity')
    const classFields = data.classType.fields.map((f) => f.name)
    expect(classFields).toContain('isAbstract')
    expect(classFields).toContain('isDeprecated')
    expect(classFields).toContain('entityType')
    const enumValues = data.propertyTypeEnum.enumValues.map((v) => v.name)
    expect(enumValues).toContain('ANNOTATION')
    expect(enumValues).toContain('OBJECT')
    expect(enumValues).toContain('DATATYPE')
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
    const resp = await fetchGraphQL<{ classes: { items: Array<{ id: string; label: string }>; total: number } }>(
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
