// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Draft state composable — preserves unsaved changes across navigation via Apollo cache
// @hlv atomicity — draft changes preserved across route navigation

import { computed, ref } from 'vue'
import { apolloClient } from '../apollo/client'
import { UPDATE_DRAFT_MUTATION } from '../apollo/queries'

interface DraftChange {
  field: string
  oldValue: unknown
  newValue: unknown
  timestamp: number
}

interface DraftState {
  ontologyId: string | null
  changes: Map<string, DraftChange>
  hasUnsavedChanges: boolean
}

const state = ref<DraftState>({
  ontologyId: null,
  changes: new Map(),
  hasUnsavedChanges: false
})

export function useDraftState() {
  const hasUnsavedChanges = computed(() => state.value.hasUnsavedChanges)
  const changeCount = computed(() => state.value.changes.size)

  function setOntologyContext(ontologyId: string) {
    state.value.ontologyId = ontologyId
  }

  function trackChange(field: string, oldValue: unknown, newValue: unknown) {
    if (!state.value.ontologyId) {
      throw new Error('GUI_STORE_NOT_INITIALIZED: Draft state accessed before ontology context')
    }

    state.value.changes.set(field, {
      field,
      oldValue,
      newValue,
      timestamp: Date.now()
    })
    state.value.hasUnsavedChanges = true
  }

  function getChange(field: string): DraftChange | undefined {
    return state.value.changes.get(field)
  }

  function getAllChanges(): DraftChange[] {
    return Array.from(state.value.changes.values())
  }

  async function saveDraft(): Promise<boolean> {
    if (!state.value.ontologyId || state.value.changes.size === 0) {
      return false
    }

    try {
      const changes = getAllChanges()
      await apolloClient.mutate({
        mutation: UPDATE_DRAFT_MUTATION,
        variables: {
          ontologyId: state.value.ontologyId,
          changes: { fields: changes }
        }
      })

      state.value.changes.clear()
      state.value.hasUnsavedChanges = false
      return true
    } catch (error) {
      console.error('Failed to save draft:', error)
      return false
    }
  }

  function discardChanges() {
    state.value.changes.clear()
    state.value.hasUnsavedChanges = false
  }

  function reset() {
    state.value.ontologyId = null
    state.value.changes.clear()
    state.value.hasUnsavedChanges = false
  }

  return {
    hasUnsavedChanges,
    changeCount,
    setOntologyContext,
    trackChange,
    getChange,
    getAllChanges,
    saveDraft,
    discardChanges,
    reset
  }
}
