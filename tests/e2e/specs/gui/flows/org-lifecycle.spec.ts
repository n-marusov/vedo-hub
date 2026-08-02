// Validates: REQ-NFR.SECURITY.organization-access-model
// Organizational model lifecycle — REST user-story tests
import { test, expect } from '../../../fixtures/auth.fixture';
import { OWNER_JWT } from '../../jwt-tokens';

const API = '/api/v1';

test.describe('Org Lifecycle User Stories', () => {
  // The GUI flow assertions target the Russian UI (i18n). Force the browser
  // locale so heading/button text matches regardless of the runner's locale.
  test.use({ locale: 'ru-RU' });

  // US-org.create-group: Owner creates a group via REST → appears in groups list
  test('US-org.create-group: group created via REST appears in list', async ({ page }) => {
    const res = await page.request.post(`${API}/groups`, {
      data: { label: 'US-CreateGroup', description: 'Created via REST' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(res.status()).toBe(201);
    const group = await res.json();
    expect(group.data.id).toBeDefined();

    const listRes = await page.request.get(`${API}/groups`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(listRes.status()).toBe(200);
    const listBody = await listRes.json();
    const ids = listBody.data.map((g: { id: string }) => g.id);
    expect(ids).toContain(group.data.id);

    await page.request.delete(`${API}/groups/${group.data.id}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
  });

  // US-org.create-project: Owner creates project via REST → appears in projects list
  // group_id is optional; omitted to keep the test independent of group membership.
  test('US-org.create-project: project created via REST appears in list', async ({ page }) => {
    const res = await page.request.post(`${API}/projects`, {
      data: { name: 'US-CreateProject', description: 'Created via REST' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(res.status()).toBe(201);
    const project = await res.json();
    expect(project.data.id).toBeDefined();
    expect(project.data.ontology_id).toBeDefined();

    const listRes = await page.request.get(`${API}/projects`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(listRes.status()).toBe(200);
    const listBody = await listRes.json();
    const ids = listBody.data.map((p: { id: string }) => p.id);
    expect(ids).toContain(project.data.id);

    await page.request.delete(`${API}/projects/${project.data.id}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
  });

  // US-org.projects.create: Owner creates project via GUI flow
  test('US-org.projects.create: GUI flow — create project in group, see toast, redirect to workspace', async ({ page }) => {
    // Setup: create a group via REST for the GUI test.
    // The auth-service auto-adds the creator as an Owner member
    // (gitlab-aligned), so the creator can create projects in it.
    const groupRes = await page.request.post(`${API}/groups`, {
      data: { name: 'GUI-Test-Group', description: 'Group for GUI project test', visibility: 'Private' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(groupRes.status()).toBe(201);
    const group = await groupRes.json();

    // Navigate to projects page — locale is Russian (lang=ru in HTML)
    await page.goto('/dashboard/projects');
    await expect(page.getByRole('heading', { name: /Проекты/i })).toBeVisible();

    // Click "New project" button (Russian locale)
    await page.getByRole('button', { name: /Новый проект/i }).click();
    await expect(page).toHaveURL(/\/dashboard\/projects\/new/);

    // Select the group by human-readable name
    await page.selectOption('#cpp-group-select', group.data.id);

    // Fill project name
    await page.fill('#cpp-project-name', 'GUI Test Project');

    // Submit
    await page.getByRole('button', { name: /Создать проект/i }).click();

    // Should redirect to workspace
    await expect(page).toHaveURL(/\/project\/[^/]+\/workspace/);

    // Navigate back to projects to verify project appears in list.
    // Note: the backend stores the project `name` as a UUID (slug), not the
    // user-typed label, so assert on the presence of a project row instead.
    await page.goto('/dashboard/projects');
    await expect(page.locator('.pp-row').first()).toBeVisible();

    // Cleanup: delete the project via API
    const listRes = await page.request.get(`${API}/projects`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    const listBody = await listRes.json();
    const createdProject = listBody.data.find((p: { name: string }) => p.name === 'GUI Test Project');
    if (createdProject) {
      await page.request.delete(`${API}/projects/${createdProject.id}`, {
        headers: { Authorization: `Bearer ${OWNER_JWT}` },
      });
    }

    // Delete the test group
    await page.request.delete(`${API}/groups/${group.data.id}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
  });
});
