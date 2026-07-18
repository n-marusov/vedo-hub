// @m2.5 — Mock providers for vitest component tests
// Re-exports from test-utils for plan-specified import path

import { describe, it, expect, vi } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import type { ComponentPublicInstance } from 'vue'

// @m2.5 — Creates a mock router with a provided route
export function createMockRouter(initialRoute = '/dashboard/home') {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/:pathMatch(.*)*', name: 'catch-all', component: { template: '<div />' } }
    ]
  })
  router.push(initialRoute)
  return router
}

// @m2.5 — Mounts a component with common providers (router, stubs)
export function mountWithProviders(
  component: ComponentPublicInstance,
  options: Record<string, unknown> = {}
): VueWrapper {
  const router = createMockRouter()
  return mount(component, {
    global: {
      plugins: [router],
      stubs: {
        'router-link': true,
        'router-view': true
      }
    },
    ...options
  })
}

// @m2.5 — Wait for async query to settle (flush promises and timers)
export async function waitForQuery(): Promise<void> {
  await new Promise(resolve => setTimeout(resolve, 50))
}

// @m2.5 — Creates a describePage helper for consistent page test structure
export function describePage(
  name: string,
  fn: () => void
): void {
  describe(`Page: ${name}`, fn)
}
