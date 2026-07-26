// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Dashboard wiring — widgets, attention items, activity feed, recent ontologies from API
//
// NOTE: Dashboard was migrated from GraphQL (DASHBOARD_QUERY) to REST.
// The REST client in api/dashboard.ts currently returns hardcoded mock data
// (no backend call). Once a real backend endpoint is implemented, these tests
// should be updated to mock the REST endpoint instead.
import { test, expect } from '../../fixtures'
import { DashboardPage } from '../../../pages/dashboard.page'

test.describe('Dashboard Wiring', () => {
  test('should render widgets from API when dashboard loads', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    const widgets = dashboard.getWidgets()
    await expect(widgets).toHaveCount(3)
    // Widget titles from REST mock in api/dashboard.ts: 'Merge requests' (x2), 'Active Comments'
    // NOTE: Two widgets share 'Merge requests' title — use .first() to avoid strict mode violation.
    await expect(widgets.getByText('Merge requests', { exact: true }).first()).toBeVisible()
    await expect(widgets.getByText('Active Comments', { exact: true })).toBeVisible()
  })

  test('should show attention items with counts from API', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    const items = dashboard.getAttentionItems()
    await expect(items.first()).toBeVisible()
  })

  test('should display activity feed loaded from API', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    const feed = dashboard.getActivityFeed()
    await expect(feed).toBeVisible()
    await expect(feed.locator('.activity-item')).not.toHaveCount(0)
  })

  test('should show recent ontologies loaded from API', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    const ontologies = dashboard.getRecentOntologies()
    await expect(ontologies.first()).toBeVisible()
  })
})
