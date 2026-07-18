import { type Page, expect } from '@playwright/test';

// Page Object Model for Members page
// Covers: screen 7 — Members list with roles, inline edit, remove

export class MembersPage {
  constructor(public readonly page: Page) {}

  async goto(ontologyId?: string) {
    const path = ontologyId ? `/ontology/${ontologyId}/members` : '/ontology/test/members';
    await this.page.goto(path);
  }

  getMembers() {
    return this.page.locator('.table-row');
  }

  async editRole(member: string, role: string) {
    await this.page.locator('.table-row', { hasText: member }).getByRole('button', { name: /edit/i }).click();
    // After clicking edit, select the role in the dialog/select that appears
    await this.page.locator('.table-row', { hasText: member }).locator('.role-pill').click();
  }

  async removeMember(member: string) {
    await this.page.locator('.table-row', { hasText: member }).getByRole('button', { name: /remove/i }).click();
    await this.page.getByRole('button', { name: /confirm/i }).click();
  }
}
