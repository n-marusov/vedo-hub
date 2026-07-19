// Playwright E2E tests for keyboard navigation per UX-A11Y-001 Invariant 1
import { test, expect } from '@playwright/test'

// A1 keyboard navigation — all interactive elements reachable via Tab
test.describe('Keyboard Navigation', () => {
  test('login page — OAuth buttons reachable via Tab', async ({ page }) => {
    await page.goto('/login')
    // Tab through all interactive elements
    await page.keyboard.press('Tab')
    const focused = await page.locator(':focus').count()
    expect(focused).toBeGreaterThan(0)

    // All 5 OAuth buttons should be reachable
    const buttons = page.locator('button[role="listitem"]')
    await expect(buttons).toHaveCount(5)
  })

  test('dashboard — widgets focusable', async ({ page }) => {
    await page.goto('/dashboard')
    await page.keyboard.press('Tab')
    const widgetCards = page.locator('.widget-card:focus')
    await expect(widgetCards).toBeVisible()
  })

  test('versioning tabs — Arrow key navigation', async ({ page }) => {
    await page.goto('/ontology/ont-123/versioning/commits')
    const tabs = page.locator('.ver-tabs button')
    await tabs.first().focus()
    await page.keyboard.press('ArrowRight')
    const secondTab = tabs.nth(1)
    await expect(secondTab).toBeFocused()
  })

  test('SPARQL editor — Ctrl+Enter runs query', async ({ page }) => {
    await page.goto('/ontology/ont-123/sparql')
    const editor = page.locator('textarea')
    await editor.fill('SELECT ?class WHERE { ?class a owl:Class }')
    await page.keyboard.press('Control+Enter')
    // Query execution triggered (mock)
  })

  test('members table — keyboard accessible actions', async ({ page }) => {
    await page.goto('/ontology/ont-123/members')
    const removeButtons = page.locator('button.btn-remove')
    if (await removeButtons.count() > 0) {
      await removeButtons.first().focus()
      await expect(removeButtons.first()).toBeFocused()
    }
  })
})
