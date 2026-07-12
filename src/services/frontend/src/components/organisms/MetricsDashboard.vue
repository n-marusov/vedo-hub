<!-- MetricsDashboard.vue -->
<template>
  <div class="metrics-dashboard" role="region" :aria-label="'Ontology metrics dashboard'">
    <div class="metrics-dashboard__counters">
      <div v-for="(counter, key) in counters" :key="key" class="metrics-dashboard__card">
        <span class="metrics-dashboard__label">{{ counterLabels[key] }}</span>
        <span class="metrics-dashboard__value">{{ counter }}</span>
      </div>
    </div>
    <div v-if="trends?.dates" class="metrics-dashboard__chart">
      <h3>Trends</h3>
      <div class="metrics-dashboard__chart-area">
        <!-- Chart placeholder — integrates with charting library -->
        <div class="metrics-dashboard__chart-placeholder">
          Chart: {{ trends.dates.length }} data points
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  counters?: { classes: number; properties: number; individuals: number; axioms: number }
  trends?: {
    dates: string[]
    class_counts: number[]
    property_counts: number[]
    individual_counts: number[]
  }
}>()

const counterLabels: Record<string, string> = {
  classes: 'Classes',
  properties: 'Properties',
  individuals: 'Individuals',
  axioms: 'Axioms'
}
</script>

<style scoped>
.metrics-dashboard { padding: var(--spacing-6); }
.metrics-dashboard__counters { display: grid; grid-template-columns: repeat(4, 1fr); gap: var(--spacing-4); margin-bottom: var(--spacing-8); }
.metrics-dashboard__card { padding: var(--spacing-4); background: var(--surface-primary); border-radius: var(--radius-lg); border: 1px solid var(--border-default); }
.metrics-dashboard__label { display: block; font-size: var(--font-size-xs); color: var(--text-secondary); margin-bottom: var(--spacing-1); }
.metrics-dashboard__value { display: block; font-size: var(--font-size-3xl); font-weight: var(--font-weight-bold); }
.metrics-dashboard__chart h3 { font-size: var(--font-size-lg); margin-bottom: var(--spacing-4); }
.metrics-dashboard__chart-placeholder { height: 200px; display: flex; align-items: center; justify-content: center; background: var(--surface-secondary); border-radius: var(--radius-lg); color: var(--text-muted); }
</style>
