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

    await groups.createGroup('Research Team', 'private')

    // Dialog should close after successful creation
    await expect(groups.getDialogOverlay()).not.toBeVisible({ timeout: 5000 })

    // New group should appear in the list
    await expect(groups.getGroupByName('Research Team')).toBeVisible()
  })

  test('should show validation error when creating group with empty name', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()

    // Open dialog without entering a name
    await groups.clickNewGroup()
    await expect(groups.getDialogOverlay()).toBeVisible({ timeout: 5000 })

    // Submit empty form
    await groups.submitFormViaBrowser()

    // Validation error should appear
    await expect(groups.getValidationError()).toBeVisible()
    await expect(groups.getValidationError()).toContainText('Group name is required')

    // Dialog should remain open
    await expect(groups.getDialogOverlay()).toBeVisible()
  })

  test('should create a group with public visibility', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()

    await groups.createGroup('Open Research', 'public')

    // Dialog should close
    await expect(groups.getDialogOverlay()).not.toBeVisible({ timeout: 5000 })

    // New group should appear
    await expect(groups.getGroupByName('Open Research')).toBeVisible()
  })

  test('should close dialog on cancel', async ({ page }) => {
    const groups = new GroupsPage(page)
    await groups.goto()

    await groups.clickNewGroup()
    await expect(groups.getDialogOverlay()).toBeVisible({ timeout: 5000 })

    await groups.clickCancel()
    await expect(groups.getDialogOverlay()).not.toBeVisible()
  })
})
