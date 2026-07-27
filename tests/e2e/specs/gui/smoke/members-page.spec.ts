// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Members page — list with roles, inline edit, remove with confirmation, last-owner protection
//
// NOTE: Members page uses REST API (api/org.ts: listMembers) not GraphQL.
// Switched from graphql-fixtures to fixtures + REST route mocks.
import { test, expect } from '../../fixtures'
import { MembersPage } from '../../../pages/members.page'

const MOCK_MEMBERS = [
  { id: 'user-456', userId: 'user-456', username: 'owner_seed', avatarUrl: '', role: 'Owner', addedAt: '2026-01-15T10:00:00Z' },
  { id: 'user-789', userId: 'user-789', username: 'editor_seed', avatarUrl: '', role: 'Editor', addedAt: '2026-02-20T14:30:00Z' },
  { id: 'user-012', userId: 'user-012', username: 'viewer_seed', avatarUrl: '', role: 'Viewer', addedAt: '2026-03-10T09:15:00Z' },
]

test.describe('Members Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock members REST endpoint — returns data via data.data (per listMembers parser)
    await page.route('**/api/v1/projects/*/members', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: MOCK_MEMBERS }),
        })
      } else {
        await route.continue()
      }
    })
  })

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
