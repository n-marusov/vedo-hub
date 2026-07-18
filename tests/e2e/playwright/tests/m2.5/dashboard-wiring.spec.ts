// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// @ctx: M2.5 Dashboard wiring — widgets, attention items, activity feed, recent ontologies from API
import { test, expect } from '../m2.5-fixtures'
import { DashboardPage } from '../../pages/dashboard.page'

test.describe('M2.5 Dashboard Wiring', () => {
  test('should render widgets from API when dashboard loads', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    const widgets = dashboard.getWidgets()
    await expect(widgets).toHaveCount(3)
    await expect(widgets.getByText('Merge Requests', { exact: true })).toBeVisible()
    await expect(widgets.getByText('Reviews', { exact: true })).toBeVisible()
    await expect(widgets.getByText('Work Items', { exact: true })).toBeVisible()
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

  test('should show error state with retry when API fails', async ({ page }) => {
    await page.route('**/api/v1/graphql', route => route.abort('connectionfailed'))
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    await expect(page.getByRole('button', { name: /retry/i })).toBeVisible()
  })
})
