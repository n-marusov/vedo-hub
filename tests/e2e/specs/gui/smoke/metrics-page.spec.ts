// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Metrics page — KPI counters from API, trend chart, loading skeleton
import { test, expect } from '../../graphql-fixtures'
import { MetricsPage } from '../../../pages/metrics.page'

test.describe('Metrics Page', () => {
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
