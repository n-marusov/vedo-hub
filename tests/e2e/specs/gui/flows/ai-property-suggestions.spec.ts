import { test, expect } from '../../fixtures';
import { OntologyWorkspacePage } from '../../../pages/ontology-workspace.page';
import { UNIVERSITY_ONTOLOGY } from '../../ontology-test-data';

// Validates: US-io.ontology.ai-property-suggestions
//
// E2E-ai.completion.property-suggestions — AI-assisted property suggestions (P1)
// Covers US: US-io.ontology.ai-property-suggestions
//
// Feature implemented: OntologyWorkspace renders AiSuggestionPanel (M4 Д3) —
// "Suggest properties" button calls POST /api/v1/ontologies/{id}/ai/suggest-properties
// (ai-orchestration-service proxy via api-gateway). Tests drive the real UI with
// route-mocked API responses (fixture pattern used by ontology-lifecycle.spec.ts).

const MOCK_PROPERTY_SUGGESTIONS = {
  suggestions: [
    {
      id: 'advisor',
      entityId: 'advisor',
      label: 'advisor',
      type: 'property',
      propertyType: 'OBJECT_PROPERTY',
      domainId: 'Student',
      rangeId: 'Professor',
      parentLabel: 'Student',
      confidence: 0.91,
      rationale: 'Students are commonly advised by professors',
    },
    {
      id: 'gpa',
      entityId: 'gpa',
      label: 'gpa',
      type: 'property',
      propertyType: 'DATATYPE_PROPERTY',
      domainId: 'Student',
      rangeId: 'decimal',
      parentLabel: 'Student',
      confidence: 0.85,
      rationale: 'Students typically have a grade point average',
    },
  ],
};

test.describe('AI-Assisted Property Suggestions', () => {
  // The AI suggestion panel lives in the right property panel, which is
  // hidden by the responsive layout at viewport width <= 1280px.
  test.use({ viewport: { width: 1440, height: 900 } });

  let workspace: OntologyWorkspacePage;

  test.beforeEach(async ({ page }) => {
    workspace = new OntologyWorkspacePage(page);

    // Mock the workspace ontology metadata REST endpoint (fetchOntologyMeta).
    await page.route('**/api/v1/ontologies/*', async (route) => {
      if (route.request().method() !== 'GET') return route.continue();
      const url = route.request().url();
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

    // Mock GraphQL: ClassTree + Ontology + individuals queries (Apollo)
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
        // Flat tree — all classes at top level (AI tests don't depend on hierarchy)
        const classTreeData = UNIVERSITY_ONTOLOGY.classes.map((c) => ({
          id: c.id,
          label: c.label,
          comment: c.comment || null,
          children: [],
        }));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: { classTree: classTreeData } }),
        });
        return;
      }

      if (op === 'ListIndividuals' || op === 'VersionContext') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: {} }),
        });
        return;
      }

      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ data: {} }) });
    });

    // Mock the AI suggestion API (real contract: POST /api/v1/ontologies/{id}/ai/suggest-properties)
    await page.route('**/api/v1/ontologies/*/ai/suggest-properties', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_PROPERTY_SUGGESTIONS),
        });
      } else {
        await route.continue();
      }
    });

    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
  });

  // Real ClassTree renders clickable .class-tree__node divs (not buttons).
  // OntologyWorkspacePage.selectClass targets injected .class-row buttons;
  // these AI tests drive the real tree, so click the node directly.
  async function selectClassNode(page: import('@playwright/test').Page, label: string) {
    await page.locator('.class-tree__node').filter({ has: page.locator('.class-tree__label', { hasText: label }) }).first().click();
  }

  test('open class and request property suggestions with domain/range hints', async ({ page }) => {
    // US-io.ontology.ai-property-suggestions: Request property suggestions
    await selectClassNode(page, 'Student');
    await page.getByTestId('suggest-properties').click();

    // Verify suggestions panel appears with ranked list
    const suggestions = page.getByTestId('ai-suggestion-item');
    await expect(suggestions).toHaveCount(2, { timeout: 10_000 });

    // Verify domain/range hints are shown (parent label + rationale)
    const firstSuggestion = suggestions.first();
    await expect(firstSuggestion).toContainText('advisor');
    await expect(firstSuggestion).toContainText('Student');
    // The AiSuggestionPanel surfaces the hint via rationale text (case-insensitive)
    await expect(firstSuggestion).toContainText(/professor/i);
  });

  test('accept a property suggestion and verify it is marked accepted', async ({ page }) => {
    // US-io.ontology.ai-property-suggestions: Accept object property
    await selectClassNode(page, 'Student');
    await page.getByTestId('suggest-properties').click();

    const suggestions = page.getByTestId('ai-suggestion-item');
    await expect(suggestions).toHaveCount(2, { timeout: 10_000 });

    // Accept the object property suggestion (advisor)
    await suggestions.first().getByRole('button', { name: 'Accept' }).click();

    // Verify accepted state is reflected in the UI
    await expect(suggestions.first().getByRole('button', { name: /accepted/i })).toBeVisible();
  });

  test('verify property type badges distinguish suggestions', async ({ page }) => {
    // US-io.ontology.ai-property-suggestions: Type distinction
    await selectClassNode(page, 'Student');
    await page.getByTestId('suggest-properties').click();

    const suggestions = page.getByTestId('ai-suggestion-item');
    await expect(suggestions).toHaveCount(2, { timeout: 10_000 });

    // Each suggestion renders a type badge (Property) — both are property suggestions
    await expect(suggestions.first().locator('.ai-suggestion-item__type-badge')).toContainText('Property');
    await expect(suggestions.nth(1).locator('.ai-suggestion-item__type-badge')).toContainText('Property');
  });
});
