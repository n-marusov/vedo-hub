// @ctx: M2.5 API Gateway integration tests — full endpoint coverage via page.request
// @hlv:artifact tests-api-gateway validates REQ-FUN.PROCESS.e2e-testing
// Covers: 20+ REST + GraphQL endpoints + auth/error scenarios
import { test, expect } from '@playwright/test'

const BASE = 'http://localhost:3000/api/v1'

test.describe('M2.5 API Gateway Integration', () => {
  test.describe('REST — Ontologies', () => {
    test('GET /api/v1/ontologies — should list ontologies', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(Array.isArray(body)).toBeTruthy()
    })

    test('POST /api/v1/ontologies — should create a new ontology', async ({ page }) => {
      const res = await page.request.post(`${BASE}/ontologies`, {
        data: { name: 'Test Ontology', visibility: 'private' }
      })
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body).toHaveProperty('id')
      expect(body).toHaveProperty('name', 'Test Ontology')
    })
  })

  test.describe('REST — Groups', () => {
    test('GET /api/v1/groups — should list groups', async ({ page }) => {
      const res = await page.request.get(`${BASE}/groups`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(Array.isArray(body)).toBeTruthy()
    })
  })

  test.describe('REST — Members', () => {
    test('GET /api/v1/ontologies/{id}/members — should list members', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies/ont-123/members`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(Array.isArray(body)).toBeTruthy()
    })

    test('PUT /api/v1/ontologies/{id}/members/{uid} — should update member role', async ({ page }) => {
      const res = await page.request.put(`${BASE}/ontologies/ont-123/members/user-456`, {
        data: { role: 'editor' }
      })
      expect(res.ok()).toBeTruthy()
    })

    test('DELETE /api/v1/ontologies/{id}/members/{uid} — should remove member', async ({ page }) => {
      const res = await page.request.delete(`${BASE}/ontologies/ont-123/members/user-456`)
      expect(res.ok()).toBeTruthy()
    })
  })

  test.describe('REST — Versioning', () => {
    test('GET /api/v1/versioning/{id}/tags — should list tags', async ({ page }) => {
      const res = await page.request.get(`${BASE}/versioning/ont-123/tags`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(Array.isArray(body)).toBeTruthy()
    })

    test('POST /api/v1/versioning/{id}/compare — should compare revisions', async ({ page }) => {
      const res = await page.request.post(`${BASE}/versioning/ont-123/compare`, {
        data: { fromRevision: 'abc123', toRevision: 'def456' }
      })
      expect(res.ok()).toBeTruthy()
    })
  })

  test.describe('REST — Metrics', () => {
    test('GET /api/v1/metrics/{id} — should return metrics + trends', async ({ page }) => {
      const res = await page.request.get(`${BASE}/metrics/ont-123`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body).toHaveProperty('counters')
      expect(body).toHaveProperty('trends')
    })
  })

  test.describe('REST — Validation', () => {
    test('POST /api/v1/validation/{id}/run — should run validation', async ({ page }) => {
      const res = await page.request.post(`${BASE}/validation/ont-123/run`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body).toHaveProperty('status')
      expect(body).toHaveProperty('violations')
    })
  })

  test.describe('REST — Deployments', () => {
    test('GET /api/v1/deployments — should list deployments', async ({ page }) => {
      const res = await page.request.get(`${BASE}/deployments`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(Array.isArray(body)).toBeTruthy()
    })
  })

  test.describe('REST — Merge Requests', () => {
    test('GET /api/v1/merge-requests — should list merge requests', async ({ page }) => {
      const res = await page.request.get(`${BASE}/merge-requests`)
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(Array.isArray(body)).toBeTruthy()
    })
  })

  test.describe('REST — Health', () => {
    test('GET /api/v1/health — should return healthy', async ({ page }) => {
      const res = await page.request.get(`${BASE}/health`)
      expect(res.ok()).toBeTruthy()
    })

    test('GET /api/v1/ready — should return ready', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ready`)
      expect(res.ok()).toBeTruthy()
    })
  })

  test.describe('GraphQL', () => {
    test('dashboard aggregate query — should return dashboard data', async ({ page }) => {
      const res = await page.request.post(`${BASE}/graphql`, {
        data: {
          query: `
            query DashboardAggregate {
              dashboard {
                widgets { title count }
                recentOntologies { id name }
                activityFeed { text timestamp }
              }
            }
          `
        }
      })
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body.data).toHaveProperty('dashboard')
    })

    test('versioning graph neighborhood query — should return graph data', async ({ page }) => {
      const res = await page.request.post(`${BASE}/graphql`, {
        data: {
          query: `
            query GraphNeighborhood($ontologyId: ID!) {
              graphNeighborhood(ontologyId: $ontologyId, depth: 2) {
                nodes { id label }
                edges { sourceId targetId }
              }
            }
          `,
          variables: { ontologyId: 'ont-123' }
        }
      })
      expect(res.ok()).toBeTruthy()
      const body = await res.json()
      expect(body.data).toHaveProperty('graphNeighborhood')
    })

    test('save draft mutation — should save draft changes', async ({ page }) => {
      const res = await page.request.post(`${BASE}/graphql`, {
        data: {
          query: `
            mutation UpdateDraft($ontologyId: ID!, $changes: DraftInput!) {
              updateDraft(ontologyId: $ontologyId, changes: $changes) {
                success
                timestamp
              }
            }
          `,
          variables: {
            ontologyId: 'ont-123',
            changes: { fields: [{ field: 'class:1', oldValue: null, newValue: { label: 'Test' } }] }
          }
        }
      })
      expect(res.ok()).toBeTruthy()
    })
  })

  test.describe('Auth / Error Handling', () => {
    test('401 — should reject unauthenticated requests', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies`, {
        headers: { Authorization: '' }
      })
      expect(res.status()).toBe(401)
    })

    test('401 — should reject expired JWT', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies`, {
        headers: { Authorization: 'Bearer expired.jwt.token' }
      })
      expect(res.status()).toBe(401)
    })

    test('403 — should reject read-only user from writes', async ({ page }) => {
      const res = await page.request.post(`${BASE}/ontologies`, {
        data: { name: 'Test', visibility: 'private' },
        headers: { Authorization: 'Bearer viewer.token' }
      })
      expect(res.status()).toBe(403)
    })

    test('400 — should reject malformed JSON', async ({ page }) => {
      const res = await page.request.post(`${BASE}/ontologies`, {
        data: 'not-json',
        headers: { 'Content-Type': 'application/json' }
      })
      expect(res.status()).toBe(400)
    })

    test('404 — should return not found for non-existent resource', async ({ page }) => {
      const res = await page.request.get(`${BASE}/ontologies/nonexistent-id`)
      expect(res.status()).toBe(404)
    })

    test('429 — should handle rate limit', async ({ page }) => {
      let rateLimited = false
      for (let i = 0; i < 100; i++) {
        const res = await page.request.get(`${BASE}/health`)
        if (res.status() === 429) {
          rateLimited = true
          break
        }
      }
      expect(rateLimited).toBeTruthy()
    })
  })
})
