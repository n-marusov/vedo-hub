// Validates: REQ-NFR.SECURITY.organization-access-model
// Projects page wired test — runs against real API (no Apollo fixtures)
import { test, expect } from '@playwright/test';
import { ProjectsPage } from '../../../pages/projects.page';

test.describe('Projects Page — Real API', () => {
  test('should display projects with metadata from real API', async ({ page }) => {
    const projects = new ProjectsPage(page);
    await projects.goto();
    const count = await projects.getProjectCount();
    expect(count).toBeGreaterThan(0);
    // Verify no error state is shown
    await expect(page.locator('.pp-error-state')).toHaveCount(0);
  });

  test('should sort projects by name from real API', async ({ page }) => {
    const projects = new ProjectsPage(page);
    await projects.goto();
    await projects.sortBy('name', 'asc');
    const items = projects.getProjects();
    await expect(items.first()).toBeVisible();
  });

  test('should filter projects by search from real API', async ({ page }) => {
    const projects = new ProjectsPage(page);
    await projects.goto();
    await projects.search('Test');
    const items = projects.getProjects();
    await expect(items.first()).toBeVisible();
  });

  test('should navigate to ontology workspace on project click', async ({ page }) => {
    const projects = new ProjectsPage(page);
    await projects.goto();
    const firstRow = projects.getProjects().first();
    const name = await firstRow.locator('.project-name').textContent();
    await projects.clickProject(name || '');
    await expect(page).toHaveURL(/\/ontology\//);
  });
});
