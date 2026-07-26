<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Header component — breadcrumb navigation per GUI-OW-001 v2 -->
<template>
  <header class="header">
    <div class="header-left">
      <button class="sidebar-toggle" @click="$emit('toggle-sidebar')" aria-label="Toggle sidebar">
        <span class="toggle-icon">☰</span>
      </button>
      <nav class="breadcrumbs" aria-label="Breadcrumb navigation">
        <ol class="breadcrumb-list">
          <li v-for="(crumb, index) in breadcrumbs" :key="crumb.route" class="breadcrumb-item">
            <router-link
              v-if="index < breadcrumbs.length - 1"
              :to="crumb.route"
              class="breadcrumb-link"
            >
              {{ crumb.label }}
            </router-link>
            <span v-else class="breadcrumb-current" aria-current="page">
              {{ crumb.label }}
            </span>
            <span v-if="index < breadcrumbs.length - 1" class="breadcrumb-separator">/</span>
          </li>
        </ol>
      </nav>
    </div>
    <div class="header-right">
      <span class="header-version" :class="{ 'header-version--dirty': dirty }">
        {{ branch }}{{ dirty ? '*' : '' }}
      </span>
    </div>
  </header>
</template>

<script setup lang="ts">
// @ctx: Header breadcrumb navigation and version context per GUI-OW-001
import { ref } from 'vue'

interface Breadcrumb {
  label: string
  route: string
}

const breadcrumbs = ref<Breadcrumb[]>([{ label: 'Dashboard', route: '/dashboard' }])

const branch = ref('main')
const dirty = ref(false)

defineEmits<{
  'toggle-sidebar': []
}>()
</script>

<style scoped>
.header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-4);
  background: var(--surface);
  border-bottom: 1px solid var(--border-default);
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: var(--text-secondary);
  transition: background var(--transition-fast);
}

.sidebar-toggle:hover {
  background: var(--surface-variant);
}

.toggle-icon {
  font-size: var(--font-size-lg);
}

.breadcrumb-list {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  list-style: none;
  padding: 0;
  margin: 0;
  min-width: 0;
}

.breadcrumb-item {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  white-space: nowrap;
}

.breadcrumb-link {
  color: var(--text-secondary);
  text-decoration: none;
  font-size: var(--font-size-sm);
  transition: color var(--transition-fast);
}

.breadcrumb-link:hover {
  color: var(--primary);
}

.breadcrumb-current {
  color: var(--text-primary);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-sm);
}

.breadcrumb-separator {
  color: var(--text-muted);
  font-size: var(--font-size-sm);
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-shrink: 0;
}

.header-version {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  font-family: var(--font-family-mono);
}

.header-version--dirty {
  color: var(--warning);
}
</style>
