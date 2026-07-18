import { type Page, expect } from '@playwright/test';

// Page Object Model for Versioning page
// Covers: screens 10–14 — Commits, Branches, Tags, Graph, Compare

export class VersioningPage {
  constructor(public readonly page: Page) {}

  async goto(ontologyId?: string, view?: string) {
    const id = ontologyId || 'test';
    const v = view || 'commits';
    await this.page.goto(`/ontology/${id}/versioning/${v}`);
  }

  async switchTab(tab: string) {
    await this.page.getByRole('tab', { name: new RegExp(tab, 'i') }).click();
  }

  getCommits() {
    return this.page.locator('.commit-item');
  }

  getBranches() {
    return this.page.locator('.branch-item');
  }

  getTags() {
    return this.page.locator('.tag-item');
  }

  getGraphNodes() {
    return this.page.locator('.versioning-graph-node, .repository-graph-node');
  }
}
