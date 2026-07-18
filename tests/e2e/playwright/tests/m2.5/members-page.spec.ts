// @ctx: M2.5 Members page — list with roles, inline edit, remove with confirmation, last-owner protection
import { test, expect } from '@playwright/test'
import { MembersPage } from '../../pages/members.page'

test.describe('M2.5 Members Page', () => {
  test('should render member list from API when page loads', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    const items = members.getMembers()
    await expect(items.first()).toBeVisible()
  })

  test('should change member role when role select is changed', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    await members.editRole('alice', 'editor')
    await expect(page.getByText(/role updated/i)).toBeVisible()
  })

  test('should show confirmation dialog before removing a member', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    await members.removeMember('bob')
    await expect(page.getByText(/confirm/i)).toBeVisible()
  })

  test('should prevent removing the last owner', async ({ page }) => {
    const members = new MembersPage(page)
    await members.goto('ont-123')
    const lastOwner = page.locator('.member-row.owner').last()
    await lastOwner.locator('.remove-member-btn').click()
    await expect(page.getByText(/cannot remove last owner/i)).toBeVisible()
  })
})
