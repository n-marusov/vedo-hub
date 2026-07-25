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
      <SPARQLQueryEditor
        :query="queryText"
        :results="resultsData"
        @update:query="queryText = $event"
        @run="onRunQuery"
        @format="onFormatQuery"
        @export="onExportQuery"
      />
    </section>

    <div v-if="loading" class="spq-loading" role="status" aria-live="polite">
      <span class="spq-loading-text">Executing query...</span>
    </div>

    <div v-if="results" class="query-results-table">
      <table class="sparql-editor__table">
        <thead>
          <tr>
            <th v-for="col in results.columns" :key="col" scope="col">{{ col }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, i) in results.rows" :key="i">
            <td v-for="(val, j) in row" :key="j">{{ val ?? '\u2014' }}</td>
          </tr>
        </tbody>
      </table>
      <div class="spq-results-summary">
        {{ results.total }} results in {{ results.executionTimeMs }}ms
      </div>
    </div>

    <p v-if="error" class="spq-error" role="alert" aria-live="assertive">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
// @aif — Migrated from Apollo GraphQL `SPARQL_EXECUTE_QUERY` (sparqlQuery resolver)
// to REST `POST /api/v1/sparql`. Per ADR-DES.API.rest-graphql-mutation-boundary.md,
// SPARQL execution is REST-only so it goes through the gateway DoS defenses
// (CircuitBreakerMiddleware, rate limiting, query complexity checks).
import { executeSparql } from "@/api/sparql";
import SPARQLQueryEditor from "@/components/organisms/SPARQLQueryEditor.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { Folder, GitBranch, Search } from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const ontologyId = ref(route.params.id as string);
const error = ref<string | null>(null);
const queryText = ref("");
const results = ref<{
	columns: string[];
	rows: unknown[][];
	total: number;
	executionTimeMs: number;
} | null>(null);
const loading = ref(false);
const { addError } = useErrorPresentation();

// Transform REST columns/rows format → SPARQL JSON format (head.vars / results.bindings)
// Same shape as the prior GraphQL path so SPARQLQueryEditor needs no changes.
const resultsData = computed(() => {
	if (!results.value) return undefined;
	const r = results.value;
	return {
		head: { vars: r.columns },
		results: {
			bindings: r.rows.map((row) => {
				const binding: Record<string, { value: string }> = {};
				r.columns.forEach((col, idx) => {
					binding[col] = {
						value:
							row[idx] !== null && row[idx] !== undefined
								? String(row[idx])
								: "",
					};
				});
				return binding;
			}),
		},
		total_results: r.total,
		execution_time_ms: r.executionTimeMs,
	};
});

async function onRunQuery(q?: string): Promise<void> {
	const sparqlQuery = q || queryText.value;
	if (!sparqlQuery || !sparqlQuery.trim()) return;

	error.value = null;
	results.value = null;
	loading.value = true;

	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "sparql.query.executing",
			ontologyId: ontologyId.value,
			queryLength: sparqlQuery.length,
			ts: new Date().toISOString(),
		}),
	);

	try {
		// REST call — goes through API Gateway DoS protection (CircuitBreakerMiddleware).
		const data = await executeSparql({
			ontologyId: ontologyId.value,
			query: sparqlQuery,
			limit: 100,
			offset: 0,
		});

		results.value = {
			columns: data.columns || [],
			rows: data.rows || [],
			total: data.total || 0,
			executionTimeMs: data.executionTimeMs || 0,
		};
	} catch (err) {
		const message = err instanceof Error ? err.message : String(err);
		error.value = message;
		addError("SPARQL_EXECUTION_ERROR", message);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "sparql.query.failed",
				ontologyId: ontologyId.value,
				error: message,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		loading.value = false;
	}
}

function onExportQuery(_q: string): void {
	if (!results.value) return;
	const format = "csv";
	let content = "";
	const cols = results.value.columns;
	if (format === "csv") {
		content = `${cols.join(",")}\n`;
		content += results.value.rows
			.map((row) =>
				row.map((v) => `"${String(v ?? "").replace(/"/g, '""')}"`).join(","),
			)
			.join("\n");
	} else {
		content = JSON.stringify(
			{ columns: cols, rows: results.value.rows, total: results.value.total },
			null,
			2,
		);
	}
	const blob = new Blob([content], { type: "text/csv" });
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = `sparql-results.${format}`;
	a.click();
	URL.revokeObjectURL(url);

	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "SPARQL.export",
			format,
			rows: results.value.total,
			ts: new Date().toISOString(),
		}),
	);
}

function onFormatQuery(q: string): void {
	if (!q || !q.trim()) return;
	// Simple SPARQL formatting — uppercase keywords, normalize whitespace, newlines before clauses
	const formatted = q
		.replace(/\s+/g, " ")
		.trim()
		.replace(
			/\b(select|where|filter|limit|offset|order\s+by|group\s+by|having|optional|union|minus|bind|values|distinct|reduced|as|desc|asc|prefix|base|construct|describe|ask|from|named|graph|service|sameTerm|isIRI|isBlank|isLiteral|str|lang|datatype|bound|if|coalesce|exists|not\s+exists|in|not\s+in|replace|regex|substr|strlen|ucase|lcase|encode_for_uri|contains|strstarts|strends|abs|round|ceil|floor|rand|now|year|month|day|hours|minutes|seconds|timezone|tz|md5|sha1|sha256|sha384|sha512|true|false|a)\b/gi,
			(match: string) => match.toUpperCase(),
		)
		.replace(
			/(SELECT|WHERE|FILTER|LIMIT|OFFSET|ORDER BY|GROUP BY|HAVING|OPTIONAL|UNION|MINUS|BIND|VALUES|DISTINCT|REDUCED|PREFIX|BASE|CONSTRUCT|DESCRIBE|ASK|FROM|NAMED|GRAPH|SERVICE)\s/g,
			"\n$1 ",
		)
		.replace(/\n\s*\n/g, "\n");
	queryText.value = formatted;
	error.value = null;

	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "sparql.query.formatted",
			ts: new Date().toISOString(),
		}),
	);
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

.spq-loading {
  padding: 12px 16px;
  border-radius: 8px;
  background: rgba(59, 130, 246, 0.08);
  border: 1px solid rgba(59, 130, 246, 0.2);
  display: flex;
  align-items: center;
  gap: 8px;
}

.spq-loading-text {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--info);
}

.query-results-table {
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--card);
  overflow: hidden;
}

.query-results-table :deep(table) {
  width: 100%;
  border-collapse: collapse;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.query-results-table :deep(th) {
  text-align: left;
  padding: 8px 12px;
  background: rgba(20, 20, 20, 0.5);
  border-bottom: 1px solid var(--border);
  font-weight: 600;
  white-space: nowrap;
}

.query-results-table :deep(td) {
  padding: 6px 12px;
  border-bottom: 1px solid var(--border);
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spq-results-summary {
  padding: 8px 12px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  border-top: 1px solid var(--border);
}

@media (max-width: 900px) {
  .spq-page { padding: 16px; }
  .spq-title-row { flex-direction: column; align-items: flex-start; gap: 10px; }
  .spq-context { flex-wrap: wrap; }
}
</style>
