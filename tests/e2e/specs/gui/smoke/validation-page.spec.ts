// Validates: REQ-USR.UI.gui-implementation
// Validates: REQ-FUN.PROCESS.e2e-testing
// Validation page — run button → spinner → results; SHACL OK stub; timestamp update
import { test, expect } from '../../graphql-fixtures'
import { ValidationPage } from '../../../pages/validation.page'

test.describe('Validation Page', () => {
  test('should show validation summary when page loads', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    await validation.runValidation()
    const summary = validation.getSummary()
    await expect(summary).toBeVisible()
  })

  test('should run validation and show spinner then results when run button is clicked', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    await validation.runValidation()
    // Loader icon renders with the .spinning class (per M2.5 task 6.1f)
    await expect(page.locator('.spinning')).toBeVisible()
    const results = validation.getResults()
    await expect(results).toBeVisible({ timeout: 5000 })
  })

  test('should show SHACL OK status when validation passes', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    await validation.runValidation()
    await expect(page.getByText(/all rules passed|^ok$/i)).toBeVisible({ timeout: 5000 })
  })

  test('should update validation timestamp after successful run', async ({ page }) => {
    const validation = new ValidationPage(page)
    await validation.goto('ont-123')
    await validation.runValidation()
    await expect(page.locator('.validation-timestamp')).toBeVisible({ timeout: 5000 })
  })
})
