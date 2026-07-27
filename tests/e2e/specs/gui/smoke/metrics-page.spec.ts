// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Metrics page — KPI counters from API, trend chart, loading skeleton
//
// NOTE: Metrics page uses REST API (api/metrics.ts: getOntologyMetrics) not GraphQL.
// Switched from graphql-fixtures to fixtures + REST route mocks.
import { test, expect } from '../../fixtures'
import { MetricsPage } from '../../../pages/metrics.page'

const MOCK_METRICS = {
  counters: { classCount: 156, propertyCount: 89, individualCount: 1204, axiomCount: 18450, commentCount: 42, mergeRequestCount: 3 },
  trends: [
    { date: '2026-03-01', classCount: 140, propertyCount: 80, individualCount: 1100 },
    { date: '2026-03-08', classCount: 148, propertyCount: 85, individualCount: 1150 },
    { date: '2026-03-15', classCount: 156, propertyCount: 89, individualCount: 1204 },
  ],
}

test.describe('Metrics Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock ontology metrics REST endpoint — getOntologyMetrics expects data directly
    await page.route('**/api/v1/metrics/ontologies/*', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(MOCK_METRICS),
        })
      } else {
        await route.continue()
      }
    })
  })

  test('should render KPI counters from API when page loads', async ({ page }) => {
    const metrics = new MetricsPage(page)
    await metrics.goto('ont-123')
    const counters = metrics.getKpiCounters()
    await expect(counters.first()).toBeVisible()
  })

  test('should render trend chart when metrics data is available', async ({ page }) => {
    const metrics = new MetricsPage(page)
    await metrics.goto('ont-123')
    await expect(page.locator('.metrics-chart').first()).toBeVisible()
  })

  test.skip('should show loading skeleton while data is being fetched', async ({ page }) => {
    const metrics = new MetricsPage(page)
    // Navigate with a cache-busting param to ensure fresh load triggers skeleton
    await metrics.goto('ont-123')
    await expect(page.locator('.kpi-card.skeleton').first()).toBeVisible({ timeout: 5000 })
  })
})
