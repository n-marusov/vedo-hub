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
    // After clicking edit, select the role in the select that appears
    await this.page.locator('.table-row', { hasText: member }).locator('.role-select').selectOption(role);
  }

  async removeMember(member: string) {
    await this.page.locator('.table-row', { hasText: member }).getByRole('button', { name: /remove/i }).click();
    await this.page.getByRole('button', { name: /confirm/i }).click();
  }

  getMemberCount(): Promise<number> {
    return this.page.locator('.table-row').count();
  }

  async addMember(username: string, role: string) {
    await this.page.getByRole('button', { name: /add member/i }).click();
    await this.page.locator('.add-member-username').fill(username);
    await this.page.locator('.add-member-role').selectOption(role);
    await this.page.getByRole('button', { name: /save|add/i }).click();
  }

  async getMemberRole(username: string): Promise<string | null> {
    const roleEl = this.page.locator('.table-row', { hasText: username }).locator('.role-badge');
    return roleEl.textContent();
  }
}
