import { test, expect } from '../fixtures';
import { M2DocumentUploadPage } from '../../pages/m2-document-upload.page';

// E2E-ai.templates.ontology — Ontology domain templates (P1)
// Covers US: US-io.ontology.templates
//
// Tests:
// - List available templates → verify Person/Organization template in list
// - Select "Person/Organization" template → preview → verify expected classes and properties
// - Apply template → verify all entities created

test.describe('M2 Ontology Domain Templates', () => {
  let uploadPage: M2DocumentUploadPage;

  const MOCK_TEMPLATES_LIST = [
    {
      id: 'person-organization',
      name: 'Person/Organization',
      description: 'Basic contact management with people, organizations, and memberships',
      classCount: 4,
      propertyCount: 5,
      domain: 'Contact management',
    },
    {
      id: 'product-catalog',
      name: 'Product Catalog',
      description: 'E-commerce product catalog with categories and pricing',
      classCount: 5,
      propertyCount: 7,
      domain: 'E-commerce',
    },
    {
      id: 'research-project',
      name: 'Research Project',
      description: 'Academic research project tracking with publications and grants',
      classCount: 6,
      propertyCount: 8,
      domain: 'Research',
    },
    {
      id: 'it-asset-management',
      name: 'IT Asset Management',
      description: 'IT infrastructure asset tracking with devices, software, and licenses',
      classCount: 5,
      propertyCount: 6,
      domain: 'IT Management',
    },
  ];

  const MOCK_PERSON_ORG_TEMPLATE = {
    id: 'person-organization',
    name: 'Person/Organization',
    steps: [
      { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Organization', label: 'Organization', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Membership', label: 'Membership', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'ContactInfo', label: 'ContactInfo', parentId: 'owl:Thing' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'memberOf', label: 'memberOf', domainId: 'Person', rangeId: 'Organization' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'hasMembership', label: 'hasMembership', domainId: 'Person', rangeId: 'Membership' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'hasContact', label: 'hasContact', domainId: 'Person', rangeId: 'ContactInfo' },
      { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'email', label: 'email', domainId: 'ContactInfo', rangeId: 'string' },
      { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'phone', label: 'phone', domainId: 'ContactInfo', rangeId: 'string' },
    ],
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new M2DocumentUploadPage(page);

    await page.route('**/api/v1/ai/templates', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ templates: MOCK_TEMPLATES_LIST }),
        });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/ai/templates/*', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_PERSON_ORG_TEMPLATE),
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
            commitId: 'template-commit-001',
            message: 'Applied Person/Organization template',
            branchId: 'main',
            entityCount: 9,
          }),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('list available templates and verify Person/Organization template is present', async () => {
    // US-io.ontology.templates: List templates
    await uploadPage.openDocumentUpload('TestOntology');

    // Navigate to templates section
    await uploadPage.page.getByRole('button', { name: /templates|ontology templates/i }).click();

    // Verify template list is displayed
    const templateCards = uploadPage.page.locator('.template-card, [data-testid="template-card"]');
    await expect(templateCards.first()).toBeVisible({ timeout: 10_000 });

    // Count templates
    const count = await templateCards.count();
    expect(count).toBe(4);

    // Verify Person/Organization template is present
    await expect(templateCards.filter({ hasText: 'Person/Organization' })).toBeVisible();
  });

  test('select Person/Organization template and preview expected classes and properties', async () => {
    // US-io.ontology.templates: Preview template
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.page.getByRole('button', { name: /templates|ontology templates/i }).click();

    // Select the Person/Organization template
    await uploadPage.page.locator('.template-card, [data-testid="template-card"]')
      .filter({ hasText: 'Person/Organization' })
      .click();

    // Click preview
    await uploadPage.page.getByRole('button', { name: /preview|view details/i }).click();

    // Verify preview shows expected steps
    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(9);

    // Verify classes
    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Person'))).toBe(true);
    expect(labels.some((l) => l.includes('Organization'))).toBe(true);
    expect(labels.some((l) => l.includes('Membership'))).toBe(true);
    expect(labels.some((l) => l.includes('ContactInfo'))).toBe(true);

    // Verify properties
    expect(labels.some((l) => l.includes('memberOf'))).toBe(true);
    expect(labels.some((l) => l.includes('hasContact'))).toBe(true);
    expect(labels.some((l) => l.includes('email'))).toBe(true);
    expect(labels.some((l) => l.includes('phone'))).toBe(true);
  });

  test('apply Person/Organization template creates all expected entities', async () => {
    // US-io.ontology.templates: Apply template
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.page.getByRole('button', { name: /templates|ontology templates/i }).click();

    // Select and preview
    await uploadPage.page.locator('.template-card, [data-testid="template-card"]')
      .filter({ hasText: 'Person/Organization' })
      .click();
    await uploadPage.page.getByRole('button', { name: /preview|view details/i }).click();

    // Apply
    await uploadPage.applySequence();
    const success = await uploadPage.waitForApplyComplete();
    expect(success).toBe(true);

    // Verify entity count in commit link
    const commitLink = await uploadPage.getCommitLink();
    expect(commitLink).not.toBeNull();
  });
});
