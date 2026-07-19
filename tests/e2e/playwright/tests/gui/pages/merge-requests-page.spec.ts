// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Merge Requests page — sections, toggle, filter tabs
import { test, expect } from '../../graphql-fixtures'
import { MergeRequestsPage } from '../../../pages/merge-requests.page'

test.describe('Merge Requests Page', () => {
  test('should render merge request sections from API when page loads', async ({ page }) => {
    const mr = new MergeRequestsPage(page)
    await mr.goto()
    const sections = mr.getSections()
    await expect(sections.first()).toBeVisible()
  })

  test('should toggle section content when section header is clicked', async ({ page }) => {
    const mr = new MergeRequestsPage(page)
    await mr.goto()
    const firstHeader = page.locator('.mrc-header').first()
    const headerText = await firstHeader.textContent()
    if (headerText) {
      // Sections start open:true; first click closes, second re-opens
      await mr.toggleSection(headerText.trim())
      await mr.toggleSection(headerText.trim())
      const content = page.locator('.mrc-body').first()
      await expect(content).toBeVisible()
    }
  })

  test('should switch tabs and show filtered content', async ({ page }) => {
    const mr = new MergeRequestsPage(page)
    await mr.goto()
    await mr.switchTab('active')
    await expect(page.locator('.mr-card').first()).toBeVisible()
  })
})
