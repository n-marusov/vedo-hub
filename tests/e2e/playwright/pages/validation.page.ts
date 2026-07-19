import { type Page, expect } from '@playwright/test';

// Page Object Model for Validation page
// Covers: screen 9 — SHACL validation with run, results, and summary

export class ValidationPage {
  constructor(public readonly page: Page) {}

  async goto(ontologyId?: string) {
    const id = ontologyId || 'test';
    await this.page.goto(`/ontology/${id}/validation`);
  }

  async runValidation() {
    await this.page.getByRole('button', { name: /run|validate/i }).click();
  }

  getSummary() {
    return this.page.locator('.validation-report__summary');
  }

  getResults() {
    return this.page.locator('.validation-report__table, .validation-report__empty').first()
  }
}
