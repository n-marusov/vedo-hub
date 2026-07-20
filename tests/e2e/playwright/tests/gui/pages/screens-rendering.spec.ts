// Playwright tests for screen rendering with mock data per GUI contracts
import { test, expect } from '@playwright/test'

// JWT with sub=user-123, name=Alice, roles=[Editor], signed with test-jwt-key.pem
const ALICE_JWT = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsInVzZXJfaWQiOiJ1c2VyLTEyMyIsIm5hbWUiOiJBbGljZSIsInByZWZlcnJlZF91c2VybmFtZSI6ImFsaWNlIiwidGVuYW50X2lkIjoiZGVmYXVsdCIsIm9yZ2FuaXphdGlvbl9pZCI6Im9yZy0wMDEiLCJyb2xlcyI6WyJFZGl0b3IiXSwiaWF0IjoxNzg0NTQ0ODMzLCJleHAiOjE3ODQ2MzEyMzN9.qZ8vf0F_B3o_Kc3ArBYpRnCKPajvPqZ5Cvno2-y4NEM8IGqxySJysAOn6_Vq-jbI84LZASf9yUzGezhCFuggWomxNuFXxZ7JF0vhptyWfDFd2hc7Qdl8TSht7EehZoNNOiMQesJImFTIhvxTQRccVjvBZLtxyQnXAnEVXx0BcFb28Wt5hlUCWXEODpSlysLP_QVxyyAtFkV7034_1EQVCa2mvm7inxOcQK9PkwFZxCfsIjzVh-b91u31rVi8oSEPrXeycGp8yas6PERkBbBvIuNWMOv9YSE6y74Q1HpKiTvve7dnNKvqSPRfQ2NpvWo1TNAxSZAcbznShOZANrBVbw'

// Helper: set up auth session + JWT in the page, mock GraphQL dashboard data
async function setupAuth(page) {
  await page.addInitScript((token) => {
    const session = {
      accessToken: token,
      refreshToken: token,
      userId: 'user-123',
      tenantId: 'default',
      roles: ['Editor'],
      expiresAt: Date.now() + 86400000,
    };
    sessionStorage.setItem('vedo_session', JSON.stringify(session));
    localStorage.setItem('vedo-jwt-token', token);
  }, ALICE_JWT);

  // Mock GraphQL dashboard data
  await page.route('**/api/v1/graphql', async (route) => {
    const body = route.request().postDataJSON();
    const query = (body?.query || '') + (body?.operationName || '');
    const data = {};

    if (/dashboard/i.test(query)) {
      data.dashboard = {
        widgets: [
          { title: 'Merge Requests', count: 3, icon: 'git-merge', route: '/merge-requests' },
          { title: 'Reviews', count: 2, icon: 'eye', route: '/reviews' },
          { title: 'Work Items', count: 7, icon: 'list-todo', route: '/work-items' },
        ],
        attentionItems: [
          { id: 'att-001', text: 'Merge request MR-001 needs review', severity: 'warning', count: 1 },
        ],
        activityFeed: [
          { id: 'act-001', text: 'Alice created University Ontology', author: 'Alice', timestamp: '2026-03-15T14:30:00Z', type: 'ontology' },
        ],
        recentOntologies: [
          { id: 'ont-123', name: 'University Ontology', description: 'Academic ontology project', visibility: 'private', updatedAt: '2026-03-15T14:30:00Z' },
        ],
      };
    }
    if (/classTree|classes|individuals/i.test(query)) {
      data.classTree = [];
      data.individuals = [];
    }
    if (/GetOntology|ontology/i.test(query)) {
      data.ontology = { id: 'ont-123', name: 'Test Ontology', branch: 'main', commit: 'abc123', dirty: false };
    }
    if (/commits|branches|tags|compare|graph|mergeRequest/i.test(query)) {
      data.commits = { items: [], total: 0 };
      data.branches = { items: [], total: 0 };
      data.tags = [];
    }
    if (/ontologyMetrics|sparql|validation|projects/i.test(query)) {
      data.ontologyMetrics = { counters: { classCount: 0, propertyCount: 0, individualCount: 0 } };
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data }),
    });
  });
}

test.describe('Screen Rendering', () => {
  test('login page — renders 5 OAuth providers', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByRole('heading', { name: 'Sign in to VEDO' })).toBeVisible()

    const providers = ['VK ID', 'Yandex ID', 'Mail.ru', 'Google', 'Corporate SSO']
    for (const provider of providers) {
      await expect(page.getByRole('button', { name: provider })).toBeVisible()
    }
  })

  test('dashboard — renders user greeting and widgets', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/dashboard')
    await expect(page.getByRole('heading', { name: 'Alice' })).toBeVisible()
    await expect(page.getByText('Editor')).toBeVisible()
    await expect(page.getByText('online')).toBeVisible()

    // @ctx: GUI-DASH-001 Invariant 1 — zero-count widgets still render
    await expect(page.getByText('Merge Requests', { exact: true })).toBeVisible()
    await expect(page.getByText('Reviews', { exact: true })).toBeVisible()
    await expect(page.getByText('Work Items', { exact: true })).toBeVisible()
  })

  test('workspace — 3-panel layout', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/ontology/ont-123/workspace')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)
    // Check that all 3 panels are rendered as expected
    await expect(page.locator('.panel-left')).toHaveCount(1)
    await expect(page.locator('.panel-center')).toHaveCount(1)
    await expect(page.locator('.panel-right')).toHaveCount(1)
  })

  test('versioning — all 6 view tabs render', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/ontology/ont-123/versioning/commits')
    const tabs = ['Commits', 'Branches', 'Compare', 'Tags', 'Graph', 'Merge Requests']
    for (const tab of tabs) {
      await expect(page.getByRole('tab', { name: tab })).toBeVisible()
    }
  })

  test('versioning — MR placeholder shows', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/ontology/ont-123/versioning/merge_requests')
    await expect(page.getByText('coming soon')).toBeVisible()
  })

  test('public ontology — renders without auth', async ({ page }) => {
    await page.goto('/public/ont-123')
    await expect(page.getByRole('heading', { name: 'Public Ontology' })).toBeVisible()
  })

  test('not found — 404 page for unknown routes', async ({ page }) => {
    await page.goto('/nonexistent-route')
    await expect(page.getByText('404')).toBeVisible()
    await expect(page.getByText('Page not found')).toBeVisible()
  })
})
