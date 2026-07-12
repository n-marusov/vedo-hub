// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Navigation state composable — sidebar collapse, active route, persisted per session
// @hlv sidebar_collapse_persist — state saved per user session

import { computed, ref } from 'vue'

const STORAGE_KEY = 'vedo-nav-state'

interface NavState {
  sidebarCollapsed: boolean
  activeRoute: string
  breadcrumbs: Array<{ label: string; route: string }>
}

const state = ref<NavState>({
  sidebarCollapsed: false,
  activeRoute: '/',
  breadcrumbs: []
})

export function useNavigationState() {
  const sidebarCollapsed = computed(() => state.value.sidebarCollapsed)
  const activeRoute = computed(() => state.value.activeRoute)
  const breadcrumbs = computed(() => state.value.breadcrumbs)

  function toggleSidebar() {
    state.value.sidebarCollapsed = !state.value.sidebarCollapsed
    persistState()
  }

  function setSidebarCollapsed(collapsed: boolean) {
    state.value.sidebarCollapsed = collapsed
    persistState()
  }

  function setActiveRoute(route: string, label?: string) {
    state.value.activeRoute = route
    if (label) {
      state.value.breadcrumbs.push({ label, route })
    }
    persistState()
  }

  function setBreadcrumbs(items: Array<{ label: string; route: string }>) {
    state.value.breadcrumbs = items
  }

  function clearBreadcrumbs() {
    state.value.breadcrumbs = []
  }

  function persistState() {
    try {
      sessionStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          sidebarCollapsed: state.value.sidebarCollapsed
        })
      )
    } catch {
      // sessionStorage unavailable — degrade gracefully
    }
  }

  function loadState() {
    try {
      const stored = sessionStorage.getItem(STORAGE_KEY)
      if (stored) {
        const parsed = JSON.parse(stored)
        state.value.sidebarCollapsed = parsed.sidebarCollapsed ?? false
      }
    } catch {
      // Use defaults
    }
  }

  // Initialize on composable creation
  loadState()

  return {
    sidebarCollapsed,
    activeRoute,
    breadcrumbs,
    toggleSidebar,
    setSidebarCollapsed,
    setActiveRoute,
    setBreadcrumbs,
    clearBreadcrumbs
  }
}
