import { type Page, type Locator } from '@playwright/test';

// Page Object Model for M2 Document Extraction and AI Ontology Creation
// Covers: US-io.document.extract-*, US-io.document.batch-extract, US-io.document.preview-sequence

export class M2DocumentUploadPage {
  constructor(public readonly page: Page) {}

  // ─── Navigation ───────────────────────────────────────────────────

  async goto() {
    await this.page.goto('/');
  }

  async openDocumentUpload(ontologyName: string) {
    await this.page.goto(`/ontology/${ontologyName}/upload`);
  }

  // ─── File Upload ──────────────────────────────────────────────────

  async uploadFile(filePath: string) {
    const fileChooserPromise = this.page.waitForEvent('filechooser');
    await this.page.getByRole('button', { name: /upload document|choose file|browse/i }).click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles([filePath]);
    // Wait for upload processing to complete
    await this.page.waitForResponse((resp) =>
      resp.url().includes('/api/v1/ontologies/') && resp.url().includes('/extract') && resp.status() === 200,
      { timeout: 15_000 }
    ).catch(() => {
      // Mock route handles this through fulfilled responses
    });
  }

  async uploadMultipleFiles(filePaths: string[]) {
    const fileChooserPromise = this.page.waitForEvent('filechooser');
    await this.page.getByRole('button', { name: /upload documents|upload files/i }).click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(filePaths);
    await this.page.waitForTimeout(2000);
  }

  // ─── Preview Sequence ─────────────────────────────────────────────

  async getPreviewSequence(): Promise<string[]> {
    return this.page.locator('.sequence-step-item').allTextContents();
  }

  async getStepLabels(): Promise<string[]> {
    return this.page.locator('.sequence-step-label').allTextContents();
  }

  async getStepCount(): Promise<number> {
    return this.page.locator('.sequence-step-item').count();
  }

  async editStepLabel(index: number, newLabel: string) {
    const step = this.page.locator('.sequence-step-item').nth(index);
    await step.locator('.step-label-edit').click();
    await step.locator('.step-label-input').fill(newLabel);
    await step.locator('.step-label-save').click();
  }

  async toggleStep(index: number) {
    await this.page.locator('.sequence-step-item').nth(index)
      .locator('.step-toggle-checkbox').click();
  }

  async toggleStepInclusion(index: number, include: boolean) {
    const checkbox = this.page.locator('.sequence-step-item').nth(index)
      .locator('.step-toggle-checkbox');
    const isChecked = await checkbox.isChecked();
    if ((include && !isChecked) || (!include && isChecked)) {
      await checkbox.click();
    }
  }

  async getStepWarning(index: number): Promise<string | null> {
    const warning = this.page.locator('.sequence-step-item').nth(index)
      .locator('.step-warning');
    if (await warning.isVisible()) {
      return warning.textContent();
    }
    return null;
  }

  // ─── Duplicate Detection ──────────────────────────────────────────

  async getDuplicateWarning(): Promise<string | null> {
    const warning = this.page.locator('.duplicate-detection-warning');
    if (await warning.isVisible()) {
      return warning.textContent();
    }
    return null;
  }

  // ─── Batch / Source Attribution ───────────────────────────────────

  async getSourceAttribution(stepIndex: number): Promise<string | null> {
    const attr = this.page.locator('.sequence-step-item').nth(stepIndex)
      .locator('.source-attribution');
    if (await attr.isVisible()) {
      return attr.textContent();
    }
    return null;
  }

  async getFileUploadStatus(fileName: string): Promise<string | null> {
    const statusEl = this.page.locator(`.file-upload-item[data-filename="${fileName}"] .upload-status`);
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
    await conflict.locator(`input[value="${strategy}"]`).check();
    await conflict.locator('.conflict-apply').click();
  }

  async resolveAllConflicts(strategy: 'keep-new' | 'keep-existing' | 'keep-both') {
    const count = await this.getConflictCount();
    for (let i = 0; i < count; i++) {
      await this.resolveConflict(0, strategy);
      await this.page.waitForTimeout(200);
    }
  }

  // ─── Apply Workflow ───────────────────────────────────────────────

  async applySequence() {
    await this.page.getByRole('button', { name: /apply/i }).click();
  }

  async getProgressBarValue(): Promise<number | null> {
    const progressBar = this.page.locator('.apply-progress-bar');
    if (await progressBar.isVisible()) {
      const value = await progressBar.getAttribute('value');
      return value ? parseInt(value, 10) : null;
    }
    return null;
  }

  async getProgressText(): Promise<string | null> {
    const text = this.page.locator('.apply-progress-text');
    if (await text.isVisible()) {
      return text.textContent();
    }
    return null;
  }

  async waitForApplyComplete(timeout = 30_000): Promise<boolean> {
    try {
      await this.page.waitForSelector('.apply-success', { timeout });
      return true;
    } catch {
      return false;
    }
  }

  async getSuccessMessage(): Promise<string | null> {
    const success = this.page.locator('.apply-success');
    if (await success.isVisible()) {
      return success.textContent();
    }
    return null;
  }

  async getCommitLink(): Promise<string | null> {
    const link = this.page.locator('.apply-success a.commit-link');
    if (await link.isVisible()) {
      return link.getAttribute('href');
    }
    return null;
  }

  async getErrorMessage(): Promise<string | null> {
    const error = this.page.locator('.apply-error');
    if (await error.isVisible()) {
      return error.textContent();
    }
    return null;
  }

  async clickRetry() {
    await this.page.getByRole('button', { name: /retry/i }).click();
  }

  // ─── Error / Edge Cases ───────────────────────────────────────────

  async getValidationError(): Promise<string | null> {
    const error = this.page.locator('.validation-error');
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
