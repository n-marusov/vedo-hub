<!-- OntologyMetadata.vue -->
<template>
  <div class="ontology-metadata" role="region" :aria-label="'Ontology metadata'">
    <h3 class="ontology-metadata__title">Metadata</h3>
    <dl class="ontology-metadata__list">
      <div v-for="(value, key) in metadata" :key="key" class="ontology-metadata__item">
        <dt>{{ formatKey(key) }}</dt>
        <dd>{{ value }}</dd>
      </div>
    </dl>
  </div>
</template>

<script setup lang="ts">
defineProps<{ metadata: Record<string, string> }>()

function formatKey(key: string): string {
  return key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}
</script>

<style scoped>
.ontology-metadata { padding: var(--spacing-4); }
.ontology-metadata__title { font-size: var(--font-size-sm); font-weight: var(--font-weight-semibold); margin-bottom: var(--spacing-3); }
.ontology-metadata__list { display: flex; flex-direction: column; gap: var(--spacing-2); }
.ontology-metadata__item dt { font-size: var(--font-size-xs); color: var(--text-muted); }
.ontology-metadata__item dd { font-size: var(--font-size-sm); }
</style>
