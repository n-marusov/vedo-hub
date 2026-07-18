// @ctx: M2.5 Merge Requests page — sections, toggle, filter tabs
import { test, expect } from '../m2.5-fixtures'
import { MergeRequestsPage } from '../../pages/merge-requests.page'

test.describe('M2.5 Merge Requests Page', () => {
  test('should render merge request sections from API when page loads', async ({ page }) => {
    const mr = new MergeRequestsPage(page)
    await mr.goto()
    const sections = mr.getSections()
    await expect(sections.first()).toBeVisible()
  })

  test('should toggle section content when section header is clicked', async ({ page }) => {
    const mr = new MergeRequestsPage(page)
    await mr.goto()
    const firstHeader = page.locator('.merge-request-section-header').first()
    const headerText = await firstHeader.textContent()
    if (headerText) {
      await mr.toggleSection(headerText.trim())
      const content = page.locator('.merge-request-section-content')
      await expect(content).toBeVisible()
    }
  })

  test('should switch tabs and show filtered content', async ({ page }) => {
    const mr = new MergeRequestsPage(page)
    await mr.goto()
    await mr.switchTab('open')
    await expect(page.locator('.merge-request-section')).toBeVisible()
  })
})
