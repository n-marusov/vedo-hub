// Validates: REQ-FUN.ORG.group-crud
// GroupsPage POM — hierarchy, expand/collapse, search, create group
import { type Page, expect } from '@playwright/test';

export class GroupsPage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/groups');
    await this.page.waitForSelector('.groups-page', { state: 'visible' });
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
    // Use evaluate + element.click() which reliably triggers Vue handlers in Docker
    await this.page.evaluate(() => {
      const btn = document.querySelector('.gp-new-btn') as HTMLButtonElement;
      if (btn) btn.click();
    });
  }

  async fillGroupName(name: string) {
    await this.page.locator('#group-name').fill(name);
  }

  async selectVisibility(visibility: 'private' | 'internal' | 'public') {
    // Use value attribute selector to pick the exact radio option
    await this.page.locator(`input[name="group-visibility"][value="${visibility}"]`).check();
  }

  async clickCreate() {
    await this.page.locator('.btn--primary').click();
  }

  async clickCancel() {
    await this.page.locator('.dialog__footer').getByRole('button', { name: /cancel/i }).click();
  }

  async submitFormViaBrowser() {
    await this.page.evaluate(() => {
      const form = document.querySelector('.create-group-form') as HTMLFormElement;
      if (form) form.requestSubmit();
    });
  }

  async createGroup(name: string, visibility: 'private' | 'internal' | 'public' = 'private') {
    await this.clickNewGroup();
    await this.page.waitForSelector('.dialog-overlay', { state: 'visible', timeout: 5000 });
    await this.fillGroupName(name);
    await this.selectVisibility(visibility);
    await this.submitFormViaBrowser();
  }

  getDialogOverlay() {
    return this.page.locator('.dialog-overlay');
  }

  getCreateDialog() {
    return this.page.locator('.create-group-form');
  }

  getValidationError() {
    return this.page.locator('.form-error');
  }

  getGroupByName(name: string) {
    return this.page.locator('.gp-row', { hasText: name });
  }

  getGlobeIcons() {
    return this.page.locator('.gp-row-vis-icon');
  }
}
