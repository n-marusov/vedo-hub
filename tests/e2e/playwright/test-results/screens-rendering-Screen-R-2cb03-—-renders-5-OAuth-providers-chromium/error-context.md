# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: screens-rendering.spec.ts >> Screen Rendering >> login page — renders 5 OAuth providers
- Location: tests\screens-rendering.spec.ts:6:7

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByRole('button', { name: 'VK ID' })
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByRole('button', { name: 'VK ID' })

```

```yaml
- main:
  - main "Sign in to VEDO":
    - text: VEDO
    - heading "Sign in to VEDO" [level=1]
    - list "OAuth providers":
      - listitem "Sign in with VK ID": VK ID
      - listitem "Sign in with Yandex ID": Yandex ID
      - listitem "Sign in with Mail.ru": Mail.ru
      - listitem "Sign in with Google": Google
      - listitem "Sign in with Corporate SSO": Corporate SSO
```

# Test source

```ts
  1  | // @ctx: Playwright tests for screen rendering with mock data per GUI contracts
  2  | // @hlv:artifact tests-screens verifies GUI-DASH-001, GUI-LOGIN-001, GUI-VER-001
  3  | import { test, expect } from '@playwright/test'
  4  | 
  5  | test.describe('Screen Rendering', () => {
  6  |   test('login page — renders 5 OAuth providers', async ({ page }) => {
  7  |     await page.goto('/login')
  8  |     await expect(page.getByRole('heading', { name: 'Sign in to VEDO' })).toBeVisible()
  9  | 
  10 |     const providers = ['VK ID', 'Yandex ID', 'Mail.ru', 'Google', 'Corporate SSO']
  11 |     for (const provider of providers) {
> 12 |       await expect(page.getByRole('button', { name: provider })).toBeVisible()
     |                                                                  ^ Error: expect(locator).toBeVisible() failed
  13 |     }
  14 |   })
  15 | 
  16 |   test('dashboard — renders user greeting and widgets', async ({ page }) => {
  17 |     await page.goto('/dashboard')
  18 |     await expect(page.getByRole('heading', { name: 'Alice' })).toBeVisible()
  19 |     await expect(page.getByText('Editor')).toBeVisible()
  20 |     await expect(page.getByText('online')).toBeVisible()
  21 | 
  22 |     // @ctx: GUI-DASH-001 Invariant 1 — zero-count widgets still render
  23 |     await expect(page.getByText('Merge Requests')).toBeVisible()
  24 |     await expect(page.getByText('Reviews')).toBeVisible()
  25 |     await expect(page.getByText('Work Items')).toBeVisible()
  26 |   })
  27 | 
  28 |   test('workspace — 3-panel layout', async ({ page }) => {
  29 |     await page.goto('/ontology/ont-123/workspace')
  30 |     await expect(page.locator('.panel-left')).toBeVisible()
  31 |     await expect(page.locator('.panel-center')).toBeVisible()
  32 |     await expect(page.locator('.panel-right')).toBeVisible()
  33 |   })
  34 | 
  35 |   test('versioning — all 6 view tabs render', async ({ page }) => {
  36 |     await page.goto('/ontology/ont-123/versioning/commits')
  37 |     const tabs = ['Commits', 'Branches', 'Compare', 'Tags', 'Graph', 'Merge Requests']
  38 |     for (const tab of tabs) {
  39 |       await expect(page.getByRole('tab', { name: tab })).toBeVisible()
  40 |     }
  41 |   })
  42 | 
  43 |   test('versioning — MR placeholder shows', async ({ page }) => {
  44 |     await page.goto('/ontology/ont-123/versioning/merge_requests')
  45 |     await expect(page.getByText('coming soon')).toBeVisible()
  46 |   })
  47 | 
  48 |   test('public ontology — renders without auth', async ({ page }) => {
  49 |     await page.goto('/public/ont-123')
  50 |     await expect(page.getByRole('heading', { name: 'Public Ontology' })).toBeVisible()
  51 |   })
  52 | 
  53 |   test('not found — 404 page for unknown routes', async ({ page }) => {
  54 |     await page.goto('/nonexistent-route')
  55 |     await expect(page.getByText('404')).toBeVisible()
  56 |     await expect(page.getByText('Page not found')).toBeVisible()
  57 |   })
  58 | })
  59 | 
```