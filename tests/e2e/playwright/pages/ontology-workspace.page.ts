import { type Page, expect } from '@playwright/test';

// Page Object Model for ontology workspace
// Covers: F1 (browse), F2 (TBox editor), F3 (versioning), F4 (ABox)

export class OntologyWorkspacePage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/');
  }

  async openOntology(name: string) {
    await this.page.goto(`/project/${name}/workspace`);
  }

  async createClass(label: string, parents: string[], comment?: string) {
    await this.page.getByRole('button', { name: /create class/i }).click();
    await this.page.getByLabel(/class name/i).fill(label);
    if (comment) {
      await this.page.getByLabel(/description/i).fill(comment);
    }
    for (const parent of parents) {
      await this.page.getByLabel(/add parent/i).fill(parent);
      await this.page.getByRole('option', { name: parent }).click();
    }
    await this.page.getByRole('button', { name: /save/i }).click();
  }

  async getClassTree(): Promise<string[]> {
    const treeItems = await this.page.locator('.class-tree-item').allTextContents();
    return treeItems;
  }

  async selectClass(label: string) {
    await this.page.locator('.class-tree-item', { hasText: label }).click();
  }

  async createDatatypeProperty(label: string, domain: string, xsdType: string) {
    await this.page.getByRole('button', { name: /create property/i }).click();
    await this.page.getByLabel(/property name/i).fill(label);
    await this.page.getByLabel(/type/i).selectOption('datatype');
    await this.page.getByLabel(/domain/i).fill(domain);
    await this.page.getByLabel(/xsd type/i).selectOption(xsdType);
    await this.page.getByRole('button', { name: /save/i }).click();
  }

  async createObjectProperty(label: string, domain: string, range: string) {
    await this.page.getByRole('button', { name: /create property/i }).click();
    await this.page.getByLabel(/property name/i).fill(label);
    await this.page.getByLabel(/type/i).selectOption('object');
    await this.page.getByLabel(/domain/i).fill(domain);
    await this.page.getByLabel(/range/i).fill(range);
    await this.page.getByRole('button', { name: /save/i }).click();
  }

  async createIndividual(classLabel: string, individualLabel: string) {
    await this.page.getByRole('button', { name: /create individual/i }).click();
    await this.page.getByLabel(/class/i).fill(classLabel);
    await this.page.getByLabel(/individual name/i).fill(individualLabel);
    await this.page.getByRole('button', { name: /save/i }).click();
  }

  async setPropertyValue(propertyLabel: string, value: string) {
    await this.page.locator(`[data-property="${propertyLabel}"]`).fill(value);
    await this.page.getByRole('button', { name: /save/i }).click();
  }

  async createCommit(message: string) {
    await this.page.getByRole('button', { name: /commit/i }).click();
    await this.page.getByLabel(/commit message/i).fill(message);
    await this.page.getByRole('button', { name: /confirm/i }).click();
  }

  async createBranch(name: string) {
    await this.page.getByRole('button', { name: /branch/i }).click();
    await this.page.getByLabel(/branch name/i).fill(name);
    await this.page.getByRole('button', { name: /create/i }).click();
  }

  async switchBranch(name: string) {
    await this.page.getByRole('combobox', { name: /branch/i }).selectOption(name);
  }

  async getCommitHistory(): Promise<string[]> {
    return this.page.locator('.commit-item').allTextContents();
  }

  async rollbackToCommit(commitIndex: number) {
    await this.page.locator('.commit-item').nth(commitIndex).click();
    await this.page.getByRole('button', { name: /rollback/i }).click();
    await this.page.getByRole('button', { name: /confirm/i }).click();
  }

  async getGraphNodes(): Promise<string[]> {
    return this.page.locator('.graph-node').allTextContents();
  }

  async getGraphEdges(): Promise<number> {
    return this.page.locator('.graph-edge').count();
  }
}
