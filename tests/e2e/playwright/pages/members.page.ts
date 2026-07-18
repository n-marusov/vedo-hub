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
    return this.page.locator('.member-row');
  }

  async editRole(member: string, role: string) {
    await this.page.locator('.member-row', { hasText: member }).locator('.role-select').selectOption(role);
  }

  async removeMember(member: string) {
    await this.page.locator('.member-row', { hasText: member }).locator('.remove-member-btn').click();
    await this.page.getByRole('button', { name: /confirm/i }).click();
  }
}
