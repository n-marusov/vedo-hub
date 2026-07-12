// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Error presentation layer — catches unhandled promise rejections, displays user-friendly errors
// @hlv catch_unhandled_rejections — all unhandled promise rejections caught and presented

import { computed, ref } from 'vue'

interface AppError {
  id: string
  code: string
  message: string
  severity: 'error' | 'warning' | 'info'
  timestamp: number
  // @hlv:sec [SECRET_HANDLING] — internal stack traces never exposed to UI
  internalDetails?: string // masked from user display
}

const errors = ref<AppError[]>([])
const isVisible = ref(false)

export function useErrorPresentation() {
  const activeErrors = computed(() => errors.value)
  const hasErrors = computed(() => errors.value.length > 0)
  const criticalErrors = computed(() => errors.value.filter((e) => e.severity === 'error'))

  function addError(code: string, message: string, internalDetails?: string) {
    const error: AppError = {
      id: `err-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      code,
      message,
      severity: 'error',
      timestamp: Date.now(),
      internalDetails // @hlv:sec — never rendered in UI
    }
    errors.value.push(error)
    isVisible.value = true

    // Structured logging per observability constraint
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'error.presentation.added',
        code,
        errorId: error.id,
        ts: new Date().toISOString()
      })
    )
  }

  function addWarning(message: string) {
    errors.value.push({
      id: `warn-${Date.now()}`,
      code: 'WARNING',
      message,
      severity: 'warning',
      timestamp: Date.now()
    })
  }

  function dismissError(id: string) {
    errors.value = errors.value.filter((e) => e.id !== id)
    if (errors.value.length === 0) {
      isVisible.value = false
    }
  }

  function clearAll() {
    errors.value = []
    isVisible.value = false
  }

  function hide() {
    isVisible.value = false
  }

  function show() {
    isVisible.value = true
  }

  // @hlv catch_unhandled_rejections — global handler for unhandled promise rejections
  function installGlobalHandler() {
    window.addEventListener('unhandledrejection', (event) => {
      const reason = event.reason
      const message = reason?.message || reason?.toString() || 'Unknown error'
      const code = reason?.code || 'UNHANDLED_REJECTION'

      addError(code, message, reason?.stack)

      // @hlv:sec — prevent default error exposure in console
      event.preventDefault()
    })
  }

  return {
    activeErrors,
    hasErrors,
    criticalErrors,
    isVisible,
    addError,
    addWarning,
    dismissError,
    clearAll,
    hide,
    show,
    installGlobalHandler
  }
}
