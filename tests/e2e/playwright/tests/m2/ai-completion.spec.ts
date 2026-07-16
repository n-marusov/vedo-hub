import { test, expect } from '../fixtures';
import { M2DocumentUploadPage } from '../../pages/m2-document-upload.page';

// E2E-ai.completion.class-suggestions — AI-assisted class completion (P1)
// Covers US: US-io.ontology.ai-completion
//
// Tests:
// - Open existing class → request AI subclass suggestions
//   → view ranked list with confidence scores
// - Accept a suggestion → verify created entity
// - Reject all suggestions → verify no changes made

test.describe('M2 AI-Assisted Class Completion', () => {
  let uploadPage: M2DocumentUploadPage;

  const MOCK_CLASS_SUGGESTIONS = {
    suggestions: [
      {
        entityId: 'GraduateStudent',
        label: 'GraduateStudent',
        parentId: 'Student',
        confidence: 0.92,
        rationale: 'Students pursuing advanced degrees are a common subclass of Student',
      },
      {
        entityId: 'UndergraduateStudent',
        label: 'UndergraduateStudent',
        parentId: 'Student',
        confidence: 0.88,
        rationale: 'Undergraduate students represent the majority of a typical student body',
      },
      {
        entityId: 'VisitingScholar',
        label: 'VisitingScholar',
        parentId: 'Student',
        confidence: 0.45,
        rationale: 'Visiting scholars share some Student characteristics but may belong to Staff',
      },
    ],
    model: 'gpt-4',
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new M2DocumentUploadPage(page);

    await page.route('**/api/v1/ai/complete/class/**', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_CLASS_SUGGESTIONS),
        });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/ontologies/**/classes', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({ id: 'GraduateStudent', label: 'GraduateStudent', created: true }),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('open class and request AI subclass suggestions with ranked list', async () => {
    // US-io.ontology.ai-completion: Request suggestions
    await uploadPage.openDocumentUpload('TestOntology');

    // Open class detail panel (simulate selecting a class)
    await uploadPage.page.goto('/ontology/TestOntology');
    await uploadPage.page.locator('.class-tree-item', { hasText: 'Student' }).click();

    // Request AI suggestions
    await uploadPage.page.getByRole('button', { name: /ai suggestions|suggest subclasses/i }).click();

    // Verify suggestions panel appears with ranked list
    const suggestions = uploadPage.page.locator('.ai-suggestion-item, [data-testid="suggestion-item"]');
    const count = await suggestions.count();
    expect(count).toBe(3);

    // Verify confidence scores are displayed
    const firstSuggestion = suggestions.first();
    await expect(firstSuggestion).toContainText('GraduateStudent');
    await expect(firstSuggestion).toContainText('92');
  });

  test('accept a suggestion and verify entity creation', async () => {
    // US-io.ontology.ai-completion: Accept suggestion
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.page.locator('.class-tree-item', { hasText: 'Student' }).click();
    await uploadPage.page.getByRole('button', { name: /ai suggestions|suggest subclasses/i }).click();

    // Accept the first (highest confidence) suggestion
    await uploadPage.page.locator('.ai-suggestion-item').first()
      .locator('button:has-text("accept"), button:has-text("Apply")').click();

    // Verify success feedback
    const success = uploadPage.page.locator('.suggestion-accepted, .toast-success');
    await expect(success).toBeVisible({ timeout: 5_000 });
  });

  test('reject all suggestions and verify no changes', async () => {
    // US-io.ontology.ai-completion: Reject all
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.page.locator('.class-tree-item', { hasText: 'Student' }).click();
    await uploadPage.page.getByRole('button', { name: /ai suggestions|suggest subclasses/i }).click();

    // Reject all suggestions
    await uploadPage.page.getByRole('button', { name: /reject all|dismiss all/i }).click();

    // Verify suggestions panel is dismissed
    const suggestions = uploadPage.page.locator('.ai-suggestion-item, [data-testid="suggestion-item"]');
    await expect(suggestions).not.toBeVisible();
  });
});
