import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';
import { UNIVERSITY_ONTOLOGY } from '../../ontology-test-data';

// Validates: US-io.ontology.iterative-refinement
// Validates: REQ-FUN.API.iterative-refinement-context
//
// E2E-ai.refinement.iterative — Iterative refinement of generated ontologies (P1)
// Covers US: US-io.ontology.iterative-refinement
//
// Feature implemented: OntologyWorkspace NL→OWL tab (M4 Д1-Д2) —
//   generateFromText → POST /api/v1/ontologies/{id}/generate-from-text
//   refineSequence   → POST /api/v1/ontologies/{id}/ai/refine
// (ai-orchestration-service proxy via api-gateway). Tests drive the real UI with
// route-mocked API responses (fixture pattern used by ontology-lifecycle.spec.ts).

const MOCK_INITIAL_GENERATE = {
  id: 'gen-001',
  ontologyId: UNIVERSITY_ONTOLOGY.name,
  steps: [
    { id: 's1', operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing', parentLabel: 'owl:Thing', included: true },
    { id: 's2', operation: 'CREATE_CLASS', entityId: 'Event', label: 'Event', parentId: 'owl:Thing', parentLabel: 'owl:Thing', included: true },
  ],
};

const MOCK_REFINED = {
  id: 'gen-001',
  ontologyId: UNIVERSITY_ONTOLOGY.name,
  steps: [
    { id: 's1', operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing', parentLabel: 'owl:Thing', included: true },
    { id: 's2', operation: 'CREATE_CLASS', entityId: 'Event', label: 'Event', parentId: 'owl:Thing', parentLabel: 'owl:Thing', included: true },
    { id: 's3', operation: 'CREATE_CLASS', entityId: 'Location', label: 'Location', parentId: 'owl:Thing', parentLabel: 'owl:Thing', included: true },
    { id: 's4', operation: 'CREATE_OBJECT_PROPERTY', entityId: 'locatedAt', label: 'locatedAt', domain: 'Event', range: 'Location', included: true },
  ],
  round: 1,
  maxRounds: 3,
};

test.describe('Iterative Refinement', () => {
  let uploadPage: DocumentUploadPage;

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);

    // Mock the workspace ontology metadata REST endpoint (fetchOntologyMeta).
    await page.route('**/api/v1/ontologies/*', async (route) => {
      if (route.request().method() !== 'GET') return route.continue();
      const url = route.request().url();
      // Let entity sub-paths (classes/properties/individuals, ai/*, generate) fall through
      if (/\/classes|\/properties|\/individuals|\/validate|\/ai\//.test(url)) return route.continue();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: UNIVERSITY_ONTOLOGY.name,
          name: UNIVERSITY_ONTOLOGY.name,
          label: UNIVERSITY_ONTOLOGY.name,
          branch: 'main',
          description: 'Academic ontology',
          class_count: UNIVERSITY_ONTOLOGY.classes.length,
          property_count: UNIVERSITY_ONTOLOGY.properties.length,
          individual_count: UNIVERSITY_ONTOLOGY.individuals.length,
        }),
      });
    });

    // Mock GraphQL: ClassTree + Ontology queries (Apollo) so the workspace mounts cleanly
    await page.route('**/api/v1/graphql', async (route) => {
      const body = JSON.parse(route.request().postData() || '{}');
      const op = body.operationName;

      if (op === 'Ontology') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              ontology: {
                id: UNIVERSITY_ONTOLOGY.name,
                name: UNIVERSITY_ONTOLOGY.name,
                branch: 'main',
                commit: 'abc123',
                dirty: false,
              },
            },
          }),
        });
        return;
      }

      if (op === 'ClassTree') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: { classTree: [] } }),
        });
        return;
      }

      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ data: {} }) });
    });

    // Mock NL→OWL generation API (real contract: POST /api/v1/ontologies/{id}/generate-from-text)
    await page.route('**/api/v1/ontologies/*/generate-from-text', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_INITIAL_GENERATE),
        });
      } else {
        await route.continue();
      }
    });

    // Mock refinement API (real contract: POST /api/v1/ontologies/{id}/ai/refine)
    await page.route('**/api/v1/ontologies/*/ai/refine', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_REFINED),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('generate initial ontology then refine with feedback', async ({ page }) => {
    // US-io.ontology.iterative-refinement: Initial generation + refinement
    await uploadPage.openNLToOWLImport(UNIVERSITY_ONTOLOGY.name);

    // Initial generation
    await uploadPage.nlInput().fill('a system with people and events');
    await page.getByTestId('nl-to-owl-generate').click();

    // Verify initial steps appear in the preview
    let labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Person'))).toBe(true);
    expect(labels.some((l) => l.includes('Event'))).toBe(true);

    // Provide feedback via the refinement input
    const feedbackInput = page.getByTestId('refinement-input');
    await feedbackInput.fill('add a Location class with a locatedAt property from Event to Location');
    await page.getByTestId('refinement-submit').click();

    // Wait for the refined steps to render before reading the preview labels
    await expect(page.locator('.preview-row__label-text').filter({ hasText: 'Location' })).toBeVisible({
      timeout: 10_000,
    });

    // Verify refined steps include Location + locatedAt
    labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Location'))).toBe(true);
    expect(labels.some((l) => l.includes('locatedAt'))).toBe(true);
  });

  test('refinement preserves previously generated elements', async ({ page }) => {
    // REQ-FUN.API.iterative-refinement-context: context preservation
    await uploadPage.openNLToOWLImport(UNIVERSITY_ONTOLOGY.name);

    await uploadPage.nlInput().fill('a system with people and events');
    await page.getByTestId('nl-to-owl-generate').click();

    // Initial steps
    let labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Person'))).toBe(true);

    // Refine
    await page.getByTestId('refinement-input').fill('add a Location class');
    await page.getByTestId('refinement-submit').click();

    // Wait for the refined steps to render
    await expect(page.locator('.preview-row__label-text').filter({ hasText: 'Location' })).toBeVisible({
      timeout: 10_000,
    });

    // All prior elements preserved + new ones added
    labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Person'))).toBe(true);
    expect(labels.some((l) => l.includes('Event'))).toBe(true);
    expect(labels.some((l) => l.includes('Location'))).toBe(true);
  });

  test('refinement round indicator is shown after refinement', async ({ page }) => {
    // REQ-FUN.API.iterative-refinement-context: round display
    await uploadPage.openNLToOWLImport(UNIVERSITY_ONTOLOGY.name);

    await uploadPage.nlInput().fill('a system with people and events');
    await page.getByTestId('nl-to-owl-generate').click();

    // Refine once
    await page.getByTestId('refinement-input').fill('add Location');
    await page.getByTestId('refinement-submit').click();

    // Round indicator should show "Refinement round 1 / 3" (REQ-FUN.API.max-refinement-iterations: ≤ 3)
    await expect(page.locator('.nl-refinement__round')).toContainText('1 / 3', { timeout: 10_000 });
  });
});
