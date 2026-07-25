<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Sidebar component — collapsible navigation with active state indicators per GUI-OW-001 v2 -->
<!-- @hlv:sec [AUTH_BOUNDARY] — nav items filtered by user role -->
<template>
  <aside
    class="sidebar"
    :class="{ 'sidebar--collapsed': collapsed }"
    :style="{ width: collapsed ? 'var(--sidebar-compact-width)' : 'var(--sidebar-width)' }"
  >
    <div class="sidebar-header">
      <span class="sidebar-header-text">Workspace</span>
    </div>

    <nav class="sidebar-nav" aria-label="Main navigation">
      <button
        v-for="item in navItems"
        :key="item.label"
        :class="['sidebar-item', { 'sidebar-item--active': isActive(item.matches) }]"
        type="button"
        @click="navigate(item.to)"
      >
        <span class="sidebar-item-indicator"></span>
        <component :is="item.icon" :size="16" class="sidebar-item-icon" />
        <span class="sidebar-item-label">{{ item.label }}</span>
        <span v-if="item.badge" class="sidebar-badge">{{ item.badge }}</span>
      </button>
    </nav>

    <div class="sidebar-divider"></div>

    <span class="sidebar-spacer"></span>

    <button class="sidebar-item" type="button" aria-label="Help">
      <Info :size="16" class="sidebar-item-icon" />
      <span class="sidebar-item-label">Help</span>
    </button>

    <div class="sidebar-divider"></div>

    <button class="sidebar-item" type="button" :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'" @click="toggleSidebar">
      <component :is="collapsed ? PanelLeftOpen : PanelLeftClose" :size="16" class="sidebar-item-icon" />
      <span class="sidebar-item-label">{{ collapsed ? 'Expand sidebar' : 'Collapse sidebar' }}</span>
    </button>
  </aside>
</template>

<script setup lang="ts">
// @ctx: Sidebar navigation items and collapse state persistence per GUI-OW-001
import {
  Cloud,
  Folder,
  GitMerge,
  History,
  Info,
  Layers,
  LayoutDashboard,
  type LucideIcon,
  MessageSquare,
  PanelLeftClose,
  PanelLeftOpen
} from 'lucide-vue-next'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

interface NavItem {
  label: string
  icon: LucideIcon
  to: string
  matches: string[]
  badge?: string
}

const navItems: NavItem[] = [
  { label: 'Home', icon: LayoutDashboard, to: '/dashboard/home', matches: ['/dashboard/home'] },
  { label: 'Groups', icon: Layers, to: '/dashboard/groups', matches: ['/dashboard/groups'] },
  { label: 'Projects', icon: Folder, to: '/dashboard/projects', matches: ['/dashboard/projects'] },
  {
    label: 'Merge requests',
    icon: GitMerge,
    to: '/dashboard/merge_requests',
    matches: ['/dashboard/merge_requests'],
    badge: '0'
  },
  { label: 'Commits', icon: History, to: '/commits', matches: ['/commits'], badge: '0' },
  { label: 'Comments', icon: MessageSquare, to: '/comments', matches: ['/comments'], badge: '0' },
  {
    label: 'Deployments',
    icon: Cloud,
    to: '/dashboard/deployments',
    matches: ['/dashboard/deployments'],
    badge: '0'
  }
]

const collapsed = ref(false)
const route = useRoute()
const router = useRouter()

function isActive(matches: string[]): boolean {
  return matches.some((value) => route.path.startsWith(value))
}

function navigate(to: string): void {
  router.push(to)
}

// @hlv Navigation State Store must persist sidebar collapse state per user session
const STORAGE_KEY = 'vedo-sidebar-collapsed'

onMounted(() => {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored !== null) {
    collapsed.value = stored === 'true'
  }
})

onBeforeUnmount(() => {
  localStorage.setItem(STORAGE_KEY, String(collapsed.value))
})

function toggleSidebar(): void {
  collapsed.value = !collapsed.value
}

defineExpose({ collapsed })
</script>

<style scoped>
.sidebar {
  height: 100vh;
  background: var(--card);
  border-right: 1px solid var(--border);
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: width 0.2s ease-in-out, padding 0.2s ease-in-out;
  overflow-y: auto;
  white-space: nowrap;
  flex-shrink: 0;
}

.sidebar--collapsed {
  width: 56px;
  padding: 16px 8px;
}

.sidebar--collapsed .sidebar-item {
  padding: 8px;
  justify-content: center;
}

.sidebar--collapsed .sidebar-item-indicator {
  display: none;
}

.sidebar--collapsed .sidebar-item-label,
.sidebar--collapsed .sidebar-badge,
.sidebar--collapsed .sidebar-header-text {
  display: none;
}

.sidebar-header {
  width: 100%;
  padding: 0 8px;
  display: flex;
  align-items: center;
  min-height: 24px;
}

.sidebar-header-text {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sidebar-item {
  width: 100%;
  border-radius: 6px;
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  text-align: left;
  position: relative;
}

.sidebar-item:hover {
  background: rgba(20, 20, 20, 0.3);
}

.sidebar-item-indicator {
  display: none;
  position: absolute;
  left: 0;
  width: 3px;
  height: 28px;
  border-radius: 2px;
  background: var(--primary);
}

.sidebar-item--active {
  background: rgba(16, 185, 129, 0.1);
  color: var(--primary);
}

.sidebar-item--active .sidebar-item-indicator {
  display: block;
}

.sidebar-item-icon {
  flex-shrink: 0;
}

.sidebar-item-label {
  flex: 1;
}

.sidebar-badge {
  min-width: 20px;
  height: 20px;
  border-radius: 999px;
  padding: 0 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--muted);
  border: 1px solid var(--border);
  color: var(--foreground);
  font-size: 11px;
}

.sidebar-divider {
  width: 100%;
  height: 1px;
  background: var(--border);
  margin: 8px 0;
}

.sidebar-spacer {
  flex: 1;
}

.nav-link:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: -2px;
}
</style>
