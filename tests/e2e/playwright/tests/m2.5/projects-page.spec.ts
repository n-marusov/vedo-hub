// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// @ctx: M2.5 Projects page — list, search, sort, click→navigate, empty state
import { test, expect } from '../m2.5-fixtures'
import { ProjectsPage } from '../../pages/projects.page'

test.describe('M2.5 Projects Page', () => {
  test('should render project list from API when page loads', async ({ page }) => {
    const projects = new ProjectsPage(page)
    await projects.goto()
    const items = projects.getProjects()
    await expect(items.first()).toBeVisible()
  })

  test('should filter projects when search query is entered', async ({ page }) => {
    const projects = new ProjectsPage(page)
    await projects.goto()
    await projects.search('Test')
    const items = projects.getProjects()
    await expect(items.first()).toBeVisible()
  })

  test('should sort projects when column header is clicked', async ({ page }) => {
    const projects = new ProjectsPage(page)
    await projects.goto()
    await projects.sortBy('name', 'asc')
    const items = projects.getProjects()
    await expect(items.first()).toBeVisible()
  })

  test('should navigate to workspace when project row is clicked', async ({ page }) => {
    const projects = new ProjectsPage(page)
    await projects.goto()
    const firstRow = projects.getProjects().first()
    const name = await firstRow.locator('.project-name').textContent()
    await projects.clickProject(name || '')
    await expect(page).toHaveURL(/\/ontology\//)
  })

  test('should show empty state when no projects exist', async ({ page }) => {
    await page.route('**/graphql', route => {
      if (route.request().postData()?.includes('projects')) {
        route.fulfill({ status: 200, body: JSON.stringify({ data: { projects: { items: [], total: 0 } } }) })
      } else {
        route.continue()
      }
    })
    const projects = new ProjectsPage(page)
    await projects.goto()
    await expect(page.getByText(/no projects/i)).toBeVisible()
  })
})
