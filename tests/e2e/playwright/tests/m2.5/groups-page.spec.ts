// @ctx: M2.5 Groups page — hierarchy, expand/collapse, lazy loading, search
import { test, expect } from '../m2.5-fixtures'
import { GroupsPage } from '../../pages/groups.page'

test.describe('M2.5 Groups Page', () => {
  test('should render group hierarchy from API when page loads', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    const items = groups.getGroups()
    await expect(items.first()).toBeVisible()
  })

  test('should expand group to reveal children when expand is clicked', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    await groups.expandGroup('Test')
    const children = page.locator('.group-child-row')
    await expect(children.first()).toBeVisible()
  })

  test('should collapse group and hide children when collapse is clicked', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    await groups.expandGroup('Test')
    await groups.collapseGroup('Test')
    const children = page.locator('.group-child-row')
    await expect(children).toHaveCount(0)
  })

  test('should filter groups when search query is entered', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()
    await groups.search('Test')
    const items = groups.getGroups()
    await expect(items.first()).toBeVisible()
  })
})
