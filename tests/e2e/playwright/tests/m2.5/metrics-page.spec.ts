// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// @ctx: M2.5 Metrics page — KPI counters from API, trend chart, loading skeleton
import { test, expect } from '../m2.5-fixtures'
import { MetricsPage } from '../../pages/metrics.page'

test.describe('M2.5 Metrics Page', () => {
  test('should render KPI counters from API when page loads', async ({ page }) => {
    const metrics = new MetricsPage(page)
    await metrics.goto('ont-123')
    const counters = metrics.getKpiCounters()
    await expect(counters.first()).toBeVisible()
  })

  test('should render trend chart when metrics data is available', async ({ page }) => {
    const metrics = new MetricsPage(page)
    await metrics.goto('ont-123')
    await expect(page.locator('.trend-chart, .metrics-chart')).toBeVisible()
  })

  test('should show loading skeleton while data is being fetched', async ({ page }) => {
    const metrics = new MetricsPage(page)
    await metrics.goto('ont-123')
    await expect(page.locator('.skeleton, .loading-skeleton')).toBeVisible({ timeout: 2000 })
  })
})
