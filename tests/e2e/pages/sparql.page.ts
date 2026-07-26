import { type Page, expect } from '@playwright/test';

// Page Object Model for SPARQL Query page
// Covers: screen 4 — SPARQL query editor, execution, and results

export class SparqlPage {
  constructor(public readonly page: Page) {}

  async goto(ontologyId?: string) {
    const id = ontologyId || 'test';
    		await this.page.goto(`/project/${id}/query`);
  }

  async enterQuery(text: string) {
    const editor = this.page.locator('.sparql-editor textarea, .cm-editor');
    await editor.fill(text);
  }

  async runQuery() {
    await this.page.getByRole('button', { name: /run|execute/i }).click();
  }

  getResults() {
    return this.page.locator('.query-results-table, .sparql-editor__table').first()
  }

  async formatQuery() {
    await this.page.getByRole('button', { name: /format/i }).click();
  }
}
