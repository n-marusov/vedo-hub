// @ctx: M2.5 Validation page — run button → spinner → results; SHACL OK stub; timestamp update
import { test, expect } from '@playwright/test'
import { ValidationPage } from '../../pages/validation.page'

test.describe('M2.5 Validation Page', () => {
  test('should show validation summary when page loads', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    const summary = validation.getSummary()
    await expect(summary).toBeVisible()
  })

  test('should run validation and show spinner then results when run button is clicked', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    await validation.runValidation()
    await expect(page.locator('.loading-indicator, .spinner')).toBeVisible()
    const results = validation.getResults()
    await expect(results).toBeVisible({ timeout: 5000 })
  })

  test('should show SHACL OK status when validation passes', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    await validation.runValidation()
    await expect(page.getByText(/valid|ok|passed|no violations/i)).toBeVisible({ timeout: 5000 })
  })

  test('should update validation timestamp after successful run', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    await validation.runValidation()
    await expect(page.locator('.validation-timestamp')).toBeVisible({ timeout: 5000 })
  })
})
