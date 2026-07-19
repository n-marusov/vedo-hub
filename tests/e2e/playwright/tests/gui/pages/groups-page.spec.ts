// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Groups page — hierarchy, expand/collapse, lazy loading, search
import { test, expect } from '../../graphql-fixtures'
import { GroupsPage } from '../../../pages/groups.page'

test.describe('Groups Page', () => {
  test('should render group hierarchy from API when page loads', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    const items = groups.getGroups()
    await expect(items.first()).toBeVisible()
  })

  test('should expand group to reveal children when expand is clicked', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    await groups.expandGroup('Engineering')
    const children = page.locator('.group-child-row')
    await expect(children.first()).toBeVisible()
  })

  test('should collapse group and hide children when collapse is clicked', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    await groups.expandGroup('Engineering')
    await groups.collapseGroup('Engineering')
    const children = page.locator('.group-child-row')
    await expect(children).toHaveCount(0)
  })

  test('should filter groups when search query is entered', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    await groups.search('Engineering')
    const items = groups.getGroups()
    await expect(items.first()).toBeVisible()
  })
})
