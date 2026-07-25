<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Ticket detail — full ticket view with comments, close/reopen actions -->
<!-- @hlv:sec [AUTH_BOUNDARY] — users can only close/reopen their own tickets -->

<template>
  <div class="ticket-detail" v-if="ticket" aria-labelledby="ticket-detail-title">
    <!-- Header -->
    <div class="ticket-detail__header">
      <h2 id="ticket-detail-title" class="ticket-detail__title">{{ ticket.title }}</h2>
      <div class="ticket-detail__badges">
        <span class="badge" :class="`badge--${ticket.status}`">{{ statusLabel(ticket.status) }}</span>
        <span class="badge badge--priority">{{ ticket.system_priority?.toUpperCase() }}</span>
        <span class="badge badge--category">{{ catLabel(ticket.category) }}</span>
      </div>
    </div>

    <!-- Meta info -->
    <div class="ticket-detail__meta">
      <span class="ticket-detail__id">ID: {{ ticket.id }}</span>
      <time class="ticket-detail__date" :datetime="ticket.created_at">
        Created {{ formatDate(ticket.created_at) }}
      </time>
    </div>

    <!-- Description -->
    <div class="ticket-detail__section">
      <h3 class="ticket-detail__section-title">Description</h3>
      <p class="ticket-detail__description">{{ ticket.description }}</p>
    </div>

    <!-- Conditional fields -->
    <div v-if="ticket.steps_to_reproduce" class="ticket-detail__section">
      <h3 class="ticket-detail__section-title">Steps to Reproduce</h3>
      <pre class="ticket-detail__pre">{{ ticket.steps_to_reproduce }}</pre>
    </div>

    <div v-if="ticket.expected_behavior" class="ticket-detail__section">
      <h3 class="ticket-detail__section-title">Expected Behavior</h3>
      <p class="ticket-detail__description">{{ ticket.expected_behavior }}</p>
    </div>

    <!-- Metadata -->
    <MetadataPreview
      v-if="ticket.metadata"
      :version="ticket.metadata.vedo_version"
      :environment="ticket.metadata.environment"
      :user-agent="ticket.metadata.user_agent ?? undefined"
      :trace-id="ticket.metadata.trace_id ?? undefined"
      :page-url="ticket.metadata.page_url ?? undefined"
    />

    <!-- Comments -->
    <TicketComments
      :comments="ticket.comments || []"
      :ticket-status="ticket.status"
      @comment="onComment"
    />

    <!-- Actions -->
    <!-- @ctx: users can close their own tickets at any time -->
    <!-- @hlv atomicity — close own ticket -->
    <!-- @ctx: users can reopen their own closed tickets -->
    <!-- @hlv atomicity — reopen closed ticket -->
    <div class="ticket-detail__actions">
      <button
        v-if="ticket.status !== 'closed' && ticket.status !== 'resolved'"
        class="btn btn--ghost btn--danger"
        @click="onClose"
      >
        Close Ticket
      </button>
      <button
        v-if="ticket.status === 'closed'"
        class="btn btn--primary"
        @click="onReopen"
      >
        Reopen Ticket
      </button>
    </div>

    <!-- Confirmation dialog for close -->
    <Dialog
      :open="showCloseConfirm"
      title="Close Ticket"
      size="sm"
      :modal="true"
      @close="showCloseConfirm = false"
    >
      <p>Are you sure you want to close this ticket?</p>
      <div class="ticket-detail__dialog-actions">
        <button class="btn btn--ghost" @click="showCloseConfirm = false">Cancel</button>
        <button class="btn btn--primary" @click="confirmClose">Close</button>
      </div>
    </Dialog>

    <!-- Confirmation dialog for reopen -->
    <Dialog
      :open="showReopenConfirm"
      :title="ticket?.system_priority === 'p0' ? 'Reopen P0 Ticket' : 'Reopen Ticket'"
      size="sm"
      :modal="true"
      @close="showReopenConfirm = false"
    >
      <p v-if="ticket?.system_priority === 'p0'">
        This is a P0 ticket. Reopening requires support engineer confirmation.
      </p>
      <p v-else>Are you sure you want to reopen this ticket?</p>
      <div class="ticket-detail__dialog-actions">
        <button class="btn btn--ghost" @click="showReopenConfirm = false">Cancel</button>
        <button class="btn btn--primary" @click="confirmReopen">Reopen</button>
      </div>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Dialog from '../ui-kit/Dialog.vue'
import MetadataPreview from './MetadataPreview.vue'
import TicketComments from './TicketComments.vue'

interface TicketDetailData {
  id: string
  title: string
  description: string
  status: string
  system_priority: string
  category: string
  created_at: string
  steps_to_reproduce?: string | null
  expected_behavior?: string | null
  metadata?: {
    vedo_version?: string
    environment?: string
    user_agent?: string | null
    trace_id?: string | null
    page_url?: string | null
  }
  comments?: Array<{ author: string; text: string; created_at: string }>
}

const props = defineProps<{
  ticket: TicketDetailData | null
}>()

const _props = props
void _props

const emit = defineEmits<{
  close: []
  reopen: []
  comment: [text: string]
}>()

const showCloseConfirm = ref(false)
const showReopenConfirm = ref(false)

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

function statusLabel(status: string): string {
  return statusLabels[status] || status
}

function catLabel(category: string): string {
  return catLabels[category] || category
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function onClose() {
  showCloseConfirm.value = true
}

function confirmClose() {
  showCloseConfirm.value = false
  emit('close')
}

function onReopen() {
  showReopenConfirm.value = true
}

function confirmReopen() {
  showReopenConfirm.value = false
  emit('reopen')
}

function onComment(text: string) {
  emit('comment', text)
}
</script>

<style scoped>
.ticket-detail { display: flex; flex-direction: column; gap: var(--spacing-4); }
.ticket-detail__header { display: flex; flex-direction: column; gap: var(--spacing-2); }
.ticket-detail__title { font-size: var(--font-size-lg); font-weight: var(--font-weight-semibold); margin: 0; }
.ticket-detail__badges { display: flex; gap: var(--spacing-2); }
.badge { font-size: var(--font-size-xs); padding: 2px var(--spacing-2); border-radius: var(--radius-full); font-weight: var(--font-weight-medium); }
.badge--new { background: var(--color-info); color: var(--font-on-info); }
.badge--in_progress { background: var(--color-warning); color: var(--font-on-warning); }
.badge--resolved { background: var(--color-success); color: var(--font-on-success); }
.badge--closed { background: var(--text-muted); color: var(--font-primary); }
.badge--reopened { background: var(--color-warning); color: var(--font-on-warning); }
.badge--priority { background: var(--color-error); color: var(--font-on-error); }
.badge--category { background: var(--surface-tertiary); color: var(--font-secondary); }
.ticket-detail__meta { display: flex; justify-content: space-between; font-size: var(--font-size-xs); color: var(--text-muted); }
.ticket-detail__section { display: flex; flex-direction: column; gap: var(--spacing-1); }
.ticket-detail__section-title { font-size: var(--font-size-sm); font-weight: var(--font-weight-semibold); margin: 0; }
.ticket-detail__description { font-size: var(--font-size-sm); margin: 0; white-space: pre-wrap; }
.ticket-detail__pre { font-size: var(--font-size-sm); background: var(--surface-secondary); padding: var(--spacing-3); border-radius: var(--radius-md); white-space: pre-wrap; font-family: var(--font-mono); }
.ticket-detail__actions { display: flex; gap: var(--spacing-2); justify-content: flex-end; padding-top: var(--spacing-2); }
.ticket-detail__dialog-actions { display: flex; gap: var(--spacing-2); justify-content: flex-end; padding-top: var(--spacing-3); }
.btn { padding: var(--spacing-2) var(--spacing-4); border-radius: var(--radius-md); font-size: var(--font-size-sm); cursor: pointer; border: none; }
.btn--primary { background: var(--color-primary); color: var(--font-on-primary); }
.btn--ghost { background: transparent; color: var(--font-secondary); border: 1px solid var(--border-default); }
.btn--danger { color: var(--color-error); border-color: var(--color-error); }
</style>
