import { type Page, expect } from '@playwright/test';

// Page Object Model for Projects page
// Covers: screen 5 — Projects list with search, sort, and navigation

export class ProjectsPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/projects');
  }

  async getProjects() {
    return this.page.locator('.project-row');
  }

  async search(query: string) {
    await this.page.getByPlaceholder(/search/i).fill(query);
    await this.page.keyboard.press('Enter');
  }

  async sortBy(field: string, dir: 'asc' | 'desc') {
    await this.page.getByRole('button', { name: new RegExp(field, 'i') }).click();
    if (dir === 'desc') {
      await this.page.getByRole('button', { name: new RegExp(field, 'i') }).click();
    }
  }

  async clickProject(name: string) {
    await this.page.locator('.project-row', { hasText: name }).click();
  }
}
