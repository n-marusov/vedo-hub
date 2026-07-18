import { type Page, expect } from '@playwright/test';

// Page Object Model for Dashboard page
// Covers: screen 2 — Dashboard with widgets, activity feed, recent ontologies

export class DashboardPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/home');
  }

  async getWidgets() {
    return this.page.locator('.dashboard-widget');
  }

  async getAttentionItems() {
    return this.page.locator('.attention-item');
  }

  async getActivityFeed() {
    return this.page.locator('.activity-feed');
  }

  async getRecentOntologies() {
    return this.page.locator('.recent-ontology-row');
  }

  async clickOntology(name: string) {
    await this.page.locator('.recent-ontology-row', { hasText: name }).click();
  }

  async setStatus(text: string) {
    await this.page.getByLabel(/status/i).fill(text);
  }

  async toggleActivityFilter(mode: string) {
    await this.page.getByRole('button', { name: new RegExp(mode, 'i') }).click();
  }
}
