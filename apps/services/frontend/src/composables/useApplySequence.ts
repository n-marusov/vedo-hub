// Composable for managing the apply-sequence workflow
//
// Handles API calls, progress tracking, and workflow state transitions
// for applying extracted ontology sequences.

import { computed, ref } from 'vue'
import { applySequence } from '../api/extraction'
import type { ApplyProgress, ApplyResult, SequenceStep } from '../types/extraction'

export function useApplySequence() {
  // ── State ────────────────────────────────────────────────────────────────

  const progress = ref<ApplyProgress>({
    totalSteps: 0,
    completedSteps: 0,
    currentStep: '',
    status: 'idle'
  })

  const applyResult = ref<ApplyResult | null>(null)
  const commitMessage = ref('')
  const abortController = ref<AbortController | null>(null)

  // ── Computed ─────────────────────────────────────────────────────────────

  const isApplying = computed(() => progress.value.status === 'applying')
  const isConfirming = computed(() => progress.value.status === 'confirming')
  const isSuccess = computed(() => progress.value.status === 'success')
  const isError = computed(() => progress.value.status === 'error')
  const isIdle = computed(() => progress.value.status === 'idle')
  const percentage = computed(() => {
    if (progress.value.totalSteps === 0) return 0
    return Math.round((progress.value.completedSteps / progress.value.totalSteps) * 100)
  })

  // ── Workflow actions ─────────────────────────────────────────────────────

  /// Show confirmation dialog
  function requestConfirm(steps: SequenceStep[]) {
    const included = steps.filter((s) => s.included)
    progress.value = {
      totalSteps: included.length,
      completedSteps: 0,
      currentStep: '',
      status: 'confirming'
    }
    commitMessage.value = `Import ontology from document extraction (${included.length} steps)`

    console.debug('[useApplySequence] confirm requested', {
      stepCount: included.length
    })
  }

  /// Execute the apply sequence API call
  async function executeApply(ontologyId: string, steps: SequenceStep[], message: string) {
    const included = steps.filter((s) => s.included)

    console.info('[useApplySequence] apply started', {
      ontologyId,
      stepCount: included.length
    })

    progress.value = {
      totalSteps: included.length,
      completedSteps: 0,
      currentStep: 'Preparing...',
      status: 'applying'
    }

    abortController.value = new AbortController()

    try {
      const result = await applySequence({
        ontologyId,
        steps: included,
        commitMessage: message,
        signal: abortController.value.signal
      })

      if (result.success) {
        console.info('[useApplySequence] apply success', {
          appliedCount: result.appliedCount,
          commitId: result.commitId
        })

        progress.value = {
          totalSteps: included.length,
          completedSteps: result.appliedCount,
          currentStep: '',
          status: 'success'
        }
        applyResult.value = result
      } else {
        console.warn('[useApplySequence] apply partial failure', {
          appliedCount: result.appliedCount,
          errors: result.errors?.length
        })

        progress.value = {
          totalSteps: included.length,
          completedSteps: result.appliedCount,
          currentStep: '',
          status: 'error',
          errorMessage: result.errors?.[0]?.message || 'Unknown error during apply'
        }
        applyResult.value = result
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error'

      console.error('[useApplySequence] apply failed', { error: message })

      progress.value = {
        ...progress.value,
        status: 'error',
        errorMessage: message
      }
    } finally {
      abortController.value = null
    }
  }

  /// Cancel the apply operation
  function cancelApply() {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
    }

    console.debug('[useApplySequence] apply cancelled')

    progress.value = {
      totalSteps: 0,
      completedSteps: 0,
      currentStep: '',
      status: 'idle'
    }
  }

  /// Reset to idle state
  function reset() {
    abortController.value = null
    applyResult.value = null
    commitMessage.value = ''
    progress.value = {
      totalSteps: 0,
      completedSteps: 0,
      currentStep: '',
      status: 'idle'
    }
  }

  return {
    // State
    progress,
    applyResult,
    commitMessage,
    isApplying,
    isConfirming,
    isSuccess,
    isError,
    isIdle,
    percentage,

    // Actions
    requestConfirm,
    executeApply,
    cancelApply,
    reset
  }
}
