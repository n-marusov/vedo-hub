<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism OntologyToolbar component — toolbar above workspace with branch, commit, actions -->
<template>
  <div class="ontology-toolbar" role="toolbar" :aria-label="'Ontology workspace toolbar'">
    <div class="ontology-toolbar__left">
      <span class="ontology-toolbar__ontology">{{ ontologyName }}</span>
      <Badge v-if="versionContext.dirty" text="Unsaved changes" variant="warning" dot />
    </div>

    <div class="ontology-toolbar__center">
      <span class="ontology-toolbar__branch">{{ versionContext.branch || 'main' }}</span>
      <span class="ontology-toolbar__commit">{{ shortCommit }}</span>
    </div>

    <div class="ontology-toolbar__right">
      <PrimaryButton :loading="saving" @click="$emit('save')">Save</PrimaryButton>
      <GhostButton @click="$emit('publish')">Publish</GhostButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Badge from '../ui-kit/Badge.vue'
import GhostButton from '../ui-kit/GhostButton.vue'
import PrimaryButton from '../ui-kit/PrimaryButton.vue'

const props = defineProps<{
  ontologyName: string
  versionContext: {
    branch?: string
    commit?: string
    dirty?: boolean
  }
  saving?: boolean
}>()

defineEmits<{
  save: []
  publish: []
}>()

const shortCommit = computed(() => {
  return props.versionContext.commit?.slice(0, 7) || '—'
})
</script>

<style scoped>
.ontology-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--toolbar-height);
  padding: 0 var(--spacing-4);
  background: var(--surface-secondary);
  border-bottom: 1px solid var(--border-default);
}

.ontology-toolbar__left,
.ontology-toolbar__center,
.ontology-toolbar__right {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.ontology-toolbar__ontology {
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
}

.ontology-toolbar__branch {
  font-size: var(--font-size-xs);
  padding: var(--spacing-1) var(--spacing-2);
  background: var(--surface-primary);
  border-radius: var(--radius-md);
  font-family: var(--font-family-mono);
}

.ontology-toolbar__commit {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  font-family: var(--font-family-mono);
}
</style>
