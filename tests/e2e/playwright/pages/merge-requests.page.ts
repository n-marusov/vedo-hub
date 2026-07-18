import { type Page, expect } from '@playwright/test';

// Page Object Model for Merge Requests page
// Covers: screen 15 — Merge request sections, toggle, and filter tabs

export class MergeRequestsPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/merge_requests');
  }

  async getSections() {
    return this.page.locator('.merge-request-section');
  }

  async toggleSection(title: string) {
    await this.page.locator('.merge-request-section-header', { hasText: title }).click();
  }

  async switchTab(tab: string) {
    await this.page.getByRole('tab', { name: new RegExp(tab, 'i') }).click();
  }
}
