<!-- SPARQLQueryEditor.vue -->
<template>
  <div class="sparql-editor" role="region" :aria-label="'SPARQL query editor'">
    <div class="sparql-editor__toolbar">
      <PrimaryButton :disabled="!query" @click="query && $emit('run', query)">
        ▶ Run
      </PrimaryButton>
      <GhostButton @click="query && $emit('format', query)">Format</GhostButton>
      <GhostButton :disabled="!results" @click="query && $emit('export', query)">Export</GhostButton>
      <Select v-model="format" :options="formatOptions" label="Result format" />
    </div>
    <textarea
      :value="query"
      class="sparql-editor__textarea"
      :aria-label="'SPARQL query input'"
      placeholder="SELECT ?s ?p ?o WHERE { ?s ?p ?o } LIMIT 10"
      @input="$emit('update:query', ($event.target as HTMLTextAreaElement).value)"
      @keydown.ctrl.enter="query && $emit('run', query)"
    />
    <div v-if="results" class="sparql-editor__results">
      <table class="sparql-editor__table">
        <thead>
          <tr>
            <th v-for="var_ in results.head.vars" :key="var_" scope="col">{{ var_ }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(binding, i) in results.results.bindings" :key="i">
            <td v-for="var_ in results.head.vars" :key="var_">{{ binding[var_]?.value || '—' }}</td>
          </tr>
        </tbody>
      </table>
      <div class="sparql-editor__summary">
        {{ results.total_results }} results in {{ results.execution_time_ms }}ms
      </div>
    </div>
    <div v-else-if="!query" class="sparql-editor__empty">
      Enter a SPARQL query and press Run
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import GhostButton from "../ui-kit/GhostButton.vue";
import PrimaryButton from "../ui-kit/PrimaryButton.vue";
import Select from "../ui-kit/Select.vue";

defineProps<{
	query?: string;
	results?: {
		head: { vars: string[] };
		results: { bindings: Record<string, { value: string }>[] };
		total_results: number;
		execution_time_ms: number;
	};
}>();
defineEmits<{
	"update:query": [v: string];
	run: [q: string];
	format: [q: string];
	export: [q: string];
}>();

const format = ref("table");
const formatOptions = [
	{ value: "table", label: "Table" },
	{ value: "json", label: "JSON" },
	{ value: "csv", label: "CSV" },
];
</script>

<style scoped>
.sparql-editor { display: flex; flex-direction: column; height: 100%; }
.sparql-editor__toolbar { display: flex; align-items: center; gap: var(--spacing-2); padding: var(--spacing-3); border-bottom: 1px solid var(--border-default); }
.sparql-editor__textarea { flex: 1; padding: var(--spacing-4); font-family: var(--font-family-mono); font-size: var(--font-size-sm); background: var(--surface-secondary); color: var(--text-primary); border: none; resize: none; min-height: 200px; }
.sparql-editor__results { flex: 1; overflow: auto; padding: var(--spacing-4); }
.sparql-editor__table { font-size: var(--font-size-sm); }
.sparql-editor__table th, .sparql-editor__table td { padding: var(--spacing-1) var(--spacing-2); border-bottom: 1px solid var(--border-default); text-align: left; }
.sparql-editor__summary { margin-top: var(--spacing-2); font-size: var(--font-size-xs); color: var(--text-muted); }
.sparql-editor__empty { flex: 1; display: flex; align-items: center; justify-content: center; color: var(--text-muted); }
</style>
