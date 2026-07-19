// Validates: REQ-NFR.SECURITY.organization-access-model
// Groups page wired test — runs against real API (no Apollo fixtures)
import { test, expect } from '@playwright/test';
import { GroupsPage } from '../../../pages/groups.page';

test.describe('Groups Page — Real API', () => {
  test('should display groups from real API after M2.1 backend is wired', async ({ page }) => {
    const groups = new GroupsPage(page);
    await groups.goto();
    const count = await groups.getGroupCount();
    expect(count).toBeGreaterThan(0);
    // Verify no error state is shown
    await expect(page.locator('.gp-error-state')).toHaveCount(0);
  });

  test('should filter groups by search query from real API', async ({ page }) => {
    const groups = new GroupsPage(page);
    await groups.goto();
    await groups.search('Engineering');
    const items = groups.getGroups();
    await expect(items.first()).toBeVisible();
  });

  test('should expand group and show child subgroups from real API', async ({ page }) => {
    const groups = new GroupsPage(page);
    await groups.goto();
    await groups.expandGroup('Engineering');
    const children = groups.getChildGroups();
    await expect(children.first()).toBeVisible();
  });

  test('should show visibility icon for each group', async ({ page }) => {
    const groups = new GroupsPage(page);
    await groups.goto();
    const icons = groups.getVisibilityIcons();
    await expect(icons.first()).toBeVisible();
  });
});
