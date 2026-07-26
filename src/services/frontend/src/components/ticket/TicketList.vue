<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Ticket list — user's own manual tickets with filtering and selection -->
<!-- @hlv:sec [AUTH_BOUNDARY] — row-level security: users only see their own manual tickets -->

<template>
  <div class="ticket-list" aria-labelledby="ticket-list-title">
    <h2 id="ticket-list-title" class="ticket-list__title">My Tickets</h2>

    <!-- Filters -->
    <TicketFilters v-model="filters" />

    <!-- Loading state -->
    <div v-if="loading" class="ticket-list__loading" role="status" aria-live="polite">
      Loading tickets...
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredTickets.length === 0" class="ticket-list__empty">
      <p>No tickets found.</p>
      <button class="btn btn--sm btn--primary" @click="$emit('create')">Create a ticket</button>
    </div>

    <!-- Ticket table -->
    <table v-else class="ticket-list__table" role="grid" aria-label="Ticket list">
      <thead>
        <tr>
          <th scope="col" class="ticket-list__th">ID</th>
          <th scope="col" class="ticket-list__th">Title</th>
          <th scope="col" class="ticket-list__th">Status</th>
          <th scope="col" class="ticket-list__th">Priority</th>
          <th scope="col" class="ticket-list__th">Category</th>
          <th scope="col" class="ticket-list__th">Created</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="ticket in filteredTickets"
          :key="ticket.id"
          class="ticket-list__row"
          :class="{ 'ticket-list__row--selected': selectedId === ticket.id }"
          @click="onSelect(ticket)"
          tabindex="0"
          role="row"
          :aria-selected="selectedId === ticket.id"
          @keydown.enter="onSelect(ticket)"
        >
          <td class="ticket-list__cell ticket-list__cell--id">{{ shortId(ticket.id) }}</td>
          <td class="ticket-list__cell">{{ ticket.title }}</td>
          <td class="ticket-list__cell">
            <span class="badge" :class="`badge--${ticket.status}`">{{ statusLabel(ticket.status) }}</span>
          </td>
          <td class="ticket-list__cell">
            <span class="badge badge--priority">{{ ticket.system_priority?.toUpperCase() }}</span>
          </td>
          <td class="ticket-list__cell">{{ catLabel(ticket.category) }}</td>
          <td class="ticket-list__cell ticket-list__cell--date">{{ formatDate(ticket.created_at) }}</td>
        </tr>
      </tbody>
    </table>

    <!-- Pagination info -->
    <div v-if="(total || 0) > (pageSize || 20)" class="ticket-list__pagination">
      <span>Showing {{ filteredTickets.length }} of {{ total }} tickets</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import TicketFilters from './TicketFilters.vue'

interface TicketListItem {
  id: string
  title: string
  status: string
  system_priority: string
  category: string
  created_at: string
}

const props = defineProps<{
  tickets: TicketListItem[]
  total?: number
  loading?: boolean
  pageSize?: number
}>()

const emit = defineEmits<{
  select: [ticket: TicketListItem]
  create: []
}>()

const selectedId = ref<string | null>(null)
const filters = ref({
  status: '',
  category: '',
  dateFrom: '',
  dateTo: ''
})

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

// @ctx: client-side filtering by status, category, date range
// @hlv atomicity — ticket list filtering consistency
const filteredTickets = computed(() => {
  let result = [...props.tickets]

  if (filters.value.status) {
    result = result.filter((t) => t.status === filters.value.status)
  }

  if (filters.value.category) {
    result = result.filter((t) => t.category === filters.value.category)
  }

  if (filters.value.dateFrom) {
    const from = new Date(filters.value.dateFrom).getTime()
    result = result.filter((t) => new Date(t.created_at).getTime() >= from)
  }

  if (filters.value.dateTo) {
    const to = new Date(filters.value.dateTo).getTime() + 86400000 // end of day
    result = result.filter((t) => new Date(t.created_at).getTime() <= to)
  }

  return result
})

function statusLabel(status: string): string {
  return statusLabels[status] || status
}

function catLabel(category: string): string {
  return catLabels[category] || category
}

function shortId(uuid: string): string {
  return uuid.slice(0, 8)
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString()
  } catch {
    return iso
  }
}

function onSelect(ticket: TicketListItem) {
  selectedId.value = ticket.id
  emit('select', ticket)
}
</script>

<style scoped>
.ticket-list { display: flex; flex-direction: column; gap: var(--spacing-4); }
.ticket-list__title { font-size: var(--font-size-lg); font-weight: var(--font-weight-semibold); margin: 0; }
.ticket-list__loading { padding: var(--spacing-6); text-align: center; color: var(--text-muted); }
.ticket-list__empty { padding: var(--spacing-6); text-align: center; color: var(--text-muted); display: flex; flex-direction: column; align-items: center; gap: var(--spacing-3); }
.ticket-list__table { width: 100%; border-collapse: collapse; font-size: var(--font-size-sm); }
.ticket-list__th { text-align: left; padding: var(--spacing-2) var(--spacing-3); border-bottom: 2px solid var(--border-default); font-weight: var(--font-weight-semibold); color: var(--text-muted); font-size: var(--font-size-xs); }
.ticket-list__row { cursor: pointer; transition: background 0.15s; }
.ticket-list__row:hover { background: var(--surface-hover); }
.ticket-list__row--selected { background: var(--surface-active); }
.ticket-list__cell { padding: var(--spacing-2) var(--spacing-3); border-bottom: 1px solid var(--border-default); }
.ticket-list__cell--id { font-family: var(--font-mono); font-size: var(--font-size-xs); }
.ticket-list__cell--date { color: var(--text-muted); font-size: var(--font-size-xs); }
.ticket-list__pagination { font-size: var(--font-size-xs); color: var(--text-muted); text-align: center; }
.badge { font-size: var(--font-size-xs); padding: 2px var(--spacing-2); border-radius: var(--radius-full); font-weight: var(--font-weight-medium); }
.badge--new { background: var(--color-info); color: var(--font-on-info); }
.badge--in_progress { background: var(--color-warning); color: var(--font-on-warning); }
.badge--resolved { background: var(--color-success); color: var(--font-on-success); }
.badge--closed { background: var(--text-muted); color: var(--font-primary); }
.badge--reopened { background: var(--color-warning); color: var(--font-on-warning); }
.badge--priority { background: var(--color-error); color: var(--font-on-error); }
.btn { padding: var(--spacing-2) var(--spacing-4); border-radius: var(--radius-md); font-size: var(--font-size-sm); cursor: pointer; border: none; }
.btn--primary { background: var(--color-primary); color: var(--font-on-primary); }
.btn--sm { padding: var(--spacing-1) var(--spacing-2); font-size: var(--font-size-xs); }
</style>
