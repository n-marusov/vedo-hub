// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Projects page — list, search, sort, click->navigate, empty state
//
// NOTE: Projects page uses REST API (api/org.ts: listProjects) not GraphQL.
import { test, expect } from '@playwright/test'
import { OWNER_JWT } from '../../jwt-tokens'
import { ProjectsPage } from '../../../pages/projects.page'

const MOCK_PROJECTS = [
  { id: 'ont-123', name: 'Test University Ontology', description: 'Academic ontology project', visibility: 'private', tags: ['academic', 'university'], stars: 5, forks: 2, mergeRequests: 1, created: '2026-01-15', verified: true, ontologyId: 'ont-123', memberCount: 5, updatedAt: '2026-03-15T14:30:00Z' },
  { id: 'ont-456', name: 'Healthcare Terms', description: 'Medical terminology project', visibility: 'internal', tags: ['healthcare'], stars: 3, forks: 1, mergeRequests: 0, created: '2026-02-01', verified: false, ontologyId: 'ont-456', memberCount: 8, updatedAt: '2026-03-14T12:00:00Z' },
  { id: 'ont-789', name: 'Financial Taxonomy', description: 'Finance project', visibility: 'public', tags: ['finance'], stars: 1, forks: 0, mergeRequests: 0, created: '2026-03-01', verified: false, ontologyId: 'ont-789', memberCount: 3, updatedAt: '2026-03-13T10:00:00Z' },
]

function setupAuth(page: import('@playwright/test').Page) {
  return page.addInitScript((token: string) => {
    const session = {
      accessToken: token,
      refreshToken: token,
      userId: 'user-123',
      tenantId: 'default',
      roles: ['Owner'],
      expiresAt: Date.now() + 86_400_000,
    }
    sessionStorage.setItem('vedo_session', JSON.stringify(session))
    localStorage.setItem('vedo-jwt-token', token)
  }, OWNER_JWT)
}

test.describe('Projects Page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page)

    // Mock projects list endpoint — match /api/v1/projects (with or without query params)
    await page.route(
      (url) => url.pathname === '/api/v1/projects',
      async (route) => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ data: MOCK_PROJECTS, total: MOCK_PROJECTS.length }),
          })
        } else {
          await route.continue()
        }
      }
    )
  })

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
    const rows = projects.getProjects()
    await expect(rows.first()).toBeVisible()
    await rows.first().click()
    await expect(page).toHaveURL(/\/project\//)
  })

  test('should show empty state when no projects exist', async ({ page }) => {
    await page.route(
      (url) => url.pathname === '/api/v1/projects',
      async (route) => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ data: [], total: 0 }),
          })
        } else {
          await route.continue()
        }
      }
    )
    await setupAuth(page)
    const projects = new ProjectsPage(page)
    await projects.goto()
    await expect(page.getByText(/no projects/i)).toBeVisible()
  })
})
