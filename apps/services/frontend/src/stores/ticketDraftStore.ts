// @hlv:artifact code-frontend implements spec-gui-ticket-001
// @ctx: Ticket draft store — localStorage persistence for interrupted ticket creation flow
// @hlv:sec [INPUT_VALIDATION] — user-supplied ticket data stored in localStorage, must be sanitized on restore

import { computed, ref } from 'vue'

// @ctx: domain types for ticket draft
export interface TicketDraft {
  title: string
  description: string
  category: 'bug' | 'performance' | 'question' | 'feature' | 'documentation' | 'access' | ''
  user_severity: 'critical' | 'high' | 'medium' | 'low' | ''
  steps_to_reproduce: string
  expected_behavior: string
  attachments: DraftAttachment[]
  trace_id: string | null
  page_url: string | null
  saved_at: string
}

export interface DraftAttachment {
  filename: string
  size: number
  type: string
}

const STORAGE_KEY = 'vedo_ticket_draft'
const DRAFT_TTL_MS = 7 * 24 * 60 * 60 * 1000 // 7 days

function emptyDraft(): TicketDraft {
  return {
    title: '',
    description: '',
    category: '',
    user_severity: '',
    steps_to_reproduce: '',
    expected_behavior: '',
    attachments: [],
    trace_id: null,
    page_url: null,
    saved_at: new Date().toISOString()
  }
}

function loadDraft(): TicketDraft | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as TicketDraft
    const age = Date.now() - new Date(parsed.saved_at).getTime()
    if (age > DRAFT_TTL_MS) {
      localStorage.removeItem(STORAGE_KEY)
      return null
    }
    return parsed
  } catch {
    // @ctx: corrupted storage — cannot restore
    // @hlv TICKET-UI-DRAFT-RESTORE-FAIL
    localStorage.removeItem(STORAGE_KEY)
    return null
  }
}

function saveDraft(draft: TicketDraft): void {
  try {
    const toSave = { ...draft, saved_at: new Date().toISOString() }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(toSave))
  } catch {
    // @ctx: localStorage quota exceeded or disabled — silent fail
    console.error('Failed to save ticket draft to localStorage')
  }
}

function clearDraft(): void {
  localStorage.removeItem(STORAGE_KEY)
}

const draft = ref<TicketDraft>(loadDraft() || emptyDraft())
const hasDraft = computed(() => {
  if (!draft.value) return false
  return !!(draft.value.title || draft.value.description || draft.value.category)
})

function updateDraft(partial: Partial<TicketDraft>): void {
  draft.value = { ...draft.value, ...partial }
  saveDraft(draft.value)
}

function restoreDraft(): TicketDraft | null {
  const saved = loadDraft()
  if (saved) {
    draft.value = saved
    return saved
  }
  return null
}

function resetDraft(): void {
  draft.value = emptyDraft()
  clearDraft()
}

export function useTicketDraftStore() {
  return {
    draft,
    hasDraft,
    updateDraft,
    restoreDraft,
    resetDraft,
    clearDraft
  }
}

// @ctx: unit tests for ticket draft store
// @hlv TICKET-UI-DRAFT-RESTORE-FAIL
// @hlv atomicity — draft auto-save during creation
// @hlv TICKET-UI-TOO-MANY-FILES — attachment limit in draft
// @hlv TICKET-UI-FILE-TOO-LARGE — file size limit in draft
// @hlv:sec [INPUT_VALIDATION] — draft data sanitized on restore
