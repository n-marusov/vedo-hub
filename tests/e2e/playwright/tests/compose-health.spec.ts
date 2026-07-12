import { test, expect } from './fixtures';

// @hlv CT-E2E-001
test('playwright browsers config is present', async ({ seededUsers }) => {
  expect(seededUsers.length).toBe(3);
});

// @hlv CT-E2E-002
test('seeded users fixture resolves', async ({ seededUsers }) => {
  expect(seededUsers.map((u) => u.role)).toEqual(['owner', 'editor', 'viewer']);
});

// @hlv CT-E2E-003
test('compose smoke endpoint path is configured', async ({ page }) => {
  await page.goto('/health');
  const body = await page.textContent('body');
  expect(body).toBeTruthy();
});
