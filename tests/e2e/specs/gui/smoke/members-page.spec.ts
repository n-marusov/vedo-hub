// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Members page — list with roles, inline edit, remove with confirmation, last-owner protection
import { test, expect } from '../../graphql-fixtures'
import { MembersPage } from '../../../pages/members.page'

test.describe('Members Page', () => {
  test('should render member list from API when page loads', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    const items = members.getMembers()
    await expect(items.first()).toBeVisible()
  })

  test('should change member role when role select is changed', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    await members.editRole('editor_seed', 'Editor')
    await expect(page.getByText(/role updated/i)).toBeVisible()
  })

  test('should show confirmation dialog before removing a member', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    // Click remove button to open the confirmation dialog
    await page.locator('.table-row', { hasText: 'viewer_seed' }).getByRole('button', { name: /remove/i }).click()
    await expect(page.getByText(/confirm/i)).toBeVisible()
  })

  test('should prevent removing the last owner', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    const lastOwner = page.locator('.table-row').filter({ hasText: 'owner_seed' }).last()
    const removeBtn = lastOwner.getByRole('button', { name: /remove/i })
    await expect(removeBtn).toBeDisabled()
    await expect(removeBtn).toHaveAttribute('title', /cannot remove last owner/i)
  })
})
