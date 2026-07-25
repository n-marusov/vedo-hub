<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism Sidebar component — navigation menu with route links, role-based visibility -->
<template>
  <nav
    :class="['sidebar', { 'sidebar--collapsed': collapsed }]"
    :aria-label="t('nav.sidebar')"
  >
    <ul class="sidebar__nav" role="list">
      <li v-for="item in visibleItems" :key="item.route" class="sidebar__item">
        <router-link
          :to="item.route"
          :class="['sidebar__link', { 'sidebar__link--active': isActive(item.route) }]"
          :aria-current="isActive(item.route) ? 'page' : undefined"
        >
          <span class="sidebar__icon" aria-hidden="true">{{ item.icon }}</span>
          <span v-if="!collapsed" class="sidebar__label">{{ t(item.labelKey) }}</span>
        </router-link>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from '../../composables/useI18n'

const { t } = useI18n()
const route = useRoute()

defineProps<{
  collapsed?: boolean
  userRole?: string
}>()

interface NavItem {
  route: string
  icon: string
  labelKey: string
  minRole?: string
}

const allItems: NavItem[] = [
  { route: '/dashboard/home', icon: '🏠', labelKey: 'nav.home' },
  { route: '/dashboard/groups', icon: '🗂️', labelKey: 'nav.groups' },
  { route: '/dashboard/projects', icon: '📁', labelKey: 'nav.projects' },
  { route: '/dashboard/merge_requests', icon: '🔗', labelKey: 'nav.merge_requests' },
  { route: '/commits', icon: '🕐', labelKey: 'nav.commits' },
  { route: '/comments', icon: '💬', labelKey: 'nav.comments' },
  { route: '/dashboard/deployments', icon: '☁️', labelKey: 'nav.deployments' }
]

const visibleItems = computed(() => {
  return allItems
})

function isActive(routePath: string): boolean {
  return route.path === routePath || route.path.startsWith(`${routePath}/`)
}
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  height: calc(100vh - var(--header-height));
  background-color: var(--surface-primary);
  border-right: 1px solid var(--border-default);
  transition: width var(--transition-normal);
  overflow: hidden;
}

.sidebar--collapsed {
  width: var(--sidebar-compact-width);
}

.sidebar__nav {
  padding: var(--spacing-4) 0;
}

.sidebar__item {
  margin: var(--spacing-1) var(--spacing-2);
}

.sidebar__link {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-decoration: none;
  transition: all var(--transition-fast);
}

.sidebar__link:hover {
  background-color: var(--surface-secondary);
  color: var(--text-primary);
}

.sidebar__link--active {
  background-color: var(--primary-muted);
  color: var(--primary);
  font-weight: var(--font-weight-medium);
}

.sidebar__icon {
  font-size: var(--font-size-lg);
  flex-shrink: 0;
}

.sidebar__label {
  font-size: var(--font-size-sm);
  white-space: nowrap;
}
</style>
