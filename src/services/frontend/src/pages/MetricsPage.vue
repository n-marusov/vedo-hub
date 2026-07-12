<!-- @ctx: Metrics page strictly mirrored from design/frontend.pen frame metDas -->
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
      <span class="context-badge context-badge--primary">ProductOntology</span>
      <GitBranch :size="14" class="muted" />
      <span class="context-badge">main</span>
      <Calendar :size="14" class="muted" />
      <span class="context-time">Last updated: May 15, 2026</span>
    </section>

    <section class="metrics-content card">
      <div class="kpi-grid">
        <article v-for="kpi in kpis" :key="kpi.label" class="kpi-card">
          <span class="kpi-label">{{ kpi.label }}</span>
          <strong class="kpi-value">{{ kpi.value }}</strong>
          <span class="kpi-sub">{{ kpi.sub }}</span>
        </article>
      </div>

      <div class="charts-grid">
        <article class="chart-card">
          <h2>Trend overview</h2>
          <div class="chart-placeholder">Activity chart</div>
        </article>
        <article class="chart-card">
          <h2>Validation distribution</h2>
          <div class="chart-placeholder">Distribution chart</div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Calendar, ChartColumn, Folder, GitBranch } from 'lucide-vue-next'

const kpis = [
  { label: 'Classes', value: '42', sub: '+3 this week' },
  { label: 'Properties', value: '18', sub: '+1 this week' },
  { label: 'Individuals', value: '156', sub: '+24 this week' },
  { label: 'Axioms', value: '890', sub: '+48 this week' }
]
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

@media (max-width: 1100px) {
  .kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .charts-grid { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .metrics-page { padding: 16px; }
  .kpi-grid { grid-template-columns: 1fr; }
}
</style>
