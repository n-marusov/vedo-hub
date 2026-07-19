import { type Page, expect } from '@playwright/test';

// Page Object Model for Groups page
// Covers: screen 6 — Groups hierarchy with expand/collapse and search

export class GroupsPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/groups');
  }

  getGroups() {
    return this.page.locator('.gp-row');
  }

  async expandGroup(name: string) {
    await this.page.locator('.gp-row', { hasText: name }).locator('.gp-row-chevron').click();
  }

  async collapseGroup(name: string) {
    await this.page.locator('.gp-row', { hasText: name }).locator('.gp-row-chevron').click();
  }

  async search(query: string) {
    await this.page.locator('.gp-search-input').fill(query);
    await this.page.keyboard.press('Enter');
  }

  getVisibilityIcons() {
    return this.page.locator('.gp-visibility-icon');
  }

  getChildGroups() {
    return this.page.locator('.group-child-row');
  }

  getGroupCount(): Promise<number> {
    return this.page.locator('.gp-row').count();
  }

  async clickNewGroup() {
    await this.page.getByRole('button', { name: /new group/i }).click();
  }
}
