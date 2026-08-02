// Validates: US-api.docs.openapi
// Validates: US-api.ontologies.read-rest
import { test, expect } from '@playwright/test';

// @skip — API docs page (/docs Swagger UI) not implemented in MVP frontend; OpenAPI
// document is served by the gateway (covered by active api-gateway-full API E2E tests).
// REST ontology read covered by active api-integration / org-api API E2E tests.
// Backlog: M6 (OpenAPI documentation) — gateway-level, already verified via API suite.
test.describe.skip('API — REST API documentation and ontology access', () => {
  test('US-api.docs.openapi: view and interact with OpenAPI docs', async ({ page }) => {
    // TODO: Navigate to /docs → verify Swagger UI loads → try an endpoint
  });

  test('US-api.ontologies.read-rest: read ontologies via REST API', async ({ page }) => {
    // TODO: Call GET /api/v1/ontologies → verify JSON response contains expected data
  });
});
