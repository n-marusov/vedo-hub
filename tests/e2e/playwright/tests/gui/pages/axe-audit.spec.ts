// Playwright axe-core automated accessibility audit per UX-A11Y-001
// A11Y screen audit — axe automated checks
import { test, expect } from '@playwright/test'

// A11Y_SCREEN_NOT_FOUND
test.describe('Automated Accessibility Audit (axe-core)', () => {
  const screens = [
    { path: '/login', name: 'login' },
    { path: '/dashboard', name: 'dashboard' },
    { path: '/ontology/ont-123/workspace', name: 'workspace' },
    { path: '/ontology/ont-123/sparql', name: 'sparql' },
    // Note: /metrics is served by nginx as plain text, not HTML
    { path: '/ontology/ont-123/members', name: 'members' },
    { path: '/ontology/ont-123/validation', name: 'validation' },
    { path: '/ontology/ont-123/versioning/commits', name: 'versioning' },
    { path: '/public/ont-123', name: 'public' }
  ]

  for (const screen of screens) {
    test(`${screen.name} — no critical axe violations`, async ({ page }) => {
      await page.goto(screen.path)

      // Inject axe-core
      await page.addScriptTag({ path: require.resolve('axe-core') })

      // Run axe audit
      const violations = await page.evaluate(async () => {
        const results = await (window as any).axe.run()
        return results.violations
      })

      // No critical or serious violations
      const criticalViolations = violations.filter(
        (v: any) => v.impact === 'critical' || v.impact === 'serious'
      )
      expect(criticalViolations).toHaveLength(0)
    })
  }
})
