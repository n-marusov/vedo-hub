import { test, expect } from '../../fixtures';
import { OntologyWorkspacePage } from '../../../pages/ontology-workspace.page';
import { UNIVERSITY_ONTOLOGY } from '../../ontology-test-data';

// Validates: US-io.ontology.ai-completion
//
// E2E-ai.completion.class-suggestions — AI-assisted class completion (P1)
// Covers US: US-io.ontology.ai-completion
//
// Feature implemented: OntologyWorkspace renders AiSuggestionPanel (M4 Д3) —
// "Suggest subclasses" button calls POST /api/v1/ontologies/{id}/ai/suggest-classes
// (ai-orchestration-service proxy via api-gateway). Tests drive the real UI with
// route-mocked API responses (fixture pattern used by ontology-lifecycle.spec.ts).

const MOCK_CLASS_SUGGESTIONS = {
  suggestions: [
    {
      id: 'GraduateStudent',
      entityId: 'GraduateStudent',
      label: 'GraduateStudent',
      type: 'class',
      parentId: 'Student',
      parentLabel: 'Student',
      confidence: 0.92,
      rationale: 'Students pursuing advanced degrees are a common subclass of Student',
    },
    {
      id: 'UndergraduateStudent',
      entityId: 'UndergraduateStudent',
      label: 'UndergraduateStudent',
      type: 'class',
      parentId: 'Student',
      parentLabel: 'Student',
      confidence: 0.88,
      rationale: 'Undergraduate students represent the majority of a typical student body',
    },
  ],
};

test.describe('AI-Assisted Class Completion', () => {
  // The AI suggestion panel lives in the right property panel, which is
  // hidden by the responsive layout at viewport width <= 1280px.
  test.use({ viewport: { width: 1440, height: 900 } });

  let workspace: OntologyWorkspacePage;

  test.beforeEach(async ({ page }) => {
    workspace = new OntologyWorkspacePage(page);

    // Mock the workspace ontology metadata REST endpoint (fetchOntologyMeta).
    // Registered BEFORE navigation so the mount-time request is intercepted.
    await page.route('**/api/v1/ontologies/*', async (route) => {
      if (route.request().method() !== 'GET') return route.continue();
      const url = route.request().url();
      // Let entity sub-paths (classes/properties/individuals, ai/*) fall through
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

      // Unknown queries: return empty data so the workspace does not error
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ data: {} }) });
    });

    // Mock the AI suggestion API (real contract: POST /api/v1/ontologies/{id}/ai/suggest-classes)
    await page.route('**/api/v1/ontologies/*/ai/suggest-classes', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_CLASS_SUGGESTIONS),
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

  test('open class and request AI subclass suggestions with ranked list', async ({ page }) => {
    // US-io.ontology.ai-completion: Request suggestions
    await selectClassNode(page, 'Student');

    // The AiSuggestionPanel renders in the right panel once a class is selected
    const panel = page.locator('.ai-suggestions');
    await expect(panel).toBeVisible({ timeout: 10_000 });

    // Request AI suggestions via the right-panel action button
    await page.getByTestId('suggest-subclasses').click();

    // Verify suggestions panel appears with ranked list
    const suggestions = page.getByTestId('ai-suggestion-item');
    await expect(suggestions).toHaveCount(2, { timeout: 10_000 });

    // Verify confidence scores are displayed (percentages)
    const firstSuggestion = suggestions.first();
    await expect(firstSuggestion).toContainText('GraduateStudent');
    await expect(firstSuggestion).toContainText('92%');
  });

  test('accept a suggestion and verify it is marked accepted', async ({ page }) => {
    // US-io.ontology.ai-completion: Accept suggestion
    await selectClassNode(page, 'Student');
    await page.getByTestId('suggest-subclasses').click();

    const suggestions = page.getByTestId('ai-suggestion-item');
    await expect(suggestions).toHaveCount(2, { timeout: 10_000 });

    // Accept the first (highest confidence) suggestion
    await suggestions.first().getByRole('button', { name: 'Accept' }).click();

    // Verify accepted state is reflected in the UI (✓ Accepted badge + counter)
    await expect(suggestions.first().getByRole('button', { name: /accepted/i })).toBeVisible();
    await expect(page.getByTestId('suggest-subclasses')).not.toBeVisible();
  });

  test('dismiss a suggestion removes it from the list', async ({ page }) => {
    // US-io.ontology.ai-completion: Dismiss (reject) suggestion
    await selectClassNode(page, 'Student');
    await page.getByTestId('suggest-subclasses').click();

    const suggestions = page.getByTestId('ai-suggestion-item');
    await expect(suggestions).toHaveCount(2, { timeout: 10_000 });

    // Dismiss the first suggestion
    await suggestions.first().getByRole('button', { name: 'Dismiss' }).click();

    // Verify one suggestion remains
    await expect(page.getByTestId('ai-suggestion-item')).toHaveCount(1);
    await expect(page.getByTestId('ai-suggestion-item')).toContainText('UndergraduateStudent');
  });
});
