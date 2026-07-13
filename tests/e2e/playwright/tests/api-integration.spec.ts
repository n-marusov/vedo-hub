import { test, expect } from './fixtures';

// E2E-api.integration.rest — REST API integration (P1)
// Covers US: US-api.classes.create-rest, US-api.auth.jwt, US-api.webhooks.manage,
//            US-io.export.canonical-turtle

test.describe('REST API Integration E2E', () => {
  const API_BASE = '/api/v1';

  test('create ontology and class via REST API', async ({ page }) => {
    // US-api.classes.create-rest: Create class via REST API
    const response = await page.request.post(`${API_BASE}/ontologies`, {
      data: { name: 'ProductCatalog' },
      headers: { Authorization: 'Bearer test-jwt-token' },
    });
    expect(response.status()).toBe(201);

    // Create class via REST
    const classResponse = await page.request.post(
      `${API_BASE}/ontologies/ProductCatalog/classes`,
      {
        data: { label: 'Product', parents: ['owl:Thing'] },
        headers: { Authorization: 'Bearer test-jwt-token' },
      }
    );
    expect(classResponse.status()).toBe(201);
    const classBody = await classResponse.json();
    expect(classBody.id).toBe('Product');
  });

  test('REST API rejects request without authentication', async ({ page }) => {
    // US-api.auth.jwt: Authentication required
    const response = await page.request.post(`${API_BASE}/ontologies`, {
      data: { name: 'Test' },
    });
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.error).toBeDefined();
  });

  test('export ontology in Turtle format', async ({ page }) => {
    // US-io.export.canonical-turtle: Export canonical Turtle
    // Setup: create ontology first
    const createResponse = await page.request.post(`${API_BASE}/ontologies`, {
      data: { name: 'ExportTest' },
      headers: { Authorization: 'Bearer test-jwt-token' },
    });
    expect(createResponse.status()).toBe(201);

    // Export as Turtle
    const exportResponse = await page.request.get(
      `${API_BASE}/ontologies/ExportTest/export?format=turtle`,
      { headers: { Authorization: 'Bearer test-jwt-token' } }
    );
    expect(exportResponse.status()).toBe(200);
    const contentType = exportResponse.headers()['content-type'];
    expect(contentType).toContain('text/turtle');
  });

  test('webhook subscription and event delivery', async ({ page }) => {
    // US-api.webhooks.manage: Webhook management
    const hookResponse = await page.request.post(`${API_BASE}/webhooks`, {
      data: {
        url: 'https://webhook.site/vedo-events',
        events: ['class.created', 'class.updated'],
      },
      headers: { Authorization: 'Bearer test-jwt-token' },
    });
    expect(hookResponse.status()).toBe(201);

    // Create class to trigger webhook
    const classResponse = await page.request.post(
      `${API_BASE}/ontologies/WebhookTest/classes`,
      {
        data: { label: 'WebhookClass' },
        headers: { Authorization: 'Bearer test-jwt-token' },
      }
    );
    expect(classResponse.status()).toBe(201);
    // NOTE: Webhook delivery verification requires external mock server
  });
});
