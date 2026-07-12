<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Root shell strictly aligned to design/frontend.pen Header + Sidebar organisms -->
<template>
  <router-view v-if="!showShell" />

  <div v-else class="shell">
    <header class="shell-header" role="banner">
      <div class="header-brand" aria-label="VEDO Core (go to dashboard)">
        <img class="header-brand-logo" src="/vedo-core-logo-1.jpg" alt="VEDO Core" />
        <span class="header-brand-text">VEDO Core</span>
      </div>

      <div class="header-fill-spacer"></div>

      <div class="header-search" role="search" aria-label="Search">
        <Search :size="15" class="header-search-icon" />
        <span class="header-search-text">Search or go to...</span>
        <span class="header-search-fill"></span>
        <span class="header-search-shortcut">/</span>
      </div>

      <div class="header-fill-spacer"></div>

      <div class="header-actions" aria-label="Header actions">
        <button class="header-icon-btn" type="button" aria-label="Create"><Plus :size="14" /></button>
        <span class="header-action-divider" aria-hidden="true"></span>
        <button class="header-combo-btn" type="button" aria-label="Merge requests">
          <GitMerge :size="16" />
          <span class="header-pill-badge">0</span>
        </button>
        <button class="header-combo-btn" type="button" aria-label="Comments">
          <MessageSquare :size="16" />
          <span class="header-pill-badge">0</span>
        </button>
        <button class="header-icon-btn" type="button" aria-label="Help"><CircleHelp :size="16" /></button>
        <button class="header-icon-btn" type="button" :aria-label="themeLabel" @click="toggleTheme">
          <component :is="themeIcon" :size="16" />
        </button>
        <button class="header-avatar-menu" type="button" aria-label="Current user menu">
          <span class="header-avatar"><User :size="16" /></span>
          <ChevronDown :size="12" class="header-avatar-chevron" />
        </button>
      </div>
    </header>

    <div class="shell-main">
      <aside :class="['shell-sidebar', { 'shell-sidebar--collapsed': collapsed }]" aria-label="Main navigation">
        <div class="sidebar-header">
          <span class="sidebar-header-text">Workspace</span>
        </div>

        <nav class="sidebar-nav">
          <button
            v-for="item in mainItems"
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

      <main class="shell-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ChevronDown,
  CircleHelp,
  Cloud,
  Folder,
  GitMerge,
  History,
  Info,
  Layers,
  LayoutDashboard,
  type LucideIcon,
  MessageSquare,
  Moon,
  PanelLeftClose,
  PanelLeftOpen,
  Plus,
  Search,
  Sun,
  User
} from 'lucide-vue-next'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { applyThemeMode } from './theme/manager'

const route = useRoute()
const router = useRouter()

const showShell = computed(() => {
  return !['login', 'not-found', 'public-ontology', 'auth-callback'].includes(String(route.name))
})

type SidebarItem = {
  label: string
  icon: LucideIcon
  to: string
  matches: string[]
  badge?: string
}

const mainItems: SidebarItem[] = [
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

function isActive(matches: string[]): boolean {
  return matches.some((value) => route.path.startsWith(value))
}

function navigate(to: string): void {
  router.push(to)
}

const collapsed = ref(false)
const currentTheme = ref<'light' | 'dark'>('dark')

const themeIcon = computed(() => (currentTheme.value === 'dark' ? Sun : Moon))
const themeLabel = computed(() =>
  currentTheme.value === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'
)

onMounted(() => {
  const saved = localStorage.getItem('sidebar-collapsed')
  if (saved === 'true') collapsed.value = true

  const theme = document.documentElement.getAttribute('data-theme') || 'dark'
  currentTheme.value = theme as 'light' | 'dark'
})

watch(collapsed, (val) => {
  localStorage.setItem('sidebar-collapsed', String(val))
})

function toggleSidebar(): void {
  collapsed.value = !collapsed.value
}

function toggleTheme(): void {
  const next = currentTheme.value === 'dark' ? 'light' : 'dark'
  applyThemeMode(next)
  currentTheme.value = next
}
</script>

<style scoped>
.shell {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--background);
  color: var(--foreground);
}

.shell-header {
  flex-shrink: 0;
  height: 56px;
  background: var(--card);
  border-bottom: 1px solid var(--border);
  padding: 0 24px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-brand {
  height: 36px;
  border-radius: 6px;
  padding: 0 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-brand-logo {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  object-fit: contain;
}

.header-fill-spacer {
  flex: 1;
  min-width: 0;
}

.header-brand-text {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 700;
  color: var(--foreground);
  white-space: nowrap;
}

.header-search {
  width: 520px;
  height: 36px;
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.header-search-icon,
.header-search-text {
  color: var(--muted-foreground);
}

.header-search-text {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.header-search-fill {
  flex: 1;
}

.header-search-shortcut {
  height: 22px;
  min-width: 22px;
  border-radius: 4px;
  border: 1px solid var(--border);
  background: var(--muted);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 6px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 600;
}

.header-actions {
  height: 36px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.header-action-divider {
  width: 1px;
  height: 20px;
  background: var(--border);
  margin: 0 2px;
}

.header-icon-btn,
.header-avatar {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--foreground);
}

.header-icon-btn {
  background: transparent;
  border: none;
  cursor: pointer;
}

.header-icon-btn:hover {
  background: var(--muted);
}

.header-combo-btn {
  height: 32px;
  border-radius: 6px;
  padding: 0 8px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--foreground);
  background: transparent;
  border: none;
  cursor: pointer;
}

.header-combo-btn:hover {
  background: var(--muted);
}

.header-pill-badge {
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  border: 1px solid var(--border);
  background: var(--muted);
  padding: 0 5px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 10px;
  font-weight: 600;
}

.header-avatar {
  border-radius: 16px;
  color: var(--muted-foreground);
  background: var(--muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.header-avatar-menu {
  height: 32px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
}

.header-avatar-chevron {
  color: var(--muted-foreground);
}

.shell-main {
  flex: 1;
  min-height: 0;
  display: flex;
}

.shell-sidebar {
  width: 256px;
  height: 100%;
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

.shell-sidebar--collapsed {
  width: 56px;
  padding: 16px 8px;
}

.shell-sidebar--collapsed .sidebar-item {
  padding: 8px;
  justify-content: center;
}

.shell-sidebar--collapsed .sidebar-item-indicator {
  display: none;
}

.shell-sidebar--collapsed .sidebar-item-label,
.shell-sidebar--collapsed .sidebar-badge,
.shell-sidebar--collapsed .sidebar-header-text {
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
  font-weight: 500;
  text-align: left;
  position: relative;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background var(--transition-fast);
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

.sidebar-item--danger {
  color: var(--destructive);
}

.shell-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  height: 100%;
  overflow-y: auto;
}
</style>
