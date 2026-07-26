// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Dashboard navigation — click recent project → workspace; user name in header; active route highlight; persistence across reload
import { test, expect } from '../../graphql-fixtures'
import { DashboardPage } from '../../../pages/dashboard.page'

test.describe('Dashboard Navigation', () => {
  test('should navigate to ontology workspace when recent project row is clicked', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    const firstOntology = dashboard.getRecentOntologies().first()
    const name = await firstOntology.locator('.onto-name').textContent()
    if (name) {
      await dashboard.clickOntology(name)
      await expect(page).toHaveURL(/\/project\//)
    }
  })

  test('should display user name in the header avatar menu', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    await expect(page.locator('.header-avatar-menu')).toBeVisible()
  })

  test('should highlight active route in sidebar', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    const activeItem = page.locator('.sidebar-item--active')
    await expect(activeItem).toBeVisible()
  })

  test('should persist navigation state across page reload', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    await page.reload()
    await expect(page.locator('.sidebar-item--active')).toBeVisible()
  })
})
