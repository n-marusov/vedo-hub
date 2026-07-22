import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';

// E2E-nl-to-owl.generation — Natural language to OWL ontology generation (P0)
// Covers US: US-io.ontology.create-nl-to-owl
//
// Tests:
// - Enter NL text "a university with students, professors, and courses"
//   → view generated sequence preview with classes and properties
// - Edit generated steps → modify label → apply → verify entities in ontology
// - Enter ambiguous NL → verify error prompt (frontend surfaces backend error)
//
// Mocked routes (match the real API contract in src/services/frontend/src/api/ai.ts
// and src/services/frontend/src/api/extraction.ts):
//   POST /api/v1/ontologies/:id/generate-from-text — NL → AiGenerationResult
//   POST /api/v1/documents/apply — applies sequence → ApplyResult

test.describe('NL→OWL Generation', () => {
  let uploadPage: DocumentUploadPage;

  // AiGenerationResult shape per src/services/frontend/src/api/ai.ts.
  // SequenceStep fields per src/services/frontend/src/types/extraction.ts:
  //   id, operation, entityId, label, parentId?, domain?, range?, included.
  const MOCK_NL_GENERATE = {
    id: 'gen-test-001',
    ontologyId: 'TestOntology',
    steps: [
      { id: 's1', operation: 'CREATE_CLASS', entityId: 'University', label: 'University', parentId: 'owl:Thing', included: true },
      { id: 's2', operation: 'CREATE_CLASS', entityId: 'Student', label: 'Student', parentId: 'owl:Thing', included: true },
      { id: 's3', operation: 'CREATE_CLASS', entityId: 'Professor', label: 'Professor', parentId: 'owl:Thing', included: true },
      { id: 's4', operation: 'CREATE_CLASS', entityId: 'Course', label: 'Course', parentId: 'owl:Thing', included: true },
      { id: 's5', operation: 'CREATE_OBJECT_PROPERTY', entityId: 'enrolledIn', label: 'enrolledIn', domain: 'Student', range: 'Course', included: true },
      { id: 's6', operation: 'CREATE_OBJECT_PROPERTY', entityId: 'teaches', label: 'teaches', domain: 'Professor', range: 'Course', included: true },
      { id: 's7', operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'name', label: 'name', domain: 'owl:Thing', range: 'string', included: true },
    ],
    tokenUsage: { prompt: 45, completion: 128, total: 173 },
  };

  // ApplyResult shape per src/services/frontend/src/types/extraction.ts.
  const MOCK_APPLY = {
    success: true,
    appliedCount: 7,
    commitId: 'nl-owl-commit-001',
    commitUrl: '/project/TestOntology/commits/nl-owl-commit-001',
    errors: [],
    timestamp: new Date().toISOString(),
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);

    // Real endpoint: POST /api/v1/ontologies/:id/generate-from-text
    // (see src/services/frontend/src/api/ai.ts → generateFromText)
    await page.route('**/api/v1/ontologies/*/generate-from-text', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_NL_GENERATE),
        });
      } else {
        await route.continue();
      }
    });

    // Real endpoint: POST /api/v1/documents/apply
    // (see src/services/frontend/src/api/extraction.ts → applySequence)
    await page.route('**/api/v1/documents/apply', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_APPLY),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('enter NL text and view generated sequence preview with classes and properties', async () => {
    // US-io.ontology.create-nl-to-owl: Generate ontology from NL
    await uploadPage.openNLToOWLImport('TestOntology');

    // Enter natural language description
    await uploadPage.nlInput().fill('a university with students, professors, and courses');

    // Start waiting for the response BEFORE clicking (prevents race condition)
    const genResponse = uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/generate-from-text') && resp.status() === 200,
      { timeout: 15_000 }
    );
    await uploadPage.page.getByRole('button', { name: /^Generate$/i }).click();
    await genResponse;

    // Verify preview shows extracted steps
    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBeGreaterThanOrEqual(4);

    // Verify key classes are present
    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('University'))).toBe(true);
    expect(labels.some((l) => l.includes('Student'))).toBe(true);
    expect(labels.some((l) => l.includes('Professor'))).toBe(true);
    expect(labels.some((l) => l.includes('Course'))).toBe(true);

    // Verify properties are present
    expect(labels.some((l) => l.includes('enrolledIn'))).toBe(true);
    expect(labels.some((l) => l.includes('teaches'))).toBe(true);
  });

  test('edit generated steps and apply sequence', async () => {
    // US-io.ontology.create-nl-to-owl: Edit and apply
    await uploadPage.openNLToOWLImport('TestOntology');

    await uploadPage.nlInput().fill('a library system with books, authors, and members');

    const genResponse = uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/generate-from-text') && resp.status() === 200,
      { timeout: 15_000 }
    );
    await uploadPage.page.getByRole('button', { name: /^Generate$/i }).click();
    await genResponse;

    // Wait for preview rows to render
    await uploadPage.page.locator('.preview-row').first().waitFor({ state: 'visible' });

    // Edit a label
    await uploadPage.editStepLabel(0, 'Library');

    // Apply the sequence — opens the confirm modal
    await uploadPage.applySequence();

    // Verify apply succeeded (clicks confirm → waits for success title)
    const success = await uploadPage.waitForApplyComplete();
    expect(success).toBe(true);
  });

  test('ambiguous natural language shows error prompt', async () => {
    // US-io.ontology.create-nl-to-owl: Ambiguity handling
    await uploadPage.page.unroute('**/api/v1/ontologies/*/generate-from-text');
    await uploadPage.page.route('**/api/v1/ontologies/*/generate-from-text', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 300,
          contentType: 'application/json',
          body: JSON.stringify({
            error: {
              code: 'AI-AMBIGUOUS-TERMS',
              message: 'Some terms are ambiguous: course, staff. Please clarify.',
              ambiguities: [
                { term: 'course', options: ['A unit of academic study', 'A direction or path', 'A sports field'] },
                { term: 'staff', options: ['Academic staff (professors)', 'Administrative staff', 'Both'] },
              ],
            },
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openNLToOWLImport('TestOntology');

    await uploadPage.nlInput().fill('a university with courses and staff');
    await uploadPage.page.getByRole('button', { name: /^Generate$/i }).click();

    // Wait for error to appear
    const errorPrompt = uploadPage.page.locator('.nl-error');
    await expect(errorPrompt).toBeVisible({ timeout: 10_000 });

    // Verify the error message mentions ambiguity
    await expect(errorPrompt).toContainText('ambiguous');

    // Verify the retry button is available
    await expect(errorPrompt.locator('.nl-error__retry')).toBeVisible();
  });
});
