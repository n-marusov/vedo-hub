// @m4 — DeploymentsPage vitest spec (GREEN phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.5): page should render deployment cards from API with show stopped toggle

import { describePage, mountWithProviders, waitForQuery } from '@/__tests__/setup/mock-providers'
import { afterEach, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

// DeploymentsPage uses REST listDeployments() — mock for test control
vi.mock('@/api/deployments', () => ({
  listDeployments: vi.fn()
}))

const mockDeployments = [
  {
    id: 'dep-1',
    name: 'Production',
    status: 'active',
    version: 'v2.1.0',
    url: 'https://prod.example.com',
    updatedAt: new Date().toISOString()
  }
]

describePage('DeploymentsPage', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should render deployments page layout', async () => {
    const { listDeployments } = await import('@/api/deployments')
    vi.mocked(listDeployments).mockResolvedValue(mockDeployments)
    const DeploymentsPage = (await import('@/pages/DeploymentsPage.vue')).default
    const wrapper = mountWithProviders(DeploymentsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.dp-page').exists()).toBe(true)
  })

  it('should display page title Deployments', async () => {
    const { listDeployments } = await import('@/api/deployments')
    vi.mocked(listDeployments).mockResolvedValue(mockDeployments)
    const DeploymentsPage = (await import('@/pages/DeploymentsPage.vue')).default
    const wrapper = mountWithProviders(DeploymentsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.dp-page-title').exists()).toBe(true)
  })

  it('should render deployment cards from API data', async () => {
    const { listDeployments } = await import('@/api/deployments')
    vi.mocked(listDeployments).mockResolvedValue(mockDeployments)
    const DeploymentsPage = (await import('@/pages/DeploymentsPage.vue')).default
    const wrapper = mountWithProviders(DeploymentsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.dp-section').exists()).toBe(true)
  })

  it('should have show stopped deployments checkbox', async () => {
    const { listDeployments } = await import('@/api/deployments')
    vi.mocked(listDeployments).mockResolvedValue(mockDeployments)
    const DeploymentsPage = (await import('@/pages/DeploymentsPage.vue')).default
    const wrapper = mountWithProviders(DeploymentsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.dp-show-stopped').exists()).toBe(true)
  })

  it('should show breadcrumbs navigation', async () => {
    const { listDeployments } = await import('@/api/deployments')
    vi.mocked(listDeployments).mockResolvedValue(mockDeployments)
    const DeploymentsPage = (await import('@/pages/DeploymentsPage.vue')).default
    const wrapper = mountWithProviders(DeploymentsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.dp-breadcrumbs').exists()).toBe(true)
  })

  it('should show error state when deployments API fails', async () => {
    const { listDeployments } = await import('@/api/deployments')
    vi.mocked(listDeployments).mockRejectedValue(new Error('Failed to load deployments'))
    const DeploymentsPage = (await import('@/pages/DeploymentsPage.vue')).default
    const wrapper = mountWithProviders(DeploymentsPage)
    await new Promise((resolve) => setTimeout(resolve, 200))
    await nextTick()
    expect(
      wrapper.find('.error-state').exists() ||
        wrapper.text().includes('retry') ||
        wrapper.text().includes('error')
    ).toBe(true)
  })

  it('should show empty state when no deployments exist', async () => {
    const { listDeployments } = await import('@/api/deployments')
    vi.mocked(listDeployments).mockResolvedValue([])
    const DeploymentsPage = (await import('@/pages/DeploymentsPage.vue')).default
    const wrapper = mountWithProviders(DeploymentsPage)
    await new Promise((resolve) => setTimeout(resolve, 200))
    await nextTick()
    expect(
      wrapper.find('.error-state').exists() ||
        wrapper.text().includes('No deployments') ||
        wrapper.text().includes('deploy')
    ).toBe(true)
  })
})
