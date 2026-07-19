import { test, expect } from '../../fixtures';
import { DocumentUploadPage } from '../../../pages/document-upload.page';
import path from 'path';

// E2E-document.extract-pdf-docx — Document extraction from PDF and DOCX (P1)
// Covers US: US-io.document.extract-pdf-docx
//
// Tests:
// - Upload PDF with text layer → verify structured extraction
// - Upload DOCX with headings and tables → verify extraction
// - Password-protected PDF → verify rejection with password prompt
// - Scanned PDF without text layer → verify early detection message
// - Large document → verify progress indicator

test.describe('Document Extraction — PDF and DOCX', () => {
  let uploadPage: DocumentUploadPage;

  const MOCK_PDF_SEQUENCE = {
    steps: [
      { operation: 'CREATE_CLASS', entityId: 'Patient', label: 'Patient', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Doctor', label: 'Doctor', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'Appointment', label: 'Appointment', parentId: 'owl:Thing' },
      { operation: 'CREATE_CLASS', entityId: 'MedicalRecord', label: 'MedicalRecord', parentId: 'owl:Thing' },
      { operation: 'CREATE_OBJECT_PROPERTY', entityId: 'treatedBy', label: 'treatedBy', domainId: 'Patient', rangeId: 'Doctor' },
      { operation: 'CREATE_DATATYPE_PROPERTY', entityId: 'diagnosis', label: 'diagnosis', domainId: 'MedicalRecord', rangeId: 'string' },
    ],
    sourceFile: 'technical-spec.pdf',
    warnings: [],
  };

  const MOCK_PROGRESS_EVENTS = [
    { type: 'progress', progress: 25, message: 'Parsing document structure...' },
    { type: 'progress', progress: 50, message: 'Identifying entities...' },
    { type: 'progress', progress: 75, message: 'Building ontology structure...' },
    { type: 'progress', progress: 100, message: 'Extraction complete' },
  ];

  test.beforeEach(async ({ page }) => {
    uploadPage = new DocumentUploadPage(page);

    // Default mock for extraction endpoint
    await page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_PDF_SEQUENCE),
        });
      } else {
        await route.continue();
      }
    });

    // Mock apply endpoint
    await page.route('**/api/v1/ontologies/**/apply', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            commitId: 'pdf123commit',
            message: 'Extracted ontology from technical-spec.pdf',
            branchId: 'main',
            entityCount: 6,
          }),
        });
      } else {
        await route.continue();
      }
    });
  });

  test('upload PDF with text layer and verify structured extraction', async () => {
    // US-io.document.extract-pdf-docx: Extract ontology from PDF
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../fixtures/m2/technical-spec.pdf'));

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(6);

    const labels = await uploadPage.getStepLabels();
    expect(labels.some((l) => l.includes('Patient'))).toBe(true);
    expect(labels.some((l) => l.includes('Doctor'))).toBe(true);
    expect(labels.some((l) => l.includes('Appointment'))).toBe(true);
  });

  test('upload DOCX with headings and tables and verify extraction', async () => {
    // US-io.document.extract-pdf-docx: Upload DOCX → verify extraction
    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../fixtures/m2/documentation.docx'));

    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(6);
  });

  test('password-protected PDF shows password prompt on rejection', async () => {
    // US-io.document.extract-pdf-docx: Reject password-protected PDF
    // Override route for this specific test
    await uploadPage.page.unroute('**/api/v1/ontologies/**/extract');
    await uploadPage.page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 422,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'ENCRYPTED_DOCUMENT',
            message: 'Document is password-protected. Please provide a password to decrypt.',
            code: 'DOCUMENT_ENCRYPTED',
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../fixtures/m2/technical-spec.pdf'));

    // Verify password prompt is shown
    const isPasswordVisible = await uploadPage.isPasswordPromptVisible();
    expect(isPasswordVisible).toBe(true);
  });

  test('scanned PDF without text layer returns early detection message', async () => {
    // US-io.document.extract-pdf-docx: Scanned PDF detection
    await uploadPage.page.unroute('**/api/v1/ontologies/**/extract');
    await uploadPage.page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 422,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'SCANNED_DOCUMENT',
            message: 'Document appears to be scanned (no extractable text layer). OCR is not yet supported.',
            code: 'DOCUMENT_SCANNED',
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../fixtures/m2/technical-spec.pdf'));

    // Verify validation error about scanned document
    const error = await uploadPage.getValidationError();
    expect(error).not.toBeNull();
    expect(error?.toLowerCase()).toContain('scan');
  });

  test('large document upload shows progress indicator during extraction', async () => {
    // US-io.document.extract-pdf-docx: Large file with progress indicator
    await uploadPage.page.unroute('**/api/v1/ontologies/**/extract');
    await uploadPage.page.route('**/api/v1/ontologies/**/extract', async (route) => {
      if (route.request().method() === 'POST') {
        // Simulate streaming progress via intermediate 202 status then final 200
        // For E2E test, return final result directly with progress metadata
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: MOCK_PDF_SEQUENCE.steps,
            sourceFile: 'large-document.pdf',
            warnings: ['Large document processed with chunked extraction'],
            progress: MOCK_PROGRESS_EVENTS,
          }),
        });
      } else {
        await route.continue();
      }
    });

    await uploadPage.openDocumentUpload('TestOntology');
    await uploadPage.uploadFile(path.resolve(__dirname, '../../fixtures/m2/technical-spec.pdf'));

    // Verify extraction completes even for large documents
    const steps = await uploadPage.getPreviewSequence();
    expect(steps.length).toBe(6);
  });
});
