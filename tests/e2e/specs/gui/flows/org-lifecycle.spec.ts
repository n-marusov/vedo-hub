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
    const res = await page.request.post(`${API}/projects`, {
      data: { label: 'US-CreateProject', description: 'Created via REST' },
      headers: { Authorization: `Bearer ${OWNER_JWT}` },
    });
    expect(res.status()).toBe(201);
    const project = await res.json();

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
});
