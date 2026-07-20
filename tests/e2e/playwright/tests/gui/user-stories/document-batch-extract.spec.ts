import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';
import path from 'path';

// E2E-document.batch-extract — Batch document extraction with deduplication and conflict resolution (P1)
// Covers US: US-io.document.batch-extract
//
// Tests:
// - Upload 3 files of different formats → verify combined preview with source attribution
// - Deduplication: same class extracted from multiple files → shown once with sources
// - Conflict resolution: different definitions for same entity → conflict UI choices
// - Partial failure: one file fails → other files still extracted successfully

// @skip — batch document upload with deduplication, conflict resolution, and partial-failure
// handling is a POST-MVP capability.
// Roadmap: M9 (Document AI 1.0 — POST-MVP). M2 covers single-file document extraction only.
// Unskip when: batch upload endpoint and deduplication/conflict resolution UI are implemented.
test.describe.skip('Document Extraction — Batch Upload', () => {
  let uploadPage: DocumentUploadPage;

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);
  });

  test('upload 3 files of different formats with combined preview and source attribution', async () => {
    // US-io.document.batch-extract: Multi-format batch extraction
    await uploadPage.page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        // Extract source filename from request to return appropriate response
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing', sourceFile: 'specification.md' },
              { operation: 'CREATE_CLASS', entityId: 'Student', label: 'Student', parentId: 'Person', sourceFile: 'specification.md' },
              { operation: 'CREATE_CLASS', entityId: 'Product', label: 'Product', parentId: 'owl:Thing', sourceFile: 'data.json' },
              { operation: 'CREATE_CLASS', entityId: 'Customer', label: 'Customer', parentId: 'owl:Thing', sourceFile: 'data.json' },
              { operation: 'CREATE_CLASS', entityId: 'Vehicle', label: 'Vehicle', parentId: 'owl:Thing', sourceFile: 'entities.csv' },
            ],
            sourceFile: 'batch',
            warnings: [],
            batchResults: [
              { fileName: 'specification.md', status: 'success', stepCount: 2 },
              { fileName: 'data.json', status: 'success', stepCount: 2 },
              { fileName: 'entities.csv', status: 'success', stepCount: 1 },
            ],
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.page.route('**/api/v1/ontologies/**/apply', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            commitId: 'batch-commit-001',
            message: 'Batch extract from 3 files',
            branchId: 'main',
            entityCount: 5,
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadMultipleFiles([
      path.resolve(__dirname, '../../fixtures/m2/specification.md'),
      path.resolve(__dirname, '../../fixtures/m2/data.json'),
      path.resolve(__dirname, '../../fixtures/m2/entities.csv'),
    ]);

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(5);

    // Verify source attribution for each step
    const personSource = await uploadPage.getSourceAttribution(0);
    expect(personSource).not.toBeNull();
    expect(personSource?.toLowerCase()).toContain('specification.md');

    const productSource = await uploadPage.getSourceAttribution(
      steps.findIndex((s) => s.includes('Product'))
    );
    expect(productSource).not.toBeNull();
    expect(productSource?.toLowerCase()).toContain('data.json');
  });

  test('deduplication: same class from multiple files shown once with combined sources', async () => {
    // US-io.document.batch-extract: Deduplication
    await uploadPage.page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing', sourceFile: 'specification.md' },
              { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing', sourceFile: 'requirements.txt', skipIfExists: true },
              { operation: 'CREATE_CLASS', entityId: 'Student', label: 'Student', parentId: 'Person', sourceFile: 'specification.md' },
            ],
            sourceFile: 'batch',
            warnings: ['Person was defined in 2 files — deduplicated to single definition'],
            deduplicated: ['Person'],
            batchResults: [
              { fileName: 'specification.md', status: 'success', stepCount: 2 },
              { fileName: 'requirements.txt', status: 'success', stepCount: 0 },
            ],
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadMultipleFiles([
      path.resolve(__dirname, '../../fixtures/m2/specification.md'),
      path.resolve(__dirname, '../../fixtures/m2/requirements.txt'),
    ]);

    // Verify deduplication warning
    const warning = await uploadPage.getDuplicateWarning();
    expect(warning).not.toBeNull();
    expect(warning?.toLowerCase()).toContain('deduplicat');

    // Verify only one Person step (not two)
    const steps = await uploadPage.getStepLabels();
    const personCount = steps.filter((l) => l.includes('Person')).length;
    expect(personCount).toBe(1);
  });

  test('conflict resolution: different definitions trigger conflict UI with resolution choices', async () => {
    // US-io.document.batch-extract: Conflict resolution
    await uploadPage.page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing', sourceFile: 'specification.md' },
            ],
            sourceFile: 'batch',
            warnings: [],
            conflicts: [
              {
                entityId: 'Person',
                label: 'Person',
                definitions: [
                  { parentId: 'owl:Thing', sourceFile: 'specification.md' },
                  { parentId: 'owl:Top', sourceFile: 'requirements.txt' },
                ],
                type: 'PARENT_MISMATCH',
              },
            ],
            batchResults: [
              { fileName: 'specification.md', status: 'success', stepCount: 1 },
              { fileName: 'requirements.txt', status: 'conflict', stepCount: 0 },
            ],
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadMultipleFiles([
      path.resolve(__dirname, '../../fixtures/m2/specification.md'),
      path.resolve(__dirname, '../../fixtures/m2/requirements.txt'),
    ]);

    // Verify conflicts are displayed
    const conflictCount = await uploadPage.getConflictCount();
    expect(conflictCount).toBeGreaterThanOrEqual(1);

    // Resolve all conflicts by keeping the new definition
    await uploadPage.resolveAllConflicts('keep-new');

    // After resolving, apply should be available
    await uploadPage.applySequence();
    const success = await uploadPage.waitForApplyComplete();
    expect(success).toBe(true);
  });

  test('partial failure: one file fails while others are still extracted successfully', async () => {
    // US-io.document.batch-extract: Partial failure handling
    await uploadPage.page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { operation: 'CREATE_CLASS', entityId: 'Person', label: 'Person', parentId: 'owl:Thing', sourceFile: 'specification.md' },
              { operation: 'CREATE_CLASS', entityId: 'Vehicle', label: 'Vehicle', parentId: 'owl:Thing', sourceFile: 'entities.csv' },
            ],
            sourceFile: 'batch',
            warnings: ['data.json failed to parse: unsupported schema version'],
            batchResults: [
              { fileName: 'specification.md', status: 'success', stepCount: 1 },
              { fileName: 'data.json', status: 'failed', error: 'Unsupported schema version' },
              { fileName: 'entities.csv', status: 'success', stepCount: 1 },
            ],
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadMultipleFiles([
      path.resolve(__dirname, '../../fixtures/m2/specification.md'),
      path.resolve(__dirname, '../../fixtures/m2/data.json'),
      path.resolve(__dirname, '../../fixtures/m2/entities.csv'),
    ]);

    // Verify successful steps are present
    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Person'))).toBe(true);
    expect(labels.some((l) => l.includes('Vehicle'))).toBe(true);

    // Verify the failed file status shows error
    const jsonStatus = await uploadPage.getFileUploadStatus('data.json');
    expect(jsonStatus).not.toBeNull();
    expect(jsonStatus?.toLowerCase()).toContain('fail');
  });
});
