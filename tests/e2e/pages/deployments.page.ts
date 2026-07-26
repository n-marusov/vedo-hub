import { type Page, expect } from '@playwright/test';

// Page Object Model for Deployments page
// Covers: screen 15 — Deployment cards with show/hide and delete

export class DeploymentsPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/deployments');
  }

  getDeployments() {
    return this.page.locator('.deployment-card');
  }

  async toggleShowStopped() {
    await this.page.locator('.dp-checkbox').click();
  }

  async deleteDeployment(url: string) {
    await this.page.locator('.deployment-card', { hasText: url }).locator('.dc-delete-btn').click();
  }
}
