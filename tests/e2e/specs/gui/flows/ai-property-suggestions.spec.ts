import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';

// E2E-ai.completion.property-suggestions — AI-assisted property suggestions (P1)
// Covers US: US-io.ontology.ai-property-suggestions
//
// Tests:
// - Open class → request property suggestions
//   → verify domain/range hints shown
// - Accept a property suggestion → verify property created with domain/range
// - Verify datatype vs object property distinction in suggestions

// @skip — backend/API for AI-assisted property completion exists from M2 planning and
// ai-orchestration-service, but the current frontend workspace has no visible
// property-suggestion entry point/panel (no `ai-suggestion-item`, domain/range hint UI,
// or “suggest properties” control found in `src/services/frontend/src`).
// AI-agent note: do not mark this as backend-missing. Wire the existing
// ai-orchestration completion API into OntologyWorkspace/Class detail UI, then unskip.
test.describe.skip('AI-Assisted Property Suggestions', () => {
  let uploadPage: DocumentUploadPage;

  const MOCK_PROPERTY_SUGGESTIONS = {
    suggestions: [
      {
        entityId: 'advisor',
        label: 'advisor',
        propertyType: 'OBJECT_PROPERTY',
        domainId: 'Student',
        rangeId: 'Professor',
        confidence: 0.91,
        rationale: 'Students are commonly advised by professors',
      },
      {
        entityId: 'gpa',
        label: 'gpa',
        propertyType: 'DATATYPE_PROPERTY',
        domainId: 'Student',
        rangeId: 'decimal',
        confidence: 0.85,
        rationale: 'Students typically have a grade point average',
      },
      {
        entityId: 'enrollmentDate',
        label: 'enrollmentDate',
        propertyType: 'DATATYPE_PROPERTY',
        domainId: 'Student',
        rangeId: 'date',
        confidence: 0.65,
        rationale: 'Student enrollment date is useful for cohort analysis',
      },
    ],
    model: 'gpt-4',
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);

    await page.route('**/api/v1/ai/complete/property/**', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_PROPERTY_SUGGESTIONS),
        });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/ontologies/**/properties', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({ id: 'advisor', label: 'advisor', created: true }),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('open class and request property suggestions with domain/range hints', async () => {
    // US-io.ontology.ai-property-suggestions: Request property suggestions
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.page.locator('.class-tree-item', { hasText: 'Student' }).click();

    // Request property suggestions
    await uploadPage.page.getByRole('button', { name: /suggest properties|ai property suggestions/i }).click();

    // Verify suggestions panel appears
    const suggestions = uploadPage.page.locator('.ai-suggestion-item, [data-testid="suggestion-item"]');
    const count = await suggestions.count();
    expect(count).toBe(3);

    // Verify domain/range hints are shown
    const firstSuggestion = suggestions.first();
    await expect(firstSuggestion).toContainText('Student');
    await expect(firstSuggestion).toContainText('Professor');
  });

  test('accept a property suggestion and verify property created with domain/range', async () => {
    // US-io.ontology.ai-property-suggestions: Accept object property
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.page.locator('.class-tree-item', { hasText: 'Student' }).click();
    await uploadPage.page.getByRole('button', { name: /suggest properties|ai property suggestions/i }).click();

    // Accept the object property suggestion (advisor)
    await uploadPage.page.locator('.ai-suggestion-item').first()
      .locator('button:has-text("accept"), button:has-text("Apply")').click();

    // Verify success feedback
    const success = uploadPage.page.locator('.suggestion-accepted, .toast-success');
    await expect(success).toBeVisible({ timeout: 5_000 });
  });

  test('verify datatype vs object property distinction in suggestions', async () => {
    // US-io.ontology.ai-property-suggestions: Type distinction
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.page.locator('.class-tree-item', { hasText: 'Student' }).click();
    await uploadPage.page.getByRole('button', { name: /suggest properties|ai property suggestions/i }).click();

    // Verify property types are displayed
    const suggestions = uploadPage.page.locator('.ai-suggestion-item, [data-testid="suggestion-item"]');

    // First suggestion should be an object property (advisor → Student → Professor)
    const firstType = await suggestions.first().locator('.property-type-badge, [data-testid="property-type"]');
    await expect(firstType).toContainText('object');

    // Second suggestion should be a datatype property (gpa → decimal)
    const secondType = await suggestions.nth(1).locator('.property-type-badge, [data-testid="property-type"]');
    await expect(secondType).toContainText('datatype');
  });
});
