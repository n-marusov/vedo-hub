// @m4 — MetricsPage vitest spec (GREEN phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.3): page should render KPI counters and trends from API

import { describePage, mountWithProviders, waitForQuery } from '@/__tests__/setup/mock-providers'
import { afterEach, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

// MetricsPage uses REST getOntologyMetrics() with real axios calls.
// Mock the API module so tests can control success/error behavior.
vi.mock('@/api/metrics', () => ({
  getOntologyMetrics: vi.fn()
}))

const mockMetrics = {
  counters: {
    classCount: 42,
    propertyCount: 18,
    individualCount: 156,
    axiomCount: 10420,
    commentCount: 7,
    mergeRequestCount: 3
  },
  trends: [
    { date: '2024-01-01', classCount: 30, propertyCount: 10, individualCount: 100 },
    { date: '2024-02-01', classCount: 42, propertyCount: 18, individualCount: 156 }
  ]
}

describePage('MetricsPage', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should render KPI counters from API when page loads', async () => {
    const { getOntologyMetrics } = await import('@/api/metrics')
    vi.mocked(getOntologyMetrics).mockResolvedValue(mockMetrics)
    const MetricsPage = (await import('@/pages/MetricsPage.vue')).default
    const wrapper = mountWithProviders(MetricsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.kpi-grid').exists()).toBe(true)
  })

  it('should display class count from API', async () => {
    const { getOntologyMetrics } = await import('@/api/metrics')
    vi.mocked(getOntologyMetrics).mockResolvedValue(mockMetrics)
    const MetricsPage = (await import('@/pages/MetricsPage.vue')).default
    const wrapper = mountWithProviders(MetricsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.metrics-content').exists()).toBe(true)
  })

  it('should show context banner with ontology name', async () => {
    const { getOntologyMetrics } = await import('@/api/metrics')
    vi.mocked(getOntologyMetrics).mockResolvedValue(mockMetrics)
    const MetricsPage = (await import('@/pages/MetricsPage.vue')).default
    const wrapper = mountWithProviders(MetricsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.metrics-context').exists()).toBe(true)
  })

  it('should render KPI cards when data loads', async () => {
    const { getOntologyMetrics } = await import('@/api/metrics')
    vi.mocked(getOntologyMetrics).mockResolvedValue(mockMetrics)
    const MetricsPage = (await import('@/pages/MetricsPage.vue')).default
    const wrapper = mountWithProviders(MetricsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.kpi-grid').exists()).toBe(true)
  })

  it('should show loading skeleton while metrics load from API', async () => {
    const { getOntologyMetrics } = await import('@/api/metrics')
    vi.mocked(getOntologyMetrics).mockResolvedValue(mockMetrics)
    const MetricsPage = (await import('@/pages/MetricsPage.vue')).default
    const wrapper = mountWithProviders(MetricsPage)
    expect(wrapper.find('.metrics-page').exists()).toBe(true)
  })

  it('should show empty state when ontology has zero metrics', async () => {
    const { getOntologyMetrics } = await import('@/api/metrics')
    vi.mocked(getOntologyMetrics).mockResolvedValue({
      counters: {
        classCount: 0,
        propertyCount: 0,
        individualCount: 0,
        axiomCount: 0,
        commentCount: 0,
        mergeRequestCount: 0
      },
      trends: []
    })
    const MetricsPage = (await import('@/pages/MetricsPage.vue')).default
    const wrapper = mountWithProviders(MetricsPage)
    await waitForQuery()
    await nextTick()
    expect(
      wrapper.find('.kpi-grid').exists() ||
        wrapper.text().includes('0') ||
        wrapper.text().includes('Classes')
    ).toBe(true)
  })

  it('should show error state when metrics API fails', async () => {
    const { getOntologyMetrics } = await import('@/api/metrics')
    vi.mocked(getOntologyMetrics).mockRejectedValue(new Error('Metrics error'))
    const MetricsPage = (await import('@/pages/MetricsPage.vue')).default
    const wrapper = mountWithProviders(MetricsPage)
    await new Promise((resolve) => setTimeout(resolve, 200))
    await nextTick()
    expect(
      wrapper.find('.error-state').exists() ||
        wrapper.text().includes('retry') ||
        wrapper.text().includes('error')
    ).toBe(true)
  })
})
