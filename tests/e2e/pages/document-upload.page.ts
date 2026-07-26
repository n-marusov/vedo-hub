import { type Page } from '@playwright/test';
import { OWNER_JWT } from '../specs/jwt-tokens';

// Page Object Model for Document Extraction and AI Ontology Creation
// Covers: US-io.document.extract-*, US-io.document.batch-extract, US-io.document.preview-sequence

export class DocumentUploadPage {
  constructor(public readonly page: Page) {}

  async setupExtractionMocks() {
    // Kept for backward compatibility with older tests; current user-story specs
    // install their own Playwright route mocks for /api/v1/documents/*.
  }

  async setupBrowserAuth() {
    const session = {
      accessToken: OWNER_JWT,
      refreshToken: OWNER_JWT,
      userId: 'user-123',
      tenantId: 'org-001',
      roles: ['Owner'],
      expiresAt: Date.now() + 86_400_000,
    };

    await this.page.addInitScript(({ sessionJson, token }) => {
      sessionStorage.setItem('vedo_session', sessionJson);
      localStorage.setItem('vedo-jwt-token', token);
    }, { sessionJson: JSON.stringify(session), token: OWNER_JWT });
  }

  // ─── Navigation ───────────────────────────────────────────────────

  async goto() {
    await this.page.goto('/');
  }

  async openDocumentUpload(ontologyName: string) {
    await this.setupBrowserAuth();

    // Mock ontology REST endpoint to avoid 404 when backend doesn't have the ontology.
    // The workspace page fetches /api/v1/ontologies/${ontologyName} on mount and
    // shows an error if it 404s, preventing the AI Import button from rendering.
    await this.page.route(`**/api/v1/ontologies/${ontologyName}`, async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: ontologyName,
            name: ontologyName,
            branch: 'main',
            dirty: false,
          }),
        });
      } else {
        await route.continue();
      }
    });

    await this.page.goto(`/project/${ontologyName}/workspace`);
    await this.page.getByRole('button', { name: /ai import/i }).click();
    await this.page.getByRole('region', { name: /document upload zone|batch document upload/i }).waitFor({ state: 'visible' });
  }

  /// Opens the AI Import panel and switches to the "NL→OWL" sub-tab.
  /// Required for natural-language tests — `openDocumentUpload` lands on
  /// the "Document Import" sub-tab by default.
  async openNLToOWLImport(ontologyName: string) {
    await this.openDocumentUpload(ontologyName);
    await this.page.getByRole('button', { name: /nl→owl/i }).click();
    await this.nlInput().waitFor({ state: 'visible' });
  }

  /// Locator for the NL→OWL textarea. Uses role+name because the component
  /// renders a labelled textbox without a stable `data-testid` or
  /// `.nl-to-owl-input` wrapper — the accessible name is stable across themes.
  nlInput() {
    return this.page.getByRole('textbox', { name: /describe your ontology in natural language/i });
  }

  // ─── File Upload ──────────────────────────────────────────────────

  async uploadFile(filePath: string) {
    const fileChooserPromise = this.page.waitForEvent('filechooser');
    await this.page.getByRole('button', { name: /upload document|choose file|browse/i }).click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles([filePath]);
    // Wait for upload processing to complete
    await this.page.waitForResponse((resp) =>
      resp.url().includes('/api/v1/documents/extract') && resp.status() === 200,
      { timeout: 15_000 }
    ).catch(() => {
      // Mock route handles this through fulfilled responses
    });
  }

  async uploadMultipleFiles(filePaths: string[]) {
    await this.page.getByRole('button', { name: /batch upload/i }).click();
    const fileChooserPromise = this.page.waitForEvent('filechooser');
    await this.page.getByRole('region', { name: /batch document upload/i }).getByRole('button', { name: /browse/i }).click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(filePaths);
    await this.page.locator('.batch-uploader__upload-btn').click();
    await this.page.waitForSelector('.preview-row', { timeout: 15_000 }).catch(() => {});
  }

  // ─── Preview Sequence ─────────────────────────────────────────────

  async getPreviewSequence(): Promise<string[]> {
    return this.page.locator('.preview-row').allTextContents();
  }

  async getStepLabels(): Promise<string[]> {
    return this.page.locator('.preview-row__label-text').allTextContents();
  }

  async getStepCount(): Promise<number> {
    return this.page.locator('.preview-row').count();
  }

  async editStepLabel(index: number, newLabel: string) {
    const step = this.page.locator('.preview-row').nth(index);
    await step.locator('.preview-row__label-text').click();
    const input = step.locator('.preview-row__edit-input');
    await input.waitFor({ state: 'visible' });
    await input.fill(newLabel);
    await input.press('Enter');
    await step.locator('.preview-row__label-text').filter({ hasText: newLabel }).waitFor({ state: 'visible' });
  }

  async toggleStep(index: number) {
    await this.page.locator('.preview-row').nth(index)
      .locator('.preview-row__toggle-input').check({ force: true });
  }

  async toggleStepInclusion(index: number, include: boolean) {
    const checkbox = this.page.locator('.preview-row').nth(index)
      .locator('.preview-row__toggle-input');
    const isChecked = await checkbox.isChecked();
    if ((include && !isChecked) || (!include && isChecked)) {
      await checkbox.evaluate((element) => (element as HTMLInputElement).click());
    }
  }

  async getStepWarning(index: number): Promise<string | null> {
    const row = this.page.locator('.preview-row').nth(index);
    const duplicate = row.locator('.preview-row__duplicate');
    if (await duplicate.isVisible()) {
      return duplicate.textContent();
    }
    const classes = await row.getAttribute('class');
    if (classes?.includes('preview-row--excluded')) {
      return 'excluded';
    }
    return null;
  }

  // ─── Duplicate Detection ──────────────────────────────────────────

  async getDuplicateWarning(): Promise<string | null> {
    const warning = this.page.locator('.preview__duplicates-info, .preview-row__duplicate, .uploader__error-message').first();
    if (await warning.isVisible()) {
      return warning.textContent();
    }
    return null;
  }

  // ─── Batch / Source Attribution ───────────────────────────────────

  async getSourceAttribution(stepIndex: number): Promise<string | null> {
    const attr = this.page.locator('.preview-row').nth(stepIndex)
      .locator('.preview-row__source-badge, .preview__group-title');
    if (await attr.isVisible()) {
      return attr.textContent();
    }
    return null;
  }

  async getFileUploadStatus(fileName: string): Promise<string | null> {
    const fileRow = this.page.locator('.batch-uploader__file').filter({ hasText: fileName }).first();
    const statusEl = fileRow.locator('.batch-uploader__file-status').first();
    if (await statusEl.isVisible()) {
      return statusEl.textContent();
    }
    return null;
  }

  // ─── Conflict Resolution ──────────────────────────────────────────

  async getConflictCount(): Promise<number> {
    return this.page.locator('.conflict-item').count();
  }

  async resolveConflict(index: number, strategy: 'keep-new' | 'keep-existing' | 'keep-both') {
    const conflict = this.page.locator('.conflict-item').nth(index);
    const buttonName = strategy === 'keep-new' ? /keep b/i : /keep a/i;
    await conflict.getByRole('button', { name: buttonName }).click();
  }

  async resolveAllConflicts(strategy: 'keep-new' | 'keep-existing' | 'keep-both') {
    const count = await this.getConflictCount();
    for (let i = 0; i < count; i++) {
      await this.resolveConflict(i, strategy);
    }
    await this.page.getByRole('button', { name: /apply resolutions/i }).click();
  }

  // ─── Apply Workflow ───────────────────────────────────────────────

  async applySequence() {
    await this.page.locator('.apply-flow .apply-btn').click();
  }

  /// Explicitly confirm the import dialog (advance from 'confirming' to 'applying')
  /// The commit message is pre-filled by the composable.
  async confirmImport() {
    await this.page.getByRole('dialog', { name: /apply import/i })
      .locator('.modal__btn--primary')
      .click();
  }

  async getProgressBarValue(): Promise<number | null> {
    const progressBar = this.page.locator('.modal__progress-fill');
    if (await progressBar.isVisible()) {
      const value = await progressBar.getAttribute('value');
      return value ? parseInt(value, 10) : null;
    }
    return null;
  }

  async getProgressText(): Promise<string | null> {
    const text = this.page.locator('.modal__progress-text');
    if (await text.isVisible()) {
      return text.textContent();
    }
    return null;
  }

  async waitForApplyComplete(timeout = 30_000): Promise<boolean> {
    try {
      // Click the modal's confirm button, scoped to the dialog context
      // to avoid matching the Retry button in error state.
      await this.page.getByRole('dialog', { name: /apply import/i })
        .locator('.modal__btn--primary')
        .click();
      await this.page.waitForSelector('.modal__title--success', { timeout });
      return true;
    } catch {
      return false;
    }
  }

  async getSuccessMessage(): Promise<string | null> {
    const success = this.page.locator('.modal__title--success, .modal__desc').first();
    if (await success.isVisible()) {
      return success.textContent();
    }
    return null;
  }

  async getCommitLink(): Promise<string | null> {
    const link = this.page.locator('.modal__commit-link a, a.modal__commit-hash').first();
    if (await link.isVisible()) {
      return link.getAttribute('href');
    }
    return null;
  }

  /// Reads the error message from the apply modal.
  /// NOTE: Does NOT have side effects — call confirmImport() first if the
  /// dialog is still in 'confirming' state.
  async getErrorMessage(): Promise<string | null> {
    const errorTitle = this.page.locator('.modal__title--error').first();
    if (await errorTitle.isVisible({ timeout: 15_000 }).catch(() => false)) {
      return errorTitle.textContent();
    }
    // Fallback: look inside the dialog for any error description text
    const dialog = this.page.getByRole('dialog', { name: /apply import/i });
    const desc = dialog.locator('.modal__desc').first();
    if (await desc.isVisible().catch(() => false)) {
      return desc.textContent();
    }
    return null;
  }

  async clickRetry() {
    await this.page.getByRole('dialog', { name: /apply import/i })
      .getByRole('button', { name: /retry/i })
      .click();
  }

  // ─── Error / Edge Cases ───────────────────────────────────────────

  async getValidationError(): Promise<string | null> {
    const error = this.page.locator('.uploader__error-message, .modal__desc').first();
    if (await error.isVisible()) {
      return error.textContent();
    }
    return null;
  }

  async isPasswordPromptVisible(): Promise<boolean> {
    return this.page.getByLabel(/password/i).isVisible();
  }

  async enterPassword(pwd: string) {
    await this.page.getByLabel(/password/i).fill(pwd);
    await this.page.getByRole('button', { name: /unlock|submit/i }).click();
  }

  async getSizeLimitWarning(): Promise<string | null> {
    const warning = this.page.locator('.size-limit-warning');
    if (await warning.isVisible()) {
      return warning.textContent();
    }
    return null;
  }
}
