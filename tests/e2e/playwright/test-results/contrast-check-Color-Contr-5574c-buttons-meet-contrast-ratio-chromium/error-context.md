# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: contrast-check.spec.ts >> Color Contrast >> login page — OAuth buttons meet contrast ratio
- Location: tests\contrast-check.spec.ts:6:7

# Error details

```
Error: Cannot find module 'axe-core'
Require stack:
- D:\Projects\vedo-core\llm\tests\e2e\playwright\tests\contrast-check.spec.ts
- D:\Projects\vedo-core\llm\tests\e2e\playwright\node_modules\playwright\lib\common\index.js
- D:\Projects\vedo-core\llm\tests\e2e\playwright\node_modules\playwright\lib\worker\workerProcessEntry.js
```

# Page snapshot

```yaml
- main [ref=e4]:
  - main "Sign in to VEDO" [ref=e5]:
    - generic [ref=e6]:
      - generic [ref=e7]: VEDO
      - heading "Sign in to VEDO" [level=1] [ref=e8]
      - list "OAuth providers" [ref=e9]:
        - listitem "Sign in with VK ID" [ref=e10] [cursor=pointer]:
          - generic [ref=e11]: VK ID
        - listitem "Sign in with Yandex ID" [ref=e12] [cursor=pointer]:
          - generic [ref=e13]: Yandex ID
        - listitem "Sign in with Mail.ru" [ref=e14] [cursor=pointer]:
          - generic [ref=e15]: Mail.ru
        - listitem "Sign in with Google" [ref=e16] [cursor=pointer]:
          - generic [ref=e17]: Google
        - listitem "Sign in with Corporate SSO" [ref=e18] [cursor=pointer]:
          - generic [ref=e19]: Corporate SSO
```

# Test source

```ts
  1  | // @ctx: Playwright tests for color contrast per UX-A11Y-001 Invariant 4
  2  | // @hlv A11Y contrast check — WCAG AA ratio (4.5:1 normal, 3:1 large)
  3  | import { test, expect } from '@playwright/test'
  4  | 
  5  | test.describe('Color Contrast', () => {
  6  |   test('login page — OAuth buttons meet contrast ratio', async ({ page }) => {
  7  |     await page.goto('/login')
> 8  |     await page.addScriptTag({ path: require.resolve('axe-core') })
     |                                             ^ Error: Cannot find module 'axe-core'
  9  | 
  10 |     const violations = await page.evaluate(async () => {
  11 |       const results = await (window as any).axe.run()
  12 |       return results.violations.filter((v: any) => v.id === 'color-contrast')
  13 |     })
  14 | 
  15 |     expect(violations).toHaveLength(0)
  16 |   })
  17 | 
  18 |   test('dashboard — text meets contrast ratio', async ({ page }) => {
  19 |     await page.goto('/dashboard')
  20 |     await page.addScriptTag({ path: require.resolve('axe-core') })
  21 | 
  22 |     const violations = await page.evaluate(async () => {
  23 |       const results = await (window as any).axe.run()
  24 |       return results.violations.filter((v: any) => v.id === 'color-contrast')
  25 |     })
  26 | 
  27 |     expect(violations).toHaveLength(0)
  28 |   })
  29 | })
  30 | 
```