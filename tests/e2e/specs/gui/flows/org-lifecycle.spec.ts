// Validates: REQ-NFR.SECURITY.organization-access-model
// Organizational model lifecycle — REST user-story tests
import { test, expect } from '@playwright/test';
import { OWNER_JWT } from '../../jwt-tokens';

const API = '/api/v1';

test.describe('Org Lifecycle User Stories', () => {
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
  	test('US-org.create-project: project created via REST appears in list', async ({ page }) => {
  		// First create a parent group
  		const groupRes = await page.request.post(`${API}/groups`, {
  			data: { name: 'US-ProjectGroup', description: 'Group for project test' },
  			headers: { Authorization: `Bearer ${OWNER_JWT}` },
  		});
  		expect(groupRes.status()).toBe(201);
  		const group = await groupRes.json();

  		const res = await page.request.post(`${API}/projects`, {
  			data: { name: 'US-CreateProject', description: 'Created via REST', group_id: group.data.id },
  			headers: { Authorization: `Bearer ${OWNER_JWT}` },
  		});
  		expect(res.status()).toBe(201);
  		const project = await res.json();
  		expect(project.data.name).toBe('US-CreateProject');
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
  		await page.request.delete(`${API}/groups/${group.data.id}`, {
  			headers: { Authorization: `Bearer ${OWNER_JWT}` },
  		});
  	});

  	// US-org.projects.create: Owner creates project via GUI flow
  	test('US-org.projects.create: GUI flow — create project in group, see toast, redirect to workspace', async ({ page }) => {
  		// Setup: create a group via REST for the GUI test
  		const groupRes = await page.request.post(`${API}/groups`, {
  			data: { name: 'GUI-Test-Group', description: 'Group for GUI project test', visibility: 'Private' },
  			headers: { Authorization: `Bearer ${OWNER_JWT}` },
  		});
  		expect(groupRes.status()).toBe(201);
  		const group = await groupRes.json();

  		// Navigate to projects page
  		await page.goto('/dashboard/projects');
  		await expect(page.getByRole('heading', { name: /Projects/i })).toBeVisible();

  		// Click "New project" — should navigate to page, not open dialog
  		await page.getByRole('button', { name: /New project/i }).click();
  		await expect(page).toHaveURL(/\/dashboard\/projects\/new/);

  		// Page should show "New project" title
  		await expect(page.getByRole('heading', { name: /New project/i })).toBeVisible();

  		// Select the group by human-readable name
  		await page.selectOption('#cpp-group-select', group.data.id);

  		// Fill project name
  		await page.fill('#cpp-project-name', 'GUI Test Project');

  		// Select visibility (default is Private)
  		// Keep default Private

  		// Submit
  		await page.getByRole('button', { name: /Create project/i }).click();

  		// Should see success toast and redirect to workspace
  		await expect(page).toHaveURL(/\/project\/[^/]+\/workspace/);

  		// Navigate back to projects to verify project appears in list
  		await page.goto('/dashboard/projects');
  		await expect(page.getByText('GUI Test Project')).toBeVisible();

  		// Cleanup: delete the project via API
  		const currentUrl = page.url();
  		const projectIdMatch = currentUrl.match(/\/project\/([^/]+)\/workspace/);
  		// Fallback: list and find by name
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
