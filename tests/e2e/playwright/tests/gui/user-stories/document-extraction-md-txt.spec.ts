import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';
import path from 'path';

// E2E-document.extract-md-txt — Document extraction from Markdown and plain text (P0)
// Covers US: US-io.document.extract-md-txt
//
// Tests:
// - Upload specification.md → verify preview with 8 classes, 5 properties
// - Edit a label in the preview → apply → verify commit success
// - Upload requirements.txt → verify structured extraction from plain text
//
// Mocked routes:
//   POST /api/v1/ontologies/:name/extract — returns extracted SequenceStep[]
//   POST /api/v1/ontologies/:name/apply — applies sequence and returns commit info

test.describe('Document Extraction — Markdown and Text', () => {
  let uploadPage: DocumentUploadPage;
  const MOCK_SEQUENCE_MD = [
    { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing' },
    { operation: 'CREATE_CLASS', entityId: 'Student', label: 'Student', parentId: 'Person' },
    { operation: 'CREATE_CLASS', entityId: 'Professor', label: 'Professor', parentId: 'Person' },
    { operation: 'CREATE_CLASS', entityId: 'Course', label: 'Course', parentId: 'owl:Thing' },
    { operation: 'CREATE_CLASS', entityId: 'Department', label: 'Department', parentId: 'owl:Thing' },
    { operation: 'CREATE_CLASS', entityId: 'Organization', label: 'Organization', parentId: 'owl:Thing' },
    { operation: 'CREATE_CLASS', entityId: 'ResearchGroup', label: 'ResearchGroup', parentId: 'Organization' },
    { operation: 'CREATE_CLASS', entityId: 'Building', label: 'Building', parentId: 'owl:Thing' },
    { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'advises', label: 'advises', domainId: 'Professor', rangeId: 'Student' },
    { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'teaches', label: 'teaches', domainId: 'Professor', rangeId: 'Course' },
    { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'studentId', label: 'studentId', domainId: 'Student', rangeId: 'string' },
    { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'gpa', label: 'gpa', domainId: 'Student', rangeId: 'float' },
    { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'employeeId', label: 'employeeId', domainId: 'Professor', rangeId: 'string' },
  ];

  const MOCK_COMMIT_RESPONSE = {
    success: true,
    commitId: 'abc123def456',
    commitUrl: '/commits/abc123def456',
    message: 'Extracted ontology from specification.md',
    branchId: 'main',
    timestamp: new Date().toISOString(),
    appliedCount: 13,
    entityCount: 13,
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);

    // Mock the extraction API endpoint for specification.md
    await page.route('**/api/v1/documents/extract', async (route) => {
      const request = route.request();
      const url = request.url();
      const method = request.method();

      if (method === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: MOCK_SEQUENCE_MD,
            sourceFile: 'specification.md',
            warnings: [],
          }),
        });
      } else {
        await route.continue();
      }
    });

    // Mock the apply endpoint
    await page.route('**/api/v1/documents/apply', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_COMMIT_RESPONSE),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('upload specification.md and verify preview with 8 classes, 5 properties', async () => {
    // US-io.document.extract-md-txt: Upload markdown → preview extracted ontology
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../../fixtures/specification.md'));

    // Verify preview shows extracted steps
    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBeGreaterThanOrEqual(13);

    // Verify class count (8 classes)
    const classSteps = steps.filter((s) => s.includes('Class'));
    expect(classSteps.length).toBe(8);

    // Verify property count (5 properties = 2 object + 3 datatype)
    const propertySteps = steps.filter(
      (s) => s.includes('CREATE_OBJECT_PROPERTY') || s.includes('CREATE_DATATYPE_PROPERTY')
    );
    expect(propertySteps.length).toBe(5);
  });

  test('edit a label in the preview and apply sequence', async () => {
    // US-io.document.extract-md-txt: Edit extracted label before applying
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../../fixtures/specification.md'));

    // Edit the first class label
    await uploadPage.editStepLabel(0, 'Human');

    // Verify label changed
    const labels = await uploadPage.getStepLabels();
    expect(labels[0]).toContain('Human');

    // Apply the sequence
    await uploadPage.applySequence();

    // Verify apply completed
    const success = await uploadPage.waitForApplyComplete();
    expect(success).toBe(true);

    // Verify commit link is displayed
    const commitLink = await uploadPage.getCommitLink();
    expect(commitLink).not.toBeNull();
  });

  test('upload requirements.txt and verify plain text extraction', async () => {
    // US-io.document.extract-md-txt: Plain text extraction from structured text
    // Use a second mock route for plain text extraction
    await uploadPage.page.unroute('**/api/v1/documents/extract');
    await uploadPage.page.route('**/api/v1/documents/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'Product', label: 'Product', parentId: 'owl:Thing' },
              { operation: 'CREATE_CLASS', entityId: 'Category', label: 'Category', parentId: 'owl:Thing' },
              { operation: 'CREATE_CLASS', entityId: 'Electronic', label: 'Electronic', parentId: 'Product' },
              { operation: 'CREATE_CLASS', entityId: 'Clothing', label: 'Clothing', parentId: 'Product' },
              { operation: 'CREATE_CLASS', entityId: 'Customer', label: 'Customer', parentId: 'owl:Thing' },
              { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'purchases', label: 'purchases', domainId: 'Customer', rangeId: 'Product' },
              { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'price', label: 'price', domainId: 'Product', rangeId: 'decimal' },
            ],
            sourceFile: 'requirements.txt',
            warnings: [],
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../../fixtures/requirements.txt'));

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBeGreaterThanOrEqual(5);

    // Verify Product class was extracted
    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Product'))).toBe(true);
    expect(labels.some((l) => l.includes('Category'))).toBe(true);
    expect(labels.some((l) => l.includes('Customer'))).toBe(true);
  });
});
