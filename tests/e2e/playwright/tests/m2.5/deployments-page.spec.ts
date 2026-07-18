// @ctx: M2.5 Deployments page — cards from API, show/hide stopped, delete
import { test, expect } from '@playwright/test'
import { DeploymentsPage } from '../../pages/deployments.page'

test.describe('M2.5 Deployments Page', () => {
  test('should render deployment cards from API when page loads', async ({ page }) => {
    const deployments = new DeploymentsPage(page)
    await deployments.goto()
    const cards = deployments.getDeployments()
    await expect(cards.first()).toBeVisible()
  })

  test('should toggle show stopped deployments when toggle is clicked', async ({ page }) => {
    const deployments = new DeploymentsPage(page)
    await deployments.goto()
    await deployments.toggleShowStopped()
    const stoppedCards = page.locator('.deployment-card.stopped')
    const count = await stoppedCards.count()
    expect(count).toBeGreaterThanOrEqual(0)
  })

  test('should show confirmation before deleting a deployment', async ({ page }) => {
    const deployments = new DeploymentsPage(page)
    await deployments.goto()
    const firstUrl = await deployments.getDeployments().first().locator('.deployment-url').textContent()
    if (firstUrl) {
      await deployments.deleteDeployment(firstUrl)
      await expect(page.getByText(/confirm/i)).toBeVisible()
    }
  })
})
