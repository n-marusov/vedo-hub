<!-- @ctx: Metrics page strictly mirrored from design/frontend.pen frame metDas -->
<!-- @m2.5 — Wired to ONTOLOGY_METRICS_QUERY via Apollo GraphQL -->
<template>
  <div class="metrics-page" role="main" aria-label="Metrics Dashboard content">
    <section class="metrics-title-row">
      <div class="metrics-title-wrap">
        <ChartColumn :size="20" class="primary" />
        <h1 class="metrics-title">Metrics Dashboard</h1>
      </div>
    </section>

    <section class="metrics-context">
      <Folder :size="14" class="muted" />
      <span class="context-label">Analyzing ontology:</span>
      <span class="context-badge context-badge--primary">{{ ontologyId || '—' }}</span>
      <GitBranch :size="14" class="muted" />
      <span class="context-badge">main</span>
      <Calendar :size="14" class="muted" />
      <span class="context-time">Last updated: {{ lastUpdated }}</span>
    </section>

    <!-- @m2.5 Loading state -->
    <section v-if="loading" class="metrics-content card">
      <div class="kpi-grid">
        <article v-for="n in 4" :key="n" class="kpi-card skeleton">
          <div class="skeleton-line skeleton-line--label"></div>
          <div class="skeleton-line skeleton-line--value"></div>
          <div class="skeleton-line skeleton-line--sub"></div>
        </article>
      </div>
    </section>

    <!-- @m2.5 Error state -->
    <section v-else-if="error" class="metrics-content card">
      <div class="error-state">
        <p>Failed to load metrics data.</p>
        <button class="retry-btn" type="button" @click="refetch()">Retry</button>
      </div>
    </section>

    <!-- @m2.5 Data state — KPI counters from ONTOLOGY_METRICS_QUERY -->
    <section v-else class="metrics-content card">
      <div class="kpi-grid">
        <article class="kpi-card">
          <span class="kpi-label">Classes</span>
          <strong class="kpi-value">{{ counters.classCount }}</strong>
          <span class="kpi-sub">{{ trendSummary.classCount }}</span>
        </article>
        <article class="kpi-card">
          <span class="kpi-label">Properties</span>
          <strong class="kpi-value">{{ counters.propertyCount }}</strong>
          <span class="kpi-sub">{{ trendSummary.propertyCount }}</span>
        </article>
        <article class="kpi-card">
          <span class="kpi-label">Individuals</span>
          <strong class="kpi-value">{{ counters.individualCount }}</strong>
          <span class="kpi-sub">{{ trendSummary.individualCount }}</span>
        </article>
        <article class="kpi-card">
          <span class="kpi-label">Axioms</span>
          <strong class="kpi-value">{{ counters.axiomCount }}</strong>
          <span class="kpi-sub">Total logical axioms</span>
        </article>
      </div>

      <div class="charts-grid">
        <article class="chart-card">
          <h2>Trend overview</h2>
          <div class="chart-placeholder metrics-chart">Activity chart (trends data loaded)</div>
        </article>
        <article class="chart-card">
          <h2>Validation distribution</h2>
          <div class="chart-placeholder metrics-chart">Distribution chart</div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ONTOLOGY_METRICS_QUERY } from "@/apollo/queries";
import { useQuery } from "@vue/apollo-composable";
import { Calendar, ChartColumn, Folder, GitBranch } from "lucide-vue-next";
import { computed, watch } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const ontologyId = computed(
	() =>
		(route.params.ontologyId as string) ||
		(route.query.ontologyId as string) ||
		"",
);

// @m2.5 — Wire metrics to ONTOLOGY_METRICS_QUERY
const { result, loading, error, refetch } = useQuery(
	ONTOLOGY_METRICS_QUERY,
	() => ({
		ontologyId: ontologyId.value || "default",
	}),
	{ enabled: true },
);

const counters = computed(() => {
	const c = result.value?.ontologyMetrics?.counters;
	return {
		classCount: c?.classCount ?? 0,
		propertyCount: c?.propertyCount ?? 0,
		individualCount: c?.individualCount ?? 0,
		axiomCount: c?.axiomCount ?? 0,
	};
});

const trendSummary = computed(() => {
	const trends = result.value?.ontologyMetrics?.trends;
	if (!trends || trends.length < 2) {
		return { classCount: "", propertyCount: "", individualCount: "" };
	}
	const latest = trends[trends.length - 1];
	const previous = trends[trends.length - 2];
	const diff = (field: string) => {
		const d = (latest[field] || 0) - (previous[field] || 0);
		return d >= 0 ? `+${d} this period` : `${d} this period`;
	};
	return {
		classCount: diff("classCount"),
		propertyCount: diff("propertyCount"),
		individualCount: diff("individualCount"),
	};
});

const lastUpdated = computed(() => {
	const trends = result.value?.ontologyMetrics?.trends;
	if (!trends?.length) return "N/A";
	return new Date(trends[trends.length - 1].date).toLocaleDateString("en-US", {
		month: "short",
		day: "numeric",
		year: "numeric",
	});
});

// ── Logging ─────────────────────────────────────────────────────────────────

watch(counters, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "metrics.counters.loaded",
			classCount: val.classCount,
			ts: new Date().toISOString(),
		}),
	);
});

watch(error, (err) => {
	if (err) {
		console.error(
			JSON.stringify({
				level: "error",
				msg: "metrics.query.error",
				error: String(err),
				ts: new Date().toISOString(),
			}),
		);
	}
});
</script>

<style scoped>
.metrics-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.metrics-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
}

.metrics-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.metrics-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

.metrics-context {
  border-radius: 6px;
  background: rgba(20, 20, 20, 0.3);
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.context-label,
.context-time {
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

.metrics-content {
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.kpi-card {
  border-radius: 10px;
  border: 1px solid var(--border);
  background: rgba(20, 20, 20, 0.25);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.kpi-label {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.kpi-value {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 28px;
  font-weight: 600;
}

.kpi-sub {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.chart-card {
  border-radius: 10px;
  border: 1px solid var(--border);
  background: rgba(20, 20, 20, 0.25);
  padding: 16px;
}

.chart-card h2 {
  margin: 0 0 12px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
}

.chart-placeholder {
  height: 240px;
  border-radius: 8px;
  border: 1px dashed var(--border);
  color: var(--muted-foreground);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.primary { color: var(--primary); }
.muted { color: var(--muted-foreground); }

/* @m2.5 Skeleton loading */
.skeleton { animation: pulse 1.5s ease-in-out infinite; }
.skeleton-line { height: 14px; border-radius: 4px; background: var(--muted); }
.skeleton-line--label { width: 50%; }
.skeleton-line--value { width: 35%; height: 28px; margin-top: 6px; }
.skeleton-line--sub { width: 60%; margin-top: 4px; }
@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

/* @m2.5 Error state */
.error-state {
  grid-column: 1 / -1;
  padding: 32px;
  text-align: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
}
.retry-btn {
  margin-top: 12px;
  height: 32px;
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0 12px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  background: var(--card);
  color: var(--foreground);
  cursor: pointer;
}
.retry-btn:hover { background: var(--muted); }

@media (max-width: 1100px) {
  .kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .charts-grid { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .metrics-page { padding: 16px; }
  .kpi-grid { grid-template-columns: 1fr; }
}
</style>
