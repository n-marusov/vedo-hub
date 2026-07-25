<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Ticket filters — status, category, date range filtering -->
<!-- @hlv:sec [INPUT_VALIDATION] — filter values validated against known enums -->

<template>
  <div class="ticket-filters" role="search" aria-label="Filter tickets">
    <div class="ticket-filters__row">
      <!-- Status filter -->
      <div class="ticket-filters__group">
        <label for="filter-status" class="ticket-filters__label">Status</label>
        <select
          id="filter-status"
          :value="modelValue.status"
          @change="onStatusChange"
          class="ticket-filters__select"
        >
          <option value="">All</option>
          <option v-for="s in statuses" :key="s" :value="s">{{ statusLabels[s] }}</option>
        </select>
      </div>

      <!-- Category filter -->
      <div class="ticket-filters__group">
        <label for="filter-category" class="ticket-filters__label">Category</label>
        <select
          id="filter-category"
          :value="modelValue.category"
          @change="onCategoryChange"
          class="ticket-filters__select"
        >
          <option value="">All</option>
          <option v-for="c in categories" :key="c" :value="c">{{ catLabels[c] }}</option>
        </select>
      </div>

      <!-- Date range filter -->
      <div class="ticket-filters__group">
        <label for="filter-date-from" class="ticket-filters__label">From</label>
        <input
          id="filter-date-from"
          type="date"
          :value="modelValue.dateFrom"
          @change="onDateFromChange"
          class="ticket-filters__input"
        />
      </div>

      <div class="ticket-filters__group">
        <label for="filter-date-to" class="ticket-filters__label">To</label>
        <input
          id="filter-date-to"
          type="date"
          :value="modelValue.dateTo"
          @change="onDateToChange"
          class="ticket-filters__input"
        />
      </div>

      <!-- Clear filters -->
      <button class="ticket-filters__clear" @click="onClear" aria-label="Clear all filters">
        Clear
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: {
    status: string
    category: string
    dateFrom: string
    dateTo: string
  }
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string>]
}>()

const statuses = ['new', 'in_review', 'in_progress', 'resolved', 'closed', 'reopened'] as const
const categories = ['bug', 'performance', 'question', 'feature', 'documentation', 'access'] as const

const statusLabels: Record<string, string> = {
  new: 'New',
  in_review: 'In Review',
  in_progress: 'In Progress',
  resolved: 'Resolved',
  closed: 'Closed',
  reopened: 'Reopened'
}

const catLabels: Record<string, string> = {
  bug: 'Bug',
  performance: 'Performance',
  question: 'Question',
  feature: 'Feature',
  documentation: 'Documentation',
  access: 'Access'
}

function onStatusChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', { ...props.modelValue, status: target.value })
}

function onCategoryChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', { ...props.modelValue, category: target.value })
}

function onDateFromChange(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', { ...props.modelValue, dateFrom: target.value })
}

function onDateToChange(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', { ...props.modelValue, dateTo: target.value })
}

function onClear() {
  emit('update:modelValue', { status: '', category: '', dateFrom: '', dateTo: '' })
}
</script>

<style scoped>
.ticket-filters { padding: var(--spacing-3); background: var(--surface-secondary); border-radius: var(--radius-md); }
.ticket-filters__row { display: flex; flex-wrap: wrap; gap: var(--spacing-3); align-items: flex-end; }
.ticket-filters__group { display: flex; flex-direction: column; gap: var(--spacing-1); }
.ticket-filters__label { font-size: var(--font-size-xs); color: var(--text-muted); }
.ticket-filters__select, .ticket-filters__input {
  padding: var(--spacing-1) var(--spacing-2); border: 1px solid var(--border-default);
  border-radius: var(--radius-sm); font-size: var(--font-size-sm); background: var(--surface-primary);
}
.ticket-filters__clear {
  padding: var(--spacing-1) var(--spacing-3); border: 1px solid var(--border-default);
  border-radius: var(--radius-sm); font-size: var(--font-size-sm); background: transparent;
  cursor: pointer; color: var(--font-secondary); margin-left: auto;
}
.ticket-filters__clear:hover { background: var(--surface-hover); }
</style>
