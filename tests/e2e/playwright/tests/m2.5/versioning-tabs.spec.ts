// @ctx: M2.5 Versioning tabs — commits with data, branches, tags, graph nodes, compare diff
import { test, expect } from '../m2.5-fixtures'
import { VersioningPage } from '../../pages/versioning.page'

test.describe('M2.5 Versioning Tabs', () => {
  test('should render commit history from API when Commits tab is active', async ({ page }) => {
    const versioning = new VersioningPage(page)
    await versioning.goto('ont-123', 'commits')
    const commits = versioning.getCommits()
    await expect(commits.first()).toBeVisible()
  })

  test('should render branch list from API when Branches tab is active', async ({ page }) => {
    const versioning = new VersioningPage(page)
    await versioning.goto('ont-123', 'branches')
    const branches = versioning.getBranches()
    await expect(branches.first()).toBeVisible()
  })

  test('should render tag list from API when Tags tab is active', async ({ page }) => {
    const versioning = new VersioningPage(page)
    await versioning.goto('ont-123', 'tags')
    const tags = versioning.getTags()
    await expect(tags.first()).toBeVisible()
  })

  test('should render version graph when Graph tab is active', async ({ page }) => {
    const versioning = new VersioningPage(page)
    await versioning.goto('ont-123', 'graph')
    const nodes = versioning.getGraphNodes()
    await expect(nodes.first()).toBeVisible()
  })

  test('should show diff view when Compare tab is active', async ({ page }) => {
    const versioning = new VersioningPage(page)
    await versioning.goto('ont-123', 'compare')
    await expect(page.locator('.diff-view')).toBeVisible()
  })

  test('should switch tabs and show different content', async ({ page }) => {
    const versioning = new VersioningPage(page)
    await versioning.goto('ont-123', 'commits')
    await versioning.switchTab('branches')
    const branches = versioning.getBranches()
    await expect(branches.first()).toBeVisible()
  })

  test('should show loading state while versioning data loads', async ({ page }) => {
    const versioning = new VersioningPage(page)
    await versioning.goto('ont-123', 'commits')
    await expect(page.locator('.loading-indicator, .spinner')).toBeVisible({ timeout: 2000 })
  })
})
