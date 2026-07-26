import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';

// E2E-ai.refinement.iterative — Iterative refinement of generated ontologies (P1)
// Covers US: US-io.ontology.iterative-refinement
//
// Tests:
// - Generate initial ontology → provide feedback "add a Location class"
//   → verify refined sequence includes Location
// - Multiple refinement rounds → verify sequence accumulates changes
// - Apply final refined version → verify all entities created

// @skip — iterative refinement backend/API exists in api-gateway and M2 plans, but
// the current frontend AI import panel only exposes document upload/preview/apply.
// No NL prompt textarea/refinement feedback controls (`nl-to-owl-input`,
// `refinement-input`) are wired in `src/services/frontend/src`.
// AI-agent note: do not mark this as backend-missing. Wire the existing refinement
// endpoint into the workspace UI and update routes to `/api/v1/documents/*` or the
// actual refine endpoint before unskipping.
test.describe.skip('Iterative Refinement', () => {
  let uploadPage: DocumentUploadPage;
  let refinementRound = 0;

  const MOCK_INITIAL_GENERATE = {
    steps: [
      { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Event', label: 'Event', parentId: 'owl:Thing' },
    ],
    promptTokens: 30,
    completionTokens: 60,
  };

  const MOCK_REFINED = {
    steps: [
      { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Event', label: 'Event', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Location', label: 'Location', parentId: 'owl:Thing' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'locatedAt', label: 'locatedAt', domainId: 'Event', rangeId: 'Location' },
    ],
    promptTokens: 45,
    completionTokens: 80,
  };

  const MOCK_DOUBLE_REFINED = {
    steps: [
      { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Event', label: 'Event', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Location', label: 'Location', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Organization', label: 'Organization', parentId: 'owl:Thing' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'locatedAt', label: 'locatedAt', domainId: 'Event', rangeId: 'Location' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'organizedBy', label: 'organizedBy', domainId: 'Event', rangeId: 'Organization' },
    ],
    promptTokens: 55,
    completionTokens: 110,
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);
    refinementRound = 0;

    // Progressive mock: first call returns initial, subsequent calls return refined
    await page.route('**/api/v1/ai/generate', async (route) => {
      if (route.request().method() === 'POST') {
        const body = JSON.parse(route.request().postData() || '{}');

        let response;
        if (body.feedback) {
          // Refinement call
          refinementRound++;
          if (refinementRound >= 2) {
            response = MOCK_DOUBLE_REFINED;
          } else {
            response = MOCK_REFINED;
          }
        } else {
          // Initial generation
          response = MOCK_INITIAL_GENERATE;
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(response),
        });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/ontologies/**/apply', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            commitId: 'refine-commit-001',
            message: 'Applied refined ontology after iteration',
            branchId: 'main',
            entityCount: 6,
          }),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('generate initial ontology then refine with feedback', async () => {
    // US-io.ontology.iterative-refinement: Initial generation + refinement
    await uploadPage.openDocumentUpload('TestOntology');

    // Initial generation
    const nlInput = uploadPage.page.locator('.nl-to-owl-input textarea, [data-testid="nl-input"]');
    await nlInput.fill('a system with people and events');
    await uploadPage.page.getByRole('button', { name: /generate|create ontology/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 10_000 }
    );

    // Verify initial steps
    let labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Person'))).toBe(true);
    expect(labels.some((l) => l.includes('Event'))).toBe(true);

    // Provide feedback
    const feedbackInput = uploadPage.page.locator('.refinement-input textarea, [data-testid="feedback-input"]');
    await feedbackInput.fill('add a Location class with a locatedAt property from Event to Location');
    await uploadPage.page.getByRole('button', { name: /refine|update/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 10_000 }
    );

    // Verify refined steps include Location
    labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Location'))).toBe(true);
    expect(labels.some((l) => l.includes('locatedAt'))).toBe(true);
  });

  test('multiple refinement rounds accumulate changes', async () => {
    // US-io.ontology.iterative-refinement: Multiple rounds
    await uploadPage.openDocumentUpload('TestOntology');

    // Initial generation
    const nlInput = uploadPage.page.locator('.nl-to-owl-input textarea, [data-testid="nl-input"]');
    await nlInput.fill('a system with people and events');
    await uploadPage.page.getByRole('button', { name: /generate|create ontology/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 10_000 }
    );

    // First refinement: add Location
    const feedbackInput1 = uploadPage.page.locator('.refinement-input textarea, [data-testid="feedback-input"]');
    await feedbackInput1.fill('add a Location class');
    await uploadPage.page.getByRole('button', { name: /refine|update/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 10_000 }
    );

    // Second refinement: add Organization
    const feedbackInput2 = uploadPage.page.locator('.refinement-input textarea, [data-testid="feedback-input"]');
    await feedbackInput2.fill('add an Organization class and organizedBy property');
    await uploadPage.page.getByRole('button', { name: /refine|update/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 10_000 }
    );

    // Verify final sequence has all accumulated entities
    const finalLabels = await uploadPage.getStepLabels();
    expect(finalLabels.some((l) => l.includes('Person'))).toBe(true);
    expect(finalLabels.some((l) => l.includes('Event'))).toBe(true);
    expect(finalLabels.some((l) => l.includes('Location'))).toBe(true);
    expect(finalLabels.some((l) => l.includes('Organization'))).toBe(true);

    // Verify accumulated properties
    expect(finalLabels.some((l) => l.includes('locatedAt'))).toBe(true);
    expect(finalLabels.some((l) => l.includes('organizedBy'))).toBe(true);
  });

  test('apply final refined version creates all entities', async () => {
    // US-io.ontology.iterative-refinement: Apply final
    await uploadPage.openDocumentUpload('TestOntology');

    const nlInput = uploadPage.page.locator('.nl-to-owl-input textarea, [data-testid="nl-input"]');
    await nlInput.fill('a system with people and events');
    await uploadPage.page.getByRole('button', { name: /generate|create ontology/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 10_000 }
    );

    // One refinement round
    const feedbackInput = uploadPage.page.locator('.refinement-input textarea, [data-testid="feedback-input"]');
    await feedbackInput.fill('add Location and organizedBy');
    await uploadPage.page.getByRole('button', { name: /refine|update/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 10_000 }
    );

    await uploadPage.applySequence();

    const success = await uploadPage.waitForApplyComplete();
    expect(success).toBe(true);

    // Verify entity count in commit message
    const successMsg = await uploadPage.getSuccessMessage();
    expect(successMsg).not.toBeNull();
  });
});
