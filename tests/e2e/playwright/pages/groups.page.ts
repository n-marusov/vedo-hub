import { type Page, expect } from '@playwright/test';

// Page Object Model for Groups page
// Covers: screen 6 — Groups hierarchy with expand/collapse and search

export class GroupsPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/groups');
  }

  getGroups() {
    return this.page.locator('.group-row');
  }

  async expandGroup(name: string) {
    await this.page.locator('.group-row', { hasText: name }).locator('.group-expand-btn').click();
  }

  async collapseGroup(name: string) {
    await this.page.locator('.group-row', { hasText: name }).locator('.group-collapse-btn').click();
  }

  async search(query: string) {
    await this.page.getByPlaceholder(/search/i).fill(query);
    await this.page.keyboard.press('Enter');
  }
}
