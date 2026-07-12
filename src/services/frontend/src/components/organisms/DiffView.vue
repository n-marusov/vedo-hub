<!-- DiffView.vue -->
<template>
  <div class="diff-view" role="region" :aria-label="'Revision comparison'">
    <div class="diff-view__selectors">
      <Select v-model="base" :options="commitOptions" label="Base" />
      <span class="diff-view__arrow">→</span>
      <Select v-model="target" :options="commitOptions" label="Target" />
    </div>
    <div class="diff-view__content">
      <div v-for="change in changes" :key="change.path" class="diff-view__change">
        <span :class="['diff-view__indicator', `diff-view__indicator--${change.type}`]">
          {{ change.type === 'added' ? '+' : change.type === 'removed' ? '−' : '~' }}
        </span>
        <code class="diff-view__path">{{ change.path }}</code>
        <pre class="diff-view__diff">{{ change.diff }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Select from '../ui-kit/Select.vue'

defineProps<{
  changes: Array<{ path: string; type: 'added' | 'removed' | 'modified'; diff: string }>
  commitOptions: Array<{ value: string; label: string }>
}>()

const base = ref('')
const target = ref('')
</script>

<style scoped>
.diff-view { padding: var(--spacing-4); }
.diff-view__selectors { display: flex; align-items: center; gap: var(--spacing-2); margin-bottom: var(--spacing-4); }
.diff-view__arrow { font-size: var(--font-size-lg); color: var(--text-muted); }
.diff-view__change { display: flex; align-items: flex-start; gap: var(--spacing-2); padding: var(--spacing-2) 0; border-bottom: 1px solid var(--border-default); }
.diff-view__indicator { font-family: var(--font-family-mono); font-weight: var(--font-weight-bold); }
.diff-view__indicator--added { color: var(--status-success); }
.diff-view__indicator--removed { color: var(--status-error); }
.diff-view__indicator--modified { color: var(--status-warning); }
.diff-view__path { font-family: var(--font-family-mono); font-size: var(--font-size-xs); }
.diff-view__diff { flex: 1; font-family: var(--font-family-mono); font-size: var(--font-size-xs); background: var(--surface-secondary); padding: var(--spacing-2); border-radius: var(--radius-sm); overflow-x: auto; }
</style>
