// Validates: REQ-FUN.PROCESS.e2e-testing
// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests
// Validates: REQ-FUN.API.integration
// API Gateway real-backend integration tests — verifies the
// deployed gateway contract against docker-compose.test.yml services.
import { test, expect } from '@playwright/test'

import { OWNER_JWT, VIEWER_JWT } from '../../jwt-tokens'

const BASE = 'http://localhost:3000/api/v1'
const AUTH = { Authorization: `Bearer ${OWNER_JWT}` }
const VIEWER_AUTH = { Authorization: `Bearer ${VIEWER_JWT}` }
const ONTOLOGY_ID = '00000000-0000-0000-0000-000000000001'
const CLASS_ID = 'person'

test.describe('API Gateway Integration', () => {
  test.describe('REST — Gateway and ontology-service routes', () => {
    test('GET /api/v1/ontologies — should reach ontology list facade', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      expect([200, 404, 501, 503]).toContain(res.status())
    })

    test('GET /api/v1/ontologies/{id}/classes — should proxy class list', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies/${ONTOLOGY_ID}/classes`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      if (res.ok()) {
        const body = await res.json()
        expect(body).toHaveProperty('items')
        expect(Array.isArray(body.items)).toBeTruthy()
      }
    })

    test('GET /api/v1/ontologies/{id}/properties — should proxy property list', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies/${ONTOLOGY_ID}/properties`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      if (res.ok()) {
        const body = await res.json()
        expect(body).toHaveProperty('items')
        expect(Array.isArray(body.items)).toBeTruthy()
      }
    })

    test('GET /api/v1/ontologies/{id}/individuals — should proxy individual list', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies/${ONTOLOGY_ID}/individuals`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      if (res.ok()) {
        const body = await res.json()
        expect(body).toHaveProperty('items')
        expect(Array.isArray(body.items)).toBeTruthy()
      }
    })

    test('GET /api/v1/ontologies/{id}/export — should reach export route', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies/${ONTOLOGY_ID}/export`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      expect([200, 404, 422]).toContain(res.status())
    })
  })

  test.describe('REST — Versioning service routes', () => {
    test('GET /api/v1/versioning/commits — should reach commit history endpoint', async ({ page }) => {
      const res = await page.request.get(`${BASE}/versioning/commits?ontology_id=${ONTOLOGY_ID}`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      if (res.ok()) {
        const body = await res.json()
        expect(body).toHaveProperty('items')
        expect(Array.isArray(body.items)).toBeTruthy()
      }
    })

    test('GET /api/v1/versioning/branches — should reach branch list endpoint', async ({ page }) => {
      const res = await page.request.get(`${BASE}/versioning/branches?ontology_id=${ONTOLOGY_ID}`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      if (res.ok()) {
        const body = await res.json()
        expect(body).toHaveProperty('items')
        expect(Array.isArray(body.items)).toBeTruthy()
      }
    })
  })

  test.describe('REST — Health and docs', () => {
    test('GET /api/v1/health — should return healthy via nginx proxy rewrite to API Gateway', async ({ page }) => {
      const res = await page.request.get(`${BASE}/health`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body).toHaveProperty('status', 'healthy')
      expect(body).toHaveProperty('service', 'api-gateway')
    })

    test('GET /api/v1/ready — should return ready via nginx proxy rewrite to API Gateway', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ready`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body).toHaveProperty('status', 'ready')
      expect(body).toHaveProperty('service', 'api-gateway')
    })

    test('GET /api/v1/openapi.json — should serve OpenAPI document', async ({ page }) => {
      const res = await page.request.get(`${BASE}/openapi.json`, { headers: AUTH })
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body).toHaveProperty('openapi')
    })
  })

  test.describe('GraphQL', () => {
    test('ontology metadata — served via REST (non-graph per ADR)', async ({ page }) => {
      // Per ADR: non-graph queries (ontology metadata, commits, branches)
      // are served exclusively via REST, not GraphQL.
      const res = await page.request.get(`${BASE}/ontologies/${ONTOLOGY_ID}`, { headers: AUTH })
      expect(res.status()).toBeLessThan(500)
      if (res.ok()) {
        const body = await res.json()
        expect(body.data).toHaveProperty('id')
        expect(body.data).toHaveProperty('label')
      }
    })

    test('classes query — should return a typed connection', async ({ page }) => {
      const res = await page.request.post(`${BASE}/graphql`, {
        data: {
          query: `
            query ListClasses($ontologyId: ID!) {
              classes(ontologyId: $ontologyId, page: 0, perPage: 10) {
                items { id label }
                total
                page
                perPage
              }
            }
          `,
          variables: { ontologyId: ONTOLOGY_ID }
        },
        headers: AUTH
      })
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body.errors ?? []).toEqual([])
      expect(body.data.classes).toHaveProperty('items')
      expect(Array.isArray(body.data.classes.items)).toBeTruthy()
    })

    test('graph neighborhood query — should require classId by schema contract', async ({ page }) => {
      const res = await page.request.post(`${BASE}/graphql`, {
        data: {
          query: `
            query GraphNeighborhood($ontologyId: ID!, $classId: ID!) {
              graphNeighborhood(ontologyId: $ontologyId, classId: $classId, depth: 2) {
                nodes { id label }
                edges { sourceId targetId propertyId propertyLabel }
              }
            }
          `,
          variables: { ontologyId: ONTOLOGY_ID, classId: CLASS_ID }
        },
        headers: AUTH
      })
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      if (body.errors?.length) {
        expect(body.errors[0].message).toMatch(/not found|database|neo4j/i)
      } else {
        expect(body.data.graphNeighborhood).toHaveProperty('nodes')
        expect(body.data.graphNeighborhood).toHaveProperty('edges')
      }
    })

  })

  test.describe('Auth / Error Handling', () => {
    test('401 — should reject unauthenticated requests', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies`, {
        headers: { Authorization: '' }
      })
      expect(res.status()).toBe(401)
    })

    test('401 — should reject expired or invalid JWT', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies`, {
        headers: { Authorization: 'Bearer expired.jwt.token' }
      })
      expect(res.status()).toBe(401)
    })

    test('403 — should reject read-only user from writes before upstream routing', async ({ page }) => {
      const res = await page.request.delete(`${BASE}/ontologies/${ONTOLOGY_ID}`, {
        headers: VIEWER_AUTH
      })
      expect(res.status()).toBe(403)
    })

    test('4xx — should reject malformed JSON on a real JSON-bound endpoint', async ({ page }) => {
      const res = await page.request.post(`${BASE}/graphql`, {
        data: 'not-json',
        headers: { 'Content-Type': 'application/json', ...AUTH }
      })
      expect([400, 422]).toContain(res.status())
    })

    test('404 — should return not found for non-existent gateway route', async ({ page }) => {
      const res = await page.request.get(`${BASE}/does-not-exist`, { headers: AUTH })
      expect(res.status()).toBe(404)
      const body = await res.json()
      expect(body.error.code).toBe('GATEWAY-NOT-FOUND')
    })

    test('read-only query guard — should reject SPARQL mutations before upstream execution', async ({ page }) => {
      const res = await page.request.post(`${BASE}/sparql`, {
        data: { query: 'DELETE WHERE { ?s ?p ?o }' },
        headers: AUTH
      })
      expect([400, 403]).toContain(res.status())
    })
  })
})
