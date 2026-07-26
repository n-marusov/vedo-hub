<!-- BranchList.vue -->
<template>
  <div class="branch-list" role="region" :aria-label="'Branch list'">
    <div class="branch-list__filter">
      <Tab v-model="filter" :tabs="filterTabs" label="Branch filter" />
    </div>
    <ul class="branch-list__list">
      <li v-for="branch in filteredBranches" :key="branch.name" class="branch-list__item">
        <span class="branch-list__name">{{ branch.name }}</span>
        <Badge :text="branch.status" :variant="branch.status === 'active' ? 'success' : 'default'" />
        <span class="branch-list__commit">{{ branch.lastCommit?.slice(0, 7) }}</span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Badge from '../ui-kit/Badge.vue'
import Tab from '../ui-kit/Tab.vue'

const props = defineProps<{
  branches: Array<{ name: string; status: string; lastCommit: string }>
}>()

const filter = ref('all')
const filterTabs = [
  { value: 'all', label: 'All' },
  { value: 'active', label: 'Active' }
]

const filteredBranches = computed(() => {
  if (filter.value === 'all') return props.branches
  return props.branches.filter((b) => b.status === 'active')
})
</script>

<style scoped>
.branch-list { padding: var(--spacing-4); }
.branch-list__filter { margin-bottom: var(--spacing-4); }
.branch-list__list { display: flex; flex-direction: column; gap: var(--spacing-2); }
.branch-list__item { display: flex; align-items: center; gap: var(--spacing-2); padding: var(--spacing-2); border-radius: var(--radius-md); }
.branch-list__item:hover { background: var(--surface-secondary); }
.branch-list__name { flex: 1; font-family: var(--font-family-mono); font-size: var(--font-size-sm); }
.branch-list__commit { font-family: var(--font-family-mono); font-size: var(--font-size-xs); color: var(--text-muted); }
</style>
