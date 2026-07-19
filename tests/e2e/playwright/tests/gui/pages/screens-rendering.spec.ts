// Playwright tests for screen rendering with mock data per GUI contracts
import { test, expect } from '@playwright/test'

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
    await page.goto('/dashboard')
    await expect(page.getByRole('heading', { name: 'Alice' })).toBeVisible()
    await expect(page.getByText('Editor')).toBeVisible()
    await expect(page.getByText('online')).toBeVisible()

    // @ctx: GUI-DASH-001 Invariant 1 — zero-count widgets still render
    await expect(page.getByText('Merge Requests')).toBeVisible()
    await expect(page.getByText('Reviews')).toBeVisible()
    await expect(page.getByText('Work Items')).toBeVisible()
  })

  test('workspace — 3-panel layout', async ({ page }) => {
    await page.goto('/ontology/ont-123/workspace')
    await expect(page.locator('.panel-left')).toBeVisible()
    await expect(page.locator('.panel-center')).toBeVisible()
    await expect(page.locator('.panel-right')).toBeVisible()
  })

  test('versioning — all 6 view tabs render', async ({ page }) => {
    await page.goto('/ontology/ont-123/versioning/commits')
    const tabs = ['Commits', 'Branches', 'Compare', 'Tags', 'Graph', 'Merge Requests']
    for (const tab of tabs) {
      await expect(page.getByRole('tab', { name: tab })).toBeVisible()
    }
  })

  test('versioning — MR placeholder shows', async ({ page }) => {
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
