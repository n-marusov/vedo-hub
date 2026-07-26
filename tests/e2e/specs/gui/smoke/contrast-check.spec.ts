// Playwright tests for color contrast per UX-A11Y-001 Invariant 4
// A11Y contrast check — WCAG AA ratio (4.5:1 normal, 3:1 large)
import { test, expect } from '@playwright/test'

test.describe('Color Contrast', () => {
  test('login page — OAuth buttons meet contrast ratio', async ({ page }) => {
    await page.goto('/login')
    await page.addScriptTag({ path: require.resolve('axe-core') })

    const violations = await page.evaluate(async () => {
      const results = await (window as any).axe.run()
      return results.violations.filter((v: any) => v.id === 'color-contrast')
    })

    expect(violations).toHaveLength(0)
  })

  test('dashboard — text meets contrast ratio', async ({ page }) => {
    await page.goto('/dashboard')
    await page.addScriptTag({ path: require.resolve('axe-core') })

    const violations = await page.evaluate(async () => {
      const results = await (window as any).axe.run()
      return results.violations.filter((v: any) => v.id === 'color-contrast')
    })

    expect(violations).toHaveLength(0)
  })
})
