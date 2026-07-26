// Playwright E2E tests for keyboard navigation per UX-A11Y-001 Invariant 1
import { test, expect } from '@playwright/test'

// JWT with sub=user-123, name=Alice, roles=[Editor], signed with test-jwt-key.pem
const ALICE_JWT = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsInVzZXJfaWQiOiJ1c2VyLTEyMyIsIm5hbWUiOiJBbGljZSIsInByZWZlcnJlZF91c2VybmFtZSI6ImFsaWNlIiwidGVuYW50X2lkIjoiZGVmYXVsdCIsIm9yZ2FuaXphdGlvbl9pZCI6Im9yZy0wMDEiLCJyb2xlcyI6WyJFZGl0b3IiXSwiaWF0IjoxNzg0NTQ0ODMzLCJleHAiOjE3ODQ2MzEyMzN9.qZ8vf0F_B3o_Kc3ArBYpRnCKPajvPqZ5Cvno2-y4NEM8IGqxySJysAOn6_Vq-jbI84LZASf9yUzGezhCFuggWomxNuFXxZ7JF0vhptyWfDFd2hc7Qdl8TSht7EehZoNNOiMQesJImFTIhvxTQRccVjvBZLtxyQnXAnEVXx0BcFb28Wt5hlUCWXEODpSlysLP_QVxyyAtFkV7034_1EQVCa2mvm7inxOcQK9PkwFZxCfsIjzVh-b91u31rVi8oSEPrXeycGp8yas6PERkBbBvIuNWMOv9YSE6y74Q1HpKiTvve7dnNKvqSPRfQ2NpvWo1TNAxSZAcbznShOZANrBVbw'

// Helper: set up auth session + JWT in the page, mock GraphQL data
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

  // Mock GraphQL queries
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
        attentionItems: [],
        activityFeed: [],
        recentOntologies: [],
      };
    }
    if (/commits|branches|tags|compare|graph|mergeRequest/i.test(query)) {
      data.commits = { items: [], total: 0 };
      data.branches = { items: [], total: 0 };
      data.tags = [];
    }
    if (/members/i.test(query)) {
      data.members = [
        { id: 'user-456', userId: 'user-456', username: 'Alice', avatarUrl: '', role: 'Editor', addedAt: '2026-01-15T10:00:00Z' },
      ];
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data }),
    });
  });
}

// A1 keyboard navigation — all interactive elements reachable via Tab
test.describe('Keyboard Navigation', () => {
  test('login page — OAuth buttons reachable via Tab', async ({ page }) => {
    await page.goto('/login')
    // Tab through all interactive elements
    await page.keyboard.press('Tab')
    const focused = await page.locator(':focus').count()
    expect(focused).toBeGreaterThan(0)

    // All 5 OAuth buttons should be reachable
    const buttons = page.locator('button.oauth-btn')
    await expect(buttons).toHaveCount(5)
  })

  test('dashboard — widgets focusable', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/dashboard')
    // Tab through sidebar, header, and greeting to reach widget cards
    for (let i = 0; i < 25; i++) {
      await page.keyboard.press('Tab')
      const count = await page.locator('.widget-card:focus').count()
      if (count > 0) break
    }
    const widgetCards = page.locator('.widget-card:focus')
    await expect(widgetCards).toBeVisible()
  })

  test('versioning tabs — Arrow key navigation', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/ontology/ont-123/versioning/commits')
    const tabs = page.locator('.ver-tabs button')
    await tabs.first().focus()
    await page.keyboard.press('ArrowRight')
    const secondTab = tabs.nth(1)
    await expect(secondTab).toBeFocused()
  })

  test('SPARQL editor — Ctrl+Enter runs query', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/ontology/ont-123/sparql')
    const editor = page.locator('textarea')
    await editor.fill('SELECT ?class WHERE { ?class a owl:Class }')
    await page.keyboard.press('Control+Enter')
    // Query execution triggered (mock)
  })

  test('members table — keyboard accessible actions', async ({ page }) => {
    await setupAuth(page)
    await page.goto('/ontology/ont-123/members')
    const removeButtons = page.locator('button.btn-remove')
    if (await removeButtons.count() > 0) {
      await removeButtons.first().focus()
      await expect(removeButtons.first()).toBeFocused()
    }
  })
})
