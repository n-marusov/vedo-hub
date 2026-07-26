// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Version context composable — exposes branch, commit, dirty state via REST.
// After GraphQL tightening: ontology metadata migrated from VERSION_CONTEXT_QUERY
// (GraphQL) to GET /api/v1/ontologies/{id} (REST).

import axios from 'axios'
import { computed, ref } from 'vue'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'X-Requested-With': 'XMLHttpRequest' }
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('vedo-jwt-token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

interface VersionContext {
  branch: string | null
  commit: string | null
  dirty: boolean
  ontologyId: string | null
}

const context = ref<VersionContext>({
  branch: null,
  commit: null,
  dirty: false,
  ontologyId: null
})

export function useVersionContext() {
  const branch = computed(() => context.value.branch)
  const commit = computed(() => context.value.commit)
  const dirty = computed(() => context.value.dirty)
  const isInitialized = computed(() => context.value.ontologyId !== null)

  // Loading/error for REST fetch
  const loading = ref(false)
  const error = ref<string | null>(null)

  function setContext(ontologyId: string, branch: string, commit: string, dirty: boolean) {
    context.value = { ontologyId, branch, commit, dirty }
  }

  /** Fetch ontology metadata via REST (replaces Apollo VERSION_CONTEXT_QUERY). */
  async function loadFromRest(ontologyId: string) {
    context.value.ontologyId = ontologyId
    loading.value = true
    error.value = null

    try {
      const { data } = await api.get(`/ontologies/${ontologyId}`)
      context.value.branch = data.branch ?? 'main'
      context.value.commit = data.commit ?? ''
      context.value.dirty = data.dirty ?? false

      console.info(
        JSON.stringify({
          level: 'info',
          msg: 'versioncontext.load.success',
          ontologyId,
          branch: context.value.branch,
          commit: context.value.commit,
          dirty: context.value.dirty,
          ts: new Date().toISOString()
        })
      )
    } catch (err: any) {
      const msg =
        err.response?.data?.error?.message ?? err.message ?? 'Failed to load ontology metadata'
      error.value = msg
      console.error(
        JSON.stringify({
          level: 'error',
          msg: 'versioncontext.load.failed',
          ontologyId,
          error: msg,
          ts: new Date().toISOString()
        })
      )
      // Fall back to defaults so the workspace still loads.
      context.value.branch = 'main'
      context.value.commit = ''
      context.value.dirty = false
    } finally {
      loading.value = false
    }

    return { loading, error }
  }

  function markDirty() {
    context.value.dirty = true
  }

  function markClean() {
    context.value.dirty = false
  }

  function reset() {
    context.value = { branch: null, commit: null, dirty: false, ontologyId: null }
    loading.value = false
    error.value = null
  }

  return {
    branch,
    commit,
    dirty,
    isInitialized,
    setContext,
    loadFromRest,
    markDirty,
    markClean,
    reset
  }
}
