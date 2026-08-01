// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Validates: REQ-FUN.ORG.group-crud
// Groups page — hierarchy, expand/collapse, search, create group
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

test.describe('Create Group', () => {
  test('should create a private group and show it in the list', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()

    // "New group" navigates to the create page (not a dialog)
    await groups.clickNewGroup()
    await expect(page).toHaveURL(/\/dashboard\/groups\/new/)

    // Fill the create-group form
    await page.locator('.cgp-input').fill('Research Team')

    // Submit
    await page.locator('.cgp-btn-create').click()

    // Should redirect to the new group detail page, then back to list shows the group
    await expect(page).toHaveURL(/\/dashboard\/groups\/[^/]+/, { timeout: 5000 })
    await page.goto('/dashboard/groups')
    await expect(groups.getGroupByName('Research Team')).toBeVisible()
  })

  test('should show validation error when creating group with empty name', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()

    // Navigate to create page
    await groups.clickNewGroup()
    await expect(page).toHaveURL(/\/dashboard\/groups\/new/)

    // Submit empty form
    await page.locator('.cgp-btn-create').click()

    // Validation error should appear (Russian locale)
    await expect(page.locator('.cgp-error-text')).toBeVisible()
  })

  test('should create a group with public visibility', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()

    await groups.clickNewGroup()
    await expect(page).toHaveURL(/\/dashboard\/groups\/new/)

    // Fill name and pick public visibility
    await page.locator('.cgp-input').fill('Open Research')
    await page.locator('input.cgp-vis-radio[value="public"]').check()

    await page.locator('.cgp-btn-create').click()

    // Should redirect to the new group detail page
    await expect(page).toHaveURL(/\/dashboard\/groups\/[^/]+/, { timeout: 5000 })
    await page.goto('/dashboard/groups')
    await expect(groups.getGroupByName('Open Research')).toBeVisible()
  })

  test('should close create page on cancel', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()

    await groups.clickNewGroup()
    await expect(page).toHaveURL(/\/dashboard\/groups\/new/)

    // Cancel returns to the groups list
    await page.locator('.cgp-btn-cancel').click()
    await expect(page).toHaveURL(/\/dashboard\/groups$/)
  })
})
