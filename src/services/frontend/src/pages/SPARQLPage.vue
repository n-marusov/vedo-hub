<!-- @ctx: SPARQL page strictly mirrored from design/frontend.pen frame spqBld -->
<template>
  <div class="spq-page" role="main" aria-label="SPARQL Query Builder content">
    <section class="spq-title-row">
      <div class="spq-title-wrap">
        <Search :size="20" class="spq-title-icon" />
        <h1 class="spq-title">SPARQL Query Builder</h1>
      </div>
      <span class="spq-readonly">Read-only queries only</span>
    </section>

    <section class="spq-context" aria-label="Ontology context">
      <Folder :size="14" class="muted" />
      <span class="context-label">Querying ontology:</span>
      <span class="context-badge context-badge--primary">ProductOntology</span>
      <GitBranch :size="14" class="muted" />
      <span class="context-badge">main</span>
    </section>

    <section class="spq-editor card">
      <SPARQLQueryEditor :ontology-id="ontologyId" @run="onRunQuery" @format="onFormatQuery" />
    </section>

    <p v-if="error" class="spq-error" role="alert" aria-live="assertive">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import SPARQLQueryEditor from '@/components/organisms/SPARQLQueryEditor.vue'
import { Folder, GitBranch, Search } from 'lucide-vue-next'
import { ref } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const ontologyId = ref(route.params.id as string)
const error = ref<string | null>(null)

async function onRunQuery(): Promise<void> {
  error.value = null
}

function onFormatQuery(): void {
  error.value = null
}
</script>

<style scoped>
.spq-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.spq-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
}

.spq-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.spq-title-icon { color: var(--primary); }

.spq-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

.spq-readonly {
  border-radius: 999px;
  border: 1px solid var(--info);
  color: var(--info);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  padding: 4px 10px;
}

.spq-context {
  border-radius: 6px;
  background: rgba(20, 20, 20, 0.3);
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.context-label {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.context-badge {
  border-radius: 999px;
  border: 1px solid var(--border);
  background: #000;
  padding: 2px 8px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.context-badge--primary {
  border-color: var(--primary);
  color: var(--primary);
}

.spq-editor {
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--card);
  overflow: hidden;
}

.spq-error {
  margin: 0;
  color: var(--destructive);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.card {
  border-radius: 12px;
}

.muted { color: var(--muted-foreground); }

@media (max-width: 900px) {
  .spq-page { padding: 16px; }
  .spq-title-row { flex-direction: column; align-items: flex-start; gap: 10px; }
  .spq-context { flex-wrap: wrap; }
}
</style>
