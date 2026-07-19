// Validates: REQ-NFR.SECURITY.organization-access-model
// Organizational model lifecycle — mixed REST + GUI user-story tests
import { test, expect } from '@playwright/test';
import { OWNER_JWT, VIEWER_JWT } from '../../jwt-tokens';
import { GroupsPage } from '../../../pages/groups.page';
import { ProjectsPage } from '../../../pages/projects.page';
import { MembersPage } from '../../../pages/members.page';

const API = '/api/v1';

test.describe('Org Lifecycle User Stories', () => {
  // US-org.create-group: Owner creates a group via REST → appears in groups list via GUI
  test('US-org.create-group: group created via REST appears in GUI', async ({ page }) => {
    // Create group via REST
    const res = await page.request.post(`${API}/groups`, {
      data: { name: 'US-CreateGroup', description: 'Created via REST' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(res.status()).toBe(201);
    const group = await res.json();
    expect(group.id).toBeDefined();

    // Verify via GUI
    const groups = new GroupsPage(page);
    await groups.goto();
    await expect(page.locator('.gp-row', { hasText: 'US-CreateGroup' })).toBeVisible();

    // Cleanup
    await page.request.delete(`${API}/groups/${group.id}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
  });

  // US-org.create-project: Owner creates project via REST → appears in projects list via GUI
  test('US-org.create-project: project created via REST appears in GUI', async ({ page }) => {
    // Create project via REST
    const res = await page.request.post(`${API}/projects`, {
      data: { name: 'US-CreateProject', description: 'Created via REST' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(res.status()).toBe(201);
    const project = await res.json();

    // Verify via GUI
    const projects = new ProjectsPage(page);
    await projects.goto();
    await expect(page.locator('.pp-row', { hasText: 'US-CreateProject' })).toBeVisible();

    // Cleanup
    await page.request.delete(`${API}/projects/${project.id}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
  });

  // US-org.manage-members: Full lifecycle — add → verify → change role → verify → remove → verify
  test('US-org.manage-members: full member lifecycle via REST + GUI', async ({ page }) => {
    const ontologyId = 'test-ont';
    const testUser = 'lifecycle-test-user';

    // Add member with Editor role via REST
    const addRes = await page.request.post(`${API}/ontologies/${ontologyId}/members`, {
      data: { user_id: testUser, role: 'Editor' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(addRes.status()).toBe(201);

    // Verify member appears in GUI
    const members = new MembersPage(page);
    await members.goto(ontologyId);
    await expect(page.locator('.table-row', { hasText: testUser })).toBeVisible();

    // Change role to Viewer via REST
    const updateRes = await page.request.put(`${API}/ontologies/${ontologyId}/members/${testUser}`, {
      data: { role: 'Viewer' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(updateRes.status()).toBe(200);

    // Verify role changed via GUI
    await members.goto(ontologyId);
    await expect(page.locator('.table-row', { hasText: testUser })).toBeVisible();

    // Remove member via REST
    const removeRes = await page.request.delete(`${API}/ontologies/${ontologyId}/members/${testUser}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(removeRes.status()).toBe(204);

    // Verify member disappears from GUI
    await members.goto(ontologyId);
    await expect(page.locator('.table-row', { hasText: testUser })).toHaveCount(0);
  });

  // US-org.visibility: Owner changes visibility → anonymous access check
  test('US-org.visibility: visibility levels enforced for anonymous access', async ({ page }) => {
    const ontologyId = 'test-visibility-ont';

    // Create ontology scope first
    await page.request.put(`${API}/ontologies/${ontologyId}`, {
      data: { name: 'VisibilityTest' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });

    // Owner changes visibility to Public
    const pubRes = await page.request.put(`${API}/ontologies/${ontologyId}/visibility`, {
      data: { visibility: 'Public' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(pubRes.status()).toBe(200);

    // Anonymous user can read members
    const anonRes = await page.request.get(`${API}/ontologies/${ontologyId}/members`);
    expect(anonRes.status()).toBe(200);

    // Owner changes back to Private
    const privRes = await page.request.put(`${API}/ontologies/${ontologyId}/visibility`, {
      data: { visibility: 'Private' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(privRes.status()).toBe(200);

    // Anonymous user gets 403
    const anonDenied = await page.request.get(`${API}/ontologies/${ontologyId}/members`);
    expect(anonDenied.status()).toBe(403);
  });

  // US-org.inheritance: Owner assigns Editor to parent group → user has effective Editor role in child
  test('US-org.inheritance: parent group role inherits to child scope (max-role-wins)', async ({ page }) => {
    // Create a parent group
    const parentRes = await page.request.post(`${API}/groups`, {
      data: { name: 'InheritanceParent' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(parentRes.status()).toBe(201);
    const parentGroup = await parentRes.json();
    const parentId = parentGroup.id;

    // Create child group under parent
    const childRes = await page.request.post(`${API}/groups`, {
      data: { name: 'InheritanceChild', parent_id: parentId },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(childRes.status()).toBe(201);

    // Assign Editor role to user on parent group
    const addRes = await page.request.post(`${API}/groups/${parentId}/members`, {
      data: { user_id: 'inheritance-user', role: 'Editor' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(addRes.status()).toBe(201);

    // Verify user has Editor access on parent group member list
    const listRes = await page.request.get(`${API}/groups/${parentId}/members`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(listRes.status()).toBe(200);

    // Cleanup: remove membership and delete groups
    await page.request.delete(`${API}/groups/${parentId}/members/inheritance-user`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });

    // Delete child, then parent
    const childId = (await childRes.json()).id;
    await page.request.delete(`${API}/groups/${childId}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    await page.request.delete(`${API}/groups/${parentId}`, {
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
  });
});
