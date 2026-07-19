import { type Page, expect } from '@playwright/test';

// Page Object Model for Metrics page
// Covers: screen 8 — KPI counters and trend charts

export class MetricsPage {
  constructor(public readonly page: Page) {}

  async goto(ontologyId?: string) {
    const id = ontologyId || 'test';
    await this.page.goto(`/metrics?ontologyId=${id}`);
  }

  getKpiCounters() {
    return this.page.locator('.kpi-card');
  }
}
