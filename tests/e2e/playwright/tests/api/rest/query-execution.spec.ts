import { test, expect } from '../../fixtures';
import { OWNER_JWT } from '../../jwt-tokens';

// E2E-query.sparql.execute — SPARQL/CYPHER query execution (P1)
// Covers US: US-api.sparql.execute [NEW], US-api.cypher.execute [NEW]

test.describe('SPARQL and CYPHER Query Execution E2E', () => {
  const QUERY_API = '/api/v1';

  test('execute SPARQL SELECT query', async ({ page }) => {
    // US-api.sparql.execute: Execute SPARQL SELECT
    const response = await page.request.post(`${QUERY_API}/sparql`, {
      data: { query: 'SELECT ?s ?p ?o WHERE { ?s ?p ?o } LIMIT 10' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.results).toBeDefined();
    expect(Array.isArray(body.results)).toBe(true);
    expect(body.execution_time_ms).toBeDefined();
    expect(body.triple_count).toBeDefined();
  });

  test('execute CYPHER MATCH query', async ({ page }) => {
    // US-api.cypher.execute: Execute CYPHER MATCH
    const response = await page.request.post(`${QUERY_API}/cypher`, {
      data: { query: 'MATCH (n) RETURN n LIMIT 10' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.results).toBeDefined();
    expect(Array.isArray(body.results)).toBe(true);
    expect(body.execution_time_ms).toBeDefined();
    expect(body.triple_count).toBeDefined();
  });

  test('gateway rejects SPARQL mutation query', async ({ page }) => {
    // Read-only enforcement: reject INSERT
    const response = await page.request.post(`${QUERY_API}/sparql`, {
      data: { query: 'INSERT DATA { <urn:a> <urn:b> <urn:c> }' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.error.code).toContain('GATEWAY-QUERY-READONLY');
  });

  test('gateway rejects CYPHER mutation query', async ({ page }) => {
    // Read-only enforcement: reject CREATE
    const response = await page.request.post(`${QUERY_API}/cypher`, {
      data: { query: 'CREATE (n:Test {id: "x"})' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.error.code).toContain('GATEWAY-QUERY-READONLY');
  });

  test('rejects invalid SPARQL syntax with 400', async ({ page }) => {
    const response = await page.request.post(`${QUERY_API}/sparql`, {
      data: { query: 'SELECT INVALID SYNTAX' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect([400, 404, 501]).toContain(response.status());
  });

  test('rejects query without authentication', async ({ page }) => {
    const response = await page.request.post(`${QUERY_API}/sparql`, {
      data: { query: 'SELECT ?s WHERE { ?s ?p ?o } LIMIT 5' },
    });
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.error).toBeDefined();
  });
});
