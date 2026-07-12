<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit TableColumn, Spacer -->

<!-- TableColumn.vue -->
<template>
  <th :class="['table-col', { 'table-col--sortable': sortable }]" :aria-sort="sortDirection" scope="col">
    <button v-if="sortable && field" class="table-col__sort" @click="$emit('sort', field)">
      {{ label }}
      <span class="table-col__indicator" aria-hidden="true">
        {{ sortDirection === 'ascending' ? '↑' : sortDirection === 'descending' ? '↓' : '↕' }}
      </span>
    </button>
    <span v-else>{{ label }}</span>
  </th>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  field?: string
  sortable?: boolean
  sortDirection?: 'ascending' | 'descending' | 'none'
}>()

defineEmits<{
  sort: [field: string]
}>()
</script>

<style scoped>
.table-col {
  padding: var(--spacing-2) var(--spacing-3);
  text-align: left;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 2px solid var(--border-default);
}

.table-col__sort {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  background: none;
  border: none;
  cursor: pointer;
  color: inherit;
  font: inherit;
}

.table-col__indicator {
  color: var(--text-muted);
  font-size: var(--font-size-xs);
}
</style>
