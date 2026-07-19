import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';

// E2E-nl-to-owl.generation — Natural language to OWL ontology generation (P0)
// Covers US: US-io.ontology.create-nl-to-owl
//
// Tests:
// - Enter NL text "a university with students, professors, and courses"
//   → view generated sequence preview with classes and properties
// - Edit generated steps → modify label → apply → verify entities in ontology
// - Enter ambiguous NL → verify disambiguation prompt
//
// Mocked routes:
//   POST /api/v1/ai/generate — NL → SequenceStep[]
//   POST /api/v1/ontologies/:name/apply — applies sequence

test.describe('NL→OWL Generation', () => {
  let uploadPage: DocumentUploadPage;

  const MOCK_NL_GENERATE = {
    steps: [
      { operation: 'CREATE_CLASS', entityId: 'University', label: 'University', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Student', label: 'Student', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Professor', label: 'Professor', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Course', label: 'Course', parentId: 'owl:Thing' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'enrolledIn', label: 'enrolledIn', domainId: 'Student', rangeId: 'Course' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'teaches', label: 'teaches', domainId: 'Professor', rangeId: 'Course' },
      { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'name', label: 'name', domainId: 'owl:Thing', rangeId: 'string' },
    ],
    promptTokens: 45,
    completionTokens: 128,
    model: 'gpt-4',
  };

  const MOCK_COMMIT = {
    commitId: 'nl-owl-commit-001',
    message: 'Created ontology from natural language description',
    branchId: 'main',
    entityCount: 7,
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);

    await page.route('**/api/v1/ai/generate', async (route) => {
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

    await page.route('**/api/v1/ontologies/**/apply', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_COMMIT),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('enter NL text and view generated sequence preview with classes and properties', async () => {
    // US-io.ontology.create-nl-to-owl: Generate ontology from NL
    await uploadPage.openDocumentUpload('TestOntology');

    // Enter natural language description
    const nlInput = uploadPage.page.locator('.nl-to-owl-input textarea, [data-testid="nl-input"]');
    await nlInput.fill('a university with students, professors, and courses');
    await uploadPage.page.getByRole('button', { name: /generate|create ontology/i }).click();

    // Wait for generation to complete
    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 15_000 }
    );

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
    await uploadPage.openDocumentUpload('TestOntology');

    const nlInput = uploadPage.page.locator('.nl-to-owl-input textarea, [data-testid="nl-input"]');
    await nlInput.fill('a library system with books, authors, and members');
    await uploadPage.page.getByRole('button', { name: /generate|create ontology/i }).click();

    await uploadPage.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/ai/generate') && resp.status() === 200,
      { timeout: 15_000 }
    );

    // Edit a label
    await uploadPage.editStepLabel(0, 'Library');

    // Apply the sequence
    await uploadPage.applySequence();

    // Verify apply succeeded
    const success = await uploadPage.waitForApplyComplete();
    expect(success).toBe(true);
  });

  test('ambiguous natural language shows disambiguation prompt', async () => {
    // US-io.ontology.create-nl-to-owl: Ambiguity handling
    await uploadPage.page.unroute('**/api/v1/ai/generate');
    await uploadPage.page.route('**/api/v1/ai/generate', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 300,
          contentType: 'application/json',
          body: JSON.stringify({
            ambiguities: [
              {
                term: 'course',
                options: ['A unit of academic study', 'A direction or path', 'A sports field'],
              },
              {
                term: 'staff',
                options: ['Academic staff (professors)', 'Administrative staff', 'Both'],
              },
            ],
            message: 'Some terms are ambiguous. Please clarify.',
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');

    const nlInput = uploadPage.page.locator('.nl-to-owl-input textarea, [data-testid="nl-input"]');
    await nlInput.fill('a university with courses and staff');
    await uploadPage.page.getByRole('button', { name: /generate|create ontology/i }).click();

    // Verify ambiguity resolution prompt is shown
    const ambiguityPrompt = uploadPage.page.locator('.ambiguity-prompt, [data-testid="ambiguity-prompt"]');
    await expect(ambiguityPrompt).toBeVisible({ timeout: 10_000 });

    // Verify ambiguous terms are listed
    await expect(ambiguityPrompt).toContainText('course');
    await expect(ambiguityPrompt).toContainText('staff');
  });
});
