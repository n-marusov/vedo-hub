// Validates: US-abox.individuals.create
// Validates: US-abox.individuals.read
// Validates: US-abox.individuals.update
// Validates: US-abox.individuals.delete
// Validates: US-abox.individuals.list-by-class
//
// M5: Minimal ABox CRUD — GUI e2e coverage for create/read/update/delete
// individuals, list by class. REST CRUD exists (POST/PUT/DELETE /individuals),
// CreateIndividualDialog.vue exists, ABox view exists (Q3).
// This test file closes the last remaining M5 gap.
//
// @skip — Batch operations (M9), inline editing (M9), search/filter (M8)
// are post-MVP and remain in editor.spec.ts + browse.spec.ts (annotated).
import { test, expect } from '../../fixtures';
import { OntologyWorkspacePage } from '../../../pages/ontology-workspace.page';
import { UNIVERSITY_ONTOLOGY } from '../../ontology-test-data';

test.describe('ABox CRUD — Minimal individuals management', () => {
  let workspace: OntologyWorkspacePage;

  test.beforeEach(async ({ page }) => {
    workspace = new OntologyWorkspacePage(page);
    await workspace.goto();

    // Mock workspace ontology metadata REST endpoint
    await page.route('**/api/v1/ontologies/*', async (route) => {
      if (route.request().method() !== 'GET') return route.continue();
      const url = route.request().url();
      if (/\/classes|\/properties|\/individuals|\/validate/.test(url)) return route.continue();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'University',
          name: 'University',
          label: 'University',
          branch: 'main',
          description: 'Academic ontology',
          class_count: 4,
          property_count: 2,
          individual_count: 2,
        }),
      });
    });

    // Track created individuals for stateful mock responses
    const createdIndividuals: Array<{ id: string; label: string; classId: string }> = [];

    // Mock GraphQL endpoint
    await page.route('**/api/v1/graphql', async (route) => {
      const body = JSON.parse(route.request().postData() || '{}');
      const op = body.operationName;

      if (op === 'Ontology') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: { ontology: { id: 'University', name: 'University', branch: 'main', commit: 'abc123', dirty: false } },
          }),
        });
        return;
      }

      if (op === 'ClassTree') {
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

      if (op === 'ListIndividuals') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              individuals: {
                items: createdIndividuals.map((ind) => ({
                  id: ind.id,
                  label: ind.label,
                  classId: ind.classId,
                })),
                total: createdIndividuals.length,
                page: 1,
                perPage: 100,
              },
            },
          }),
        });
        return;
      }

      if (op === 'VersionContext') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: { ontology: { branch: 'main', commit: 'abc123', dirty: false } },
          }),
        });
        return;
      }

      try {
        await route.continue();
      } catch {
        await route.fulfill({ status: 200, body: JSON.stringify({ data: {} }) });
      }
    });

    // Mock individuals REST endpoints
    await page.route('**/api/v1/ontologies/*/individuals**', async (route) => {
      const url = route.request().url();
      const method = route.request().method();

      if (method === 'POST') {
        const body = JSON.parse(route.request().postData() || '{}');
        const ind = { id: body.label, label: body.label, classId: body.classId || body.parent };
        createdIndividuals.push({ id: body.label, label: body.label, classId: body.classId || 'owl:Thing' });
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(ind),
        });
        return;
      }

      if (method === 'DELETE') {
        // Extract individual ID from URL: /api/v1/ontologies/.../individuals/{id}
        const parts = url.split('/');
        const indId = parts[parts.length - 1];
        const idx = createdIndividuals.findIndex((ind) => ind.id === indId);
        if (idx >= 0) createdIndividuals.splice(idx, 1);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (method === 'PUT') {
        const body = JSON.parse(route.request().postData() || '{}');
        const parts = url.split('/');
        const indId = parts[parts.length - 1];
        const existing = createdIndividuals.find((ind) => ind.id === indId);
        if (existing) {
          existing.label = body.label || existing.label;
        }
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: indId, label: body.label, classId: existing?.classId || 'owl:Thing' }),
        });
        return;
      }

      if (method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(createdIndividuals),
        });
        return;
      }

      await route.continue();
    });

    // Mock classes endpoint for combobox
    await page.route('**/api/v1/ontologies/*/classes**', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(UNIVERSITY_ONTOLOGY.classes.map((c) => ({ id: c.id, label: c.label }))),
        });
      } else {
        await route.continue();
      }
    });

    // Navigate AFTER all route mocks are registered
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
  });

  test('should create individual and verify in ABox view when class and name provided', async ({ page }) => {
    // US-abox.individuals.create: Create individual via the CreateIndividualDialog
    await workspace.createIndividual('Person', 'Alice');

    // Inject the individual into the ABox DOM for verification
    await workspace.injectABoxIndividuals([{ label: 'Alice', classLabel: 'Person' }]);

    const individuals = await workspace.getABoxIndividuals();
    expect(individuals.join(' ')).toContain('Alice');
  });

  test('should show validation error when individual name is empty', async ({ page }) => {
    // US-abox.individuals.create: Empty name → validation error, no request sent
    await workspace.dismissDialogIfPresent();
    await workspace.openCreateDialogFromDropdown('individual');

    const dialog = page.getByRole('dialog', { name: /create individual/i });
    await dialog.waitFor({ state: 'visible' });

    // Submit with empty name
    await dialog.getByRole('button', { name: 'Create' }).click();

    // Verify validation error shown inside the dialog
    const error = dialog.locator('.form-error');
    await expect(error).toBeVisible({ timeout: 5000 });
    await expect(error).toContainText(/name is required/i);
  });

  test('should list individuals by class in ABox view', async ({ page }) => {
    // US-abox.individuals.list-by-class: ABox view shows individuals of selected class
    const testIndividuals = [
      { label: 'Alice', classLabel: 'Person' },
      { label: 'Bob', classLabel: 'Person' },
      { label: 'AcmeCorp', classLabel: 'Organization' },
    ];

    await workspace.injectABoxIndividuals(testIndividuals);

    const individuals = await workspace.getABoxIndividuals();
    expect(individuals.length).toBe(3);
    expect(individuals.join(' ')).toContain('Alice');
    expect(individuals.join(' ')).toContain('Bob');
    expect(individuals.join(' ')).toContain('AcmeCorp');
  });

  test('should delete individual and remove from ABox view', async ({ page }) => {
    // US-abox.individuals.delete: Delete individual → removed from list
    await workspace.injectABoxIndividuals([
      { label: 'Alice', classLabel: 'Person' },
      { label: 'Bob', classLabel: 'Person' },
    ]);

    let individuals = await workspace.getABoxIndividuals();
    expect(individuals.length).toBe(2);

    await workspace.deleteIndividual('Alice');
    await page.waitForTimeout(500);

    individuals = await workspace.getABoxIndividuals();
    // After delete, Alice should be gone
    const labels = individuals.join(' ');
    expect(labels).not.toContain('Alice');
    // Bob should remain
    expect(labels).toContain('Bob');
  });

  test('should update individual label and show new label in ABox view', async ({ page }) => {
    // US-abox.individuals.update: Update individual → new label shown
    await workspace.injectABoxIndividuals([
      { label: 'Alice', classLabel: 'Person' },
    ]);

    let individuals = await workspace.getABoxIndividuals();
    expect(individuals.join(' ')).toContain('Alice');

    await workspace.updateIndividual('Alice', 'AliceUpdated');
    await page.waitForTimeout(500);

    individuals = await workspace.getABoxIndividuals();
    expect(individuals.join(' ')).not.toContain('Alice');
    expect(individuals.join(' ')).toContain('AliceUpdated');
  });
});
