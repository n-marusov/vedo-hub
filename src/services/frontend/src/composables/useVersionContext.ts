// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Version context composable — exposes branch, commit, dirty state via Apollo cache
// @hlv version_context_exposed — branch, commit, dirty state available to all components

import { useQuery } from '@vue/apollo-composable'
import { computed, ref } from 'vue'
import { VERSION_CONTEXT_QUERY } from '../apollo/queries'

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

  function setContext(ontologyId: string, branch: string, commit: string, dirty: boolean) {
    context.value = { ontologyId, branch, commit, dirty }
  }

  function loadFromApollo(ontologyId: string) {
    context.value.ontologyId = ontologyId
    // Uses Apollo Client cache — data fetched via VERSION_CONTEXT_QUERY
    const { result, loading, error } = useQuery(VERSION_CONTEXT_QUERY, { id: ontologyId })

    if (result.value?.ontology) {
      context.value.branch = result.value.ontology.branch
      context.value.commit = result.value.ontology.commit
      context.value.dirty = result.value.ontology.dirty
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
  }

  return {
    branch,
    commit,
    dirty,
    isInitialized,
    setContext,
    loadFromApollo,
    markDirty,
    markClean,
    reset
  }
}
