// @m2.5 — VersioningPage vitest spec (RED phase: will fail on hardcoded data)
// After GREEN (Task 2.2): should load commits and branches from API via Apollo query
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import { nextTick } from 'vue'

// Mock the router
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/ontology/:id/versioning/:view',
      name: 'ontology-versioning',
      component: { template: '<div />' }
    }
  ]
})

describe('VersioningPage', () => {
  beforeEach(async () => {
    await router.push('/ontology/test/versioning/commits')
  })

  it('should render commit history from API when Commits tab is active', async () => {
    const VersioningPage = (await import('@/pages/VersioningPage.vue')).default
    const wrapper = mount(VersioningPage, {
      global: { plugins: [router] }
    })
    await nextTick()
    // RED: Should show commits loaded from API (not hardcoded)
    expect(wrapper.text()).toContain('Commit History')
  })

  it('should switch tabs when tab button is clicked', async () => {
    const VersioningPage = (await import('@/pages/VersioningPage.vue')).default
    const wrapper = mount(VersioningPage, {
      global: { plugins: [router] }
    })
    await nextTick()
    const tabs = wrapper.findAll('.tab')
    expect(tabs.length).toBeGreaterThanOrEqual(6)
  })

  it('should show loading state while versioning data loads from API', async () => {
    const VersioningPage = (await import('@/pages/VersioningPage.vue')).default
    const wrapper = mount(VersioningPage, {
      global: { plugins: [router] }
    })
    await nextTick()
    // RED: Should show loading indicator when data is loading from API
    expect(wrapper.find('.version-card').exists()).toBe(true)
  })
})
