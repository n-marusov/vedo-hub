<template>
  <div class="groups-page" role="main" aria-label="Groups page">
    <section class="gp-top">
      <div class="gp-title-col">
        <div class="gp-breadcrumbs">
          <span class="gp-breadcrumb-text">Workspace</span>
          <ChevronRight :size="12" class="gp-breadcrumb-sep" />
          <span class="gp-breadcrumb-text">Groups</span>
        </div>
        <h1 class="gp-title">Groups</h1>
      </div>
      <button class="gp-new-btn" type="button"><Plus :size="14" />New group</button>
    </section>

    <section class="gp-toolbar">
      <div class="gp-search-wrap">
        <Search :size="14" class="gp-search-icon" />
        <input class="gp-search-input" type="text" placeholder="Search groups" aria-label="Search groups" />
      </div>
      <div class="gp-sort-wrap">
        <span class="gp-sort-label">Name</span>
        <ChevronDown :size="12" class="gp-sort-chevron" />
        <span class="gp-sort-divider"></span>
        <span class="gp-sort-label">Ascending</span>
        <ChevronDown :size="12" class="gp-sort-chevron" />
      </div>
    </section>

    <section class="gp-list">
      <div v-for="(row, i) in groupRows" :key="row.name + i" class="gp-row">
        <div class="gp-row-body">
          <div class="gp-row-body-top">
            <div class="gp-row-indent" :style="{ width: row.indent + 'px' }"></div>
            <component v-if="row.type === 'group'" :is="row.chevronIcon" :size="12" class="gp-row-chevron" />
            <FolderTree v-if="row.type === 'group'" :size="20" class="gp-row-folder-icon" />
            <Folder v-else :size="20" class="gp-row-folder-icon" />
            <span class="gp-row-logo" :style="{ background: row.logoBg }">{{ row.logoLetter }}</span>
            <span :class="['gp-row-name', { 'gp-row-name-active': row.active, 'gp-row-name--project': row.type === 'project' }]">{{ row.name }}</span>
            <Globe v-if="row.visibility === 'public'" :size="12" class="gp-row-vis-icon" />
            <Lock v-if="row.visibility === 'private'" :size="12" class="gp-row-vis-icon" />
          </div>
          <span
            class="gp-row-desc"
            :style="{ paddingLeft: (row.indent + (row.type === 'group' ? 86 : 68)) + 'px' }"
          >{{ row.description }}</span>
        </div>
        <div class="gp-row-actions">
          <div class="gp-row-counters">
            <div v-if="row.type === 'group'" class="gp-counter">
              <FolderTree :size="14" /><span>{{ row.subgroups }}</span>
            </div>
            <div v-if="row.type === 'group'" class="gp-counter">
              <Folder :size="14" /><span>{{ row.projects }}</span>
            </div>
            <div v-if="row.type === 'group'" class="gp-counter">
              <Users :size="14" /><span>{{ row.members }}</span>
            </div>
            <div v-if="row.type === 'project'" class="gp-counter">
              <Star :size="14" /><span>{{ row.stars }}</span>
            </div>
          </div>
          <span class="gp-row-created">{{ row.created }}</span>
        </div>
        <div class="gp-row-menu-wrap">
          <MoreVertical :size="16" class="gp-row-menu" />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import {
  ChevronDown,
  ChevronRight,
  Folder,
  FolderTree,
  Globe,
  Lock,
  MoreVertical,
  Plus,
  Search,
  Star,
  Users
} from 'lucide-vue-next'
import type { Component } from 'vue'

type RowType = 'group' | 'project'

interface GroupRow {
  name: string
  indent: number
  chevronIcon: Component
  logoLetter: string
  logoBg: string
  visibility: 'public' | 'private'
  description: string
  type: RowType
  subgroups?: number
  projects?: number
  members?: number
  stars?: number
  created: string
  active: boolean
}

const groupRows: GroupRow[] = [
  {
    name: 'Platform',
    indent: 0,
    chevronIcon: ChevronDown,
    logoLetter: 'P',
    logoBg: '#10b98126',
    visibility: 'public',
    description: 'Infrastructure and core platform services',
    type: 'group',
    subgroups: 3,
    projects: 8,
    members: 12,
    created: 'Created 2 weeks ago',
    active: false
  },
  {
    name: 'Core Services',
    indent: 18,
    chevronIcon: ChevronRight,
    logoLetter: 'C',
    logoBg: '#6366f126',
    visibility: 'private',
    description: 'Shared backend services',
    type: 'group',
    subgroups: 2,
    projects: 5,
    members: 7,
    created: 'Created 1 month ago',
    active: false
  },
  {
    name: 'Data Models',
    indent: 18,
    chevronIcon: ChevronDown,
    logoLetter: 'D',
    logoBg: '#10b98126',
    visibility: 'public',
    description: 'Ontology and data model definitions',
    type: 'group',
    subgroups: 4,
    projects: 10,
    members: 15,
    created: 'Created 3 weeks ago',
    active: true
  },
  {
    name: 'ProductOntology',
    indent: 54,
    chevronIcon: 'div' as unknown as Component,
    logoLetter: 'P',
    logoBg: '#d9770626',
    visibility: 'public',
    description: 'Core product ontology',
    type: 'project',
    stars: 3,
    created: 'Created 1 week ago',
    active: false
  },
  {
    name: 'OrganizationOntology',
    indent: 54,
    chevronIcon: 'div' as unknown as Component,
    logoLetter: 'O',
    logoBg: '#05966926',
    visibility: 'private',
    description: 'Org structure model',
    type: 'project',
    stars: 2,
    created: 'Created 5 days ago',
    active: false
  },
  {
    name: 'CustomerOntology',
    indent: 54,
    chevronIcon: 'div' as unknown as Component,
    logoLetter: 'C',
    logoBg: '#0891b226',
    visibility: 'public',
    description: 'Customer domain model',
    type: 'project',
    stars: 4,
    created: 'Created 2 days ago',
    active: false
  },
  {
    name: 'API Integrations',
    indent: 18,
    chevronIcon: ChevronRight,
    logoLetter: 'A',
    logoBg: '#dc262626',
    visibility: 'private',
    description: 'External API gateway configs',
    type: 'project',
    stars: 5,
    created: 'Created 2 months ago',
    active: false
  }
]
</script>

<style scoped>
.groups-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.gp-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.gp-title-col {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.gp-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.gp-breadcrumb-text {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.gp-breadcrumb-sep {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 24px;
  font-weight: 600;
}

.gp-new-btn {
  height: 36px;
  border-radius: 6px;
  padding: 0 14px;
  background: var(--primary);
  color: var(--primary-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.gp-new-btn:hover {
  background: var(--primary-hover);
}

.gp-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.gp-search-wrap {
  flex: 1;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--card);
  padding: 0 10px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.gp-search-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-search-input {
  border: none;
  outline: none;
  background: transparent;
  width: 100%;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--foreground);
  opacity: 0.8;
}

.gp-search-input::placeholder {
  color: var(--muted-foreground);
}

.gp-sort-wrap {
  width: 372px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--card);
  padding: 0 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.gp-sort-label {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.gp-sort-chevron {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-sort-divider {
  width: 1px;
  height: 18px;
  background: var(--border);
  margin: 0 4px;
  flex-shrink: 0;
}

.gp-list {
  display: flex;
  flex-direction: column;
}

.gp-row {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  flex-direction: row;
  align-items: stretch;
  gap: 12px;
}

.gp-row:last-child {
  border-bottom: none;
}

.gp-row-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  justify-content: center;
}

.gp-row-body-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.gp-row-indent {
  flex-shrink: 0;
}

.gp-row-chevron {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-row-folder-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-row-logo {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 16px;
  font-weight: 600;
  color: var(--primary-foreground);
  flex-shrink: 0;
}

.gp-row-name {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  color: var(--foreground);
}

.gp-row-name--project {
  font-weight: 500;
}

.gp-row-name-active {
  color: var(--primary);
}

.gp-row-vis-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-row-desc {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.gp-row-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  justify-content: center;
  flex-shrink: 0;
}

.gp-row-counters {
  display: flex;
  flex-direction: row;
  gap: 4px;
}

.gp-row-created {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
  text-align: right;
}

.gp-counter {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: 70px;
  color: #6b7280;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 500;
  flex-shrink: 0;
}

.gp-counter svg {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.gp-row-menu-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 32px;
}

.gp-row-menu {
  color: #6b7280;
  flex-shrink: 0;
}


</style>
