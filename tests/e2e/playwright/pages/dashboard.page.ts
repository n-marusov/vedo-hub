import { type Page, expect } from '@playwright/test';

// Page Object Model for Dashboard page
// Covers: screen 2 — Dashboard with widgets, activity feed, recent ontologies

export class DashboardPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/home');
  }

  getWidgets() {
    return this.page.locator('.widget.card');
  }

  getAttentionItems() {
    return this.page.locator('.attention-item');
  }

  getActivityFeed() {
    return this.page.locator('.activity-card');
  }

  getRecentOntologies() {
    return this.page.locator('.onto-item');
  }

  async clickOntology(name: string) {
    await this.page.locator('.onto-item', { hasText: name }).click();
  }

  async setStatus(text: string) {
    await this.page.getByLabel(/status/i).fill(text);
  }

  async toggleActivityFilter(mode: string) {
    await this.page.getByRole('button', { name: new RegExp(mode, 'i') }).click();
  }
}
