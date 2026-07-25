<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit ExpandableSection, TableColumn, Spacer -->

<!-- ExpandableSection.vue -->
<template>
  <details :class="['expandable', { 'expandable--open': open }]" :open="open">
    <summary class="expandable__header" @click="$emit('toggle')">
      <span class="expandable__icon" aria-hidden="true">{{ open ? '▼' : '▶' }}</span>
      <span class="expandable__title">{{ title }}</span>
      <span v-if="badge" class="expandable__badge">{{ badge }}</span>
    </summary>
    <div class="expandable__content"><slot /></div>
  </details>
</template>

<script setup lang="ts">
defineProps<{ title: string; open?: boolean; badge?: string }>()
defineEmits<{ toggle: [] }>()
</script>

<style scoped>
.expandable { border: 1px solid var(--border-default); border-radius: var(--radius-md); }
.expandable__header {
  display: flex; align-items: center; gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-4); cursor: pointer;
  font-weight: var(--font-weight-medium); color: var(--text-primary);
  list-style: none;
}
.expandable__header::-webkit-details-marker { display: none; }
.expandable__icon { font-size: var(--font-size-xs); color: var(--text-muted); }
.expandable__badge {
  margin-left: auto; padding: var(--spacing-1) var(--spacing-2);
  font-size: var(--font-size-xs); background: var(--surface-secondary);
  border-radius: var(--radius-full); color: var(--text-secondary);
}
.expandable__content { padding: var(--spacing-4); border-top: 1px solid var(--border-default); }
</style>
