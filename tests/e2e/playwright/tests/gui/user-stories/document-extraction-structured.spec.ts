import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';
import path from 'path';

// E2E-document.extract-structured — Document extraction from structured data formats (P1)
// Covers US: US-io.document.extract-structured
//
// Tests:
// - Upload XLSX → verify column mapping for Class/Parent/Property columns
// - Upload JSON with nested objects → verify hierarchy preservation
// - Upload XML with tags and attributes → verify class mapping
// - Upload CSV with delimiter detection → verify parsing accuracy

test.describe('Document Extraction — Structured Data', () => {
  let uploadPage: DocumentUploadPage;

  const MOCK_JSON_SEQUENCE = {
    steps: [
      { operation: 'CREATE_CLASS', entityId: 'Product', label: 'Product', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Book', label: 'Book', parentId: 'Product', parentLabel: 'Product' },
      { operation: 'CREATE_CLASS', entityId: 'Electronics', label: 'Electronics', parentId: 'Product', parentLabel: 'Product' },
      { operation: 'CREATE_CLASS', entityId: 'Customer', label: 'Customer', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Order', label: 'Order', parentId: 'owl:Thing' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'purchases', label: 'purchases', domain: 'Customer', range: 'Product' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'contains', label: 'contains', domain: 'Order', range: 'Product' },
      { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'price', label: 'price', domain: 'Product', range: 'decimal' },
      { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'quantity', label: 'quantity', domain: 'Order', range: 'integer' },
    ],
    sourceFile: 'data.json',
    warnings: [],
  };

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);

    // Default mock for extraction - overridden per test
    await page.route('**/api/v1/documents/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [],
            sourceFile: 'unknown',
            warnings: [],
          }),
        });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/documents/apply', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            commitId: 'struct-commit-001',
            message: 'Extracted ontology from structured data',
            branchId: 'main',
            entityCount: 9,
          }),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('upload XLSX with Class/Parent/Property columns and verify column mapping', async () => {
    // US-io.document.extract-structured: XLSX column mapping
    await uploadPage.page.unroute('**/api/v1/documents/extract');
    await uploadPage.page.route('**/api/v1/documents/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'Vehicle', label: 'Vehicle', parentId: 'owl:Thing' },
              { operation: 'CREATE_CLASS', entityId: 'Car', label: 'Car', parentId: 'Vehicle', parentLabel: 'Vehicle' },
              { operation: 'CREATE_CLASS', entityId: 'Truck', label: 'Truck', parentId: 'Vehicle', parentLabel: 'Vehicle' },
              { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'hasSpeed', label: 'hasSpeed', domain: 'Vehicle', range: 'decimal' },
              { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'hasCapacity', label: 'hasCapacity', domain: 'Vehicle', range: 'integer' },
            ],
            sourceFile: 'classes.xlsx',
            warnings: [],
            columnMapping: {
              classColumn: 'Class',
              parentColumn: 'Parent',
              propertyColumn: 'Property',
              domainColumn: 'Domain',
              rangeColumn: 'Range',
            },
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../../fixtures/classes.xlsx'));

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(5);

    // Verify classes extracted with hierarchy
    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Vehicle'))).toBe(true);
    expect(labels.some((l) => l.includes('Car'))).toBe(true);
    expect(labels.some((l) => l.includes('Truck'))).toBe(true);

    // Verify property extraction
    expect(labels.some((l) => l.includes('hasSpeed'))).toBe(true);
    expect(labels.some((l) => l.includes('hasCapacity'))).toBe(true);
  });

  test('upload JSON with nested objects and verify hierarchy', async () => {
    // US-io.document.extract-structured: JSON nested objects → hierarchy
    await uploadPage.page.unroute('**/api/v1/documents/extract');
    await uploadPage.page.route('**/api/v1/documents/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_JSON_SEQUENCE),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../../fixtures/data.json'));

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(9);

    // Verify class hierarchy (Book → Product)
    const classSteps = await uploadPage.getPreviewSequence();
    const bookStep = classSteps.find((s) => s.includes('Book'));
    expect(bookStep).toBeDefined();
    expect(bookStep).toContain('Product');

    // Verify object properties (purchases: Customer → Product)
    const purchasesStep = classSteps.find((s) => s.includes('purchases'));
    expect(purchasesStep).toBeDefined();

    // Verify datatype properties (price on Product)
    const priceStep = classSteps.find((s) => s.includes('price'));
    expect(priceStep).toBeDefined();
  });

  test('upload XML with tags and attributes and verify class mapping', async () => {
    // US-io.document.extract-structured: XML tags/attributes → class mapping
    await uploadPage.page.unroute('**/api/v1/documents/extract');
    await uploadPage.page.route('**/api/v1/documents/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'LibraryItem', label: 'LibraryItem', parentId: 'owl:Thing' },
              { operation: 'CREATE_CLASS', entityId: 'Book', label: 'Book', parentId: 'LibraryItem', parentLabel: 'LibraryItem' },
              { operation: 'CREATE_CLASS', entityId: 'Magazine', label: 'Magazine', parentId: 'LibraryItem', parentLabel: 'LibraryItem' },
              { operation: 'CREATE_CLASS', entityId: 'Author', label: 'Author', parentId: 'owl:Thing' },
              { operation: 'CREATE_CLASS', entityId: 'Member', label: 'Member', parentId: 'owl:Thing' },
              { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'writtenBy', label: 'writtenBy', domain: 'Book', range: 'Author' },
              { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'borrowedBy', label: 'borrowedBy', domain: 'LibraryItem', range: 'Member' },
              { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'isbn', label: 'isbn', domain: 'Book', range: 'string' },
            ],
            sourceFile: 'schema.xml',
            warnings: [],
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../../fixtures/schema.xml'));

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(8);

    // Verify classes and hierarchy
    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('LibraryItem'))).toBe(true);
    expect(labels.some((l) => l.includes('Book'))).toBe(true);
    expect(labels.some((l) => l.includes('Author'))).toBe(true);
  });

  test('upload CSV with delimiter detection and verify parsing', async () => {
    // US-io.document.extract-structured: CSV with delimiter detection
    await uploadPage.page.unroute('**/api/v1/documents/extract');
    await uploadPage.page.route('**/api/v1/documents/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'Vehicle', label: 'Vehicle', parentId: 'owl:Thing' },
              { operation: 'CREATE_CLASS', entityId: 'Car', label: 'Car', parentId: 'Vehicle', parentLabel: 'Vehicle' },
              { operation: 'CREATE_CLASS', entityId: 'Motorcycle', label: 'Motorcycle', parentId: 'Vehicle', parentLabel: 'Vehicle' },
              { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'hasSpeed', label: 'hasSpeed', domain: 'Vehicle', range: 'decimal' },
              { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'hasColor', label: 'hasColor', domain: 'Vehicle', range: 'string' },
              { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'drives', label: 'drives', domain: 'Person', range: 'Vehicle' },
            ],
            sourceFile: 'entities.csv',
            warnings: [],
            detectedDelimiter: ',',
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../../fixtures/entities.csv'));

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(6);

    // Verify all expected entities
    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Vehicle'))).toBe(true);
    expect(labels.some((l) => l.includes('Car'))).toBe(true);
    expect(labels.some((l) => l.includes('Motorcycle'))).toBe(true);
    expect(labels.some((l) => l.includes('drives'))).toBe(true);
  });
});
