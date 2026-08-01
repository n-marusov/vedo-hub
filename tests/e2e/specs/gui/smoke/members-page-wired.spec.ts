// Validates: REQ-NFR.SECURITY.organization-access-model
// Members page wired test — runs against real API (no Apollo fixtures)
import { test, expect } from '../../../fixtures/auth.fixture';
import { MembersPage } from '../../../pages/members.page';

// Seeded demo project id from deploy/postgres/init/02-seed-org-data.sql
const SEED_PROJECT_ID = '44444444-4444-4444-8444-444444444444';

test.describe('Members Page — Real API', () => {
  test('should display member list with roles from real API', async ({ page }) => {
    const members = new MembersPage(page);
    await members.goto(SEED_PROJECT_ID);
    const count = await members.getMemberCount();
    expect(count).toBeGreaterThan(0);
    // Verify no error state is shown
    await expect(page.locator('.members-error-state')).toHaveCount(0);
  });

  test('should show edit role dropdown for members', async ({ page }) => {
    const members = new MembersPage(page);
    await members.goto(SEED_PROJECT_ID);
    // Click edit on the first member row
    const editBtn = page.locator('.table-row').first().getByRole('button', { name: /edit/i });
    await editBtn.click();
    const roleSelect = page.locator('.role-select');
    await expect(roleSelect).toBeVisible();
  });

  test('should show remove button for members', async ({ page }) => {
    const members = new MembersPage(page);
    await members.goto(SEED_PROJECT_ID);
    const removeBtn = page.locator('.table-row').first().getByRole('button', { name: /remove/i });
    await expect(removeBtn).toBeVisible();
  });
});
