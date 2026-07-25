<!-- @hlv:artifact code-frontend implements spec-feedback-001 -->
<!-- @ctx: Product digest — metrics display for PM and Support Lead roles -->
<!-- @hlv:sec [AUTH_BOUNDARY] — accessible only to Product Manager and Support Lead roles -->

<template>
  <div class="product-digest" aria-labelledby="digest-title">
    <h2 id="digest-title" class="product-digest__title">Product Feedback Digest</h2>

    <!-- @ctx: access denied for non-PM roles -->
    <!-- @hlv atomicity — role-based access to digest -->
    <div v-if="!hasAccess" class="product-digest__denied" role="alert">
      <p>Access denied. Product digest is available to Product Managers and Support Leads only.</p>
    </div>

    <template v-else>
      <!-- Period selector -->
      <div class="product-digest__period">
        <button
          v-for="p in periods"
          :key="p"
          class="product-digest__period-btn"
          :class="{ 'product-digest__period-btn--active': period === p }"
          @click="period = p"
        >
          {{ periodLabels[p] }}
        </button>
      </div>

      <!-- Loading state -->
      <div v-if="loading" class="product-digest__loading">Loading metrics...</div>

      <!-- Empty state -->
      <div v-else-if="!hasData" class="product-digest__empty">
        <p>No feedback received this period.</p>
      </div>

      <!-- Metrics -->
      <template v-else>
        <!-- Summary cards -->
        <div class="product-digest__cards">
          <div class="product-digest__card">
            <span class="product-digest__card-value">{{ metrics?.total }}</span>
            <span class="product-digest__card-label">Total Feedback</span>
          </div>
          <div class="product-digest__card">
            <span class="product-digest__card-value">{{ metrics?.nps_score }}</span>
            <span class="product-digest__card-label">NPS Score</span>
          </div>
          <div class="product-digest__card">
            <span class="product-digest__card-value">{{ metrics?.ticket_linked }}</span>
            <span class="product-digest__card-label">Linked to Tickets</span>
          </div>
        </div>

        <!-- By type breakdown -->
        <div class="product-digest__section">
          <h3 class="product-digest__section-title">Feedback by Type</h3>
          <div class="product-digest__bars">
            <div v-for="(count, type) in metrics?.by_type" :key="type" class="product-digest__bar">
              <span class="product-digest__bar-label">{{ typeLabels[type] || type }}</span>
              <div class="product-digest__bar-track">
                <div
                  class="product-digest__bar-fill"
                  :style="{ width: barPercent(count, metrics?.total || 0) + '%' }"
                />
              </div>
              <span class="product-digest__bar-value">{{ count }}</span>
            </div>
          </div>
        </div>

        <!-- NPS breakdown -->
        <div class="product-digest__section">
          <h3 class="product-digest__section-title">NPS Breakdown</h3>
          <div class="product-digest__nps-breakdown">
            <div class="product-digest__nps-item">
              <span class="product-digest__nps-label">Promoters (9-10)</span>
              <span class="product-digest__nps-value">{{ metrics?.nps_promoters }}</span>
            </div>
            <div class="product-digest__nps-item">
              <span class="product-digest__nps-label">Passives (7-8)</span>
              <span class="product-digest__nps-value">{{ metrics?.nps_passives }}</span>
            </div>
            <div class="product-digest__nps-item">
              <span class="product-digest__nps-label">Detractors (0-6)</span>
              <span class="product-digest__nps-value">{{ metrics?.nps_detractors }}</span>
            </div>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{
  userRole: string
  metrics?: {
    total: number
    nps_score: number
    ticket_linked: number
    by_type: Record<string, number>
    nps_promoters: number
    nps_passives: number
    nps_detractors: number
  }
  loading?: boolean
}>()

const period = ref<'7d' | '30d' | '90d'>('30d')
const periods = ['7d', '30d', '90d'] as const
const periodLabels: Record<string, string> = {
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
  '90d': 'Last 90 days'
}

const typeLabels: Record<string, string> = {
  bug: 'Bug',
  suggestion: 'Suggestion',
  ux: 'UX Issue',
  question: 'Question',
  praise: 'Praise'
}

// @ctx: role-based access check
const hasAccess = computed(() => {
  return props.userRole === 'ProductOwner' || props.userRole === 'SupportEngineer'
})

const hasData = computed(() => {
  return props.metrics && props.metrics.total > 0
})

function barPercent(value: number, total: number): number {
  if (total === 0) return 0
  return Math.round((value / total) * 100)
}
</script>

<style scoped>
.product-digest { display: flex; flex-direction: column; gap: var(--spacing-4); padding: var(--spacing-4); }
.product-digest__title { font-size: var(--font-size-lg); font-weight: var(--font-weight-semibold); margin: 0; }
.product-digest__denied { padding: var(--spacing-4); background: var(--color-error); color: var(--font-on-error); border-radius: var(--radius-md); }
.product-digest__period { display: flex; gap: var(--spacing-2); }
.product-digest__period-btn {
  padding: var(--spacing-1) var(--spacing-3); border: 1px solid var(--border-default);
  border-radius: var(--radius-sm); font-size: var(--font-size-sm); background: transparent; cursor: pointer;
}
.product-digest__period-btn--active { background: var(--color-primary); color: var(--font-on-primary); border-color: var(--color-primary); }
.product-digest__loading { padding: var(--spacing-6); text-align: center; color: var(--text-muted); }
.product-digest__empty { padding: var(--spacing-6); text-align: center; color: var(--text-muted); }
.product-digest__cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--spacing-3); }
.product-digest__card {
  padding: var(--spacing-4); background: var(--surface-secondary); border-radius: var(--radius-md);
  display: flex; flex-direction: column; align-items: center; gap: var(--spacing-1);
}
.product-digest__card-value { font-size: var(--font-size-xl); font-weight: var(--font-weight-bold); }
.product-digest__card-label { font-size: var(--font-size-xs); color: var(--text-muted); }
.product-digest__section { display: flex; flex-direction: column; gap: var(--spacing-2); }
.product-digest__section-title { font-size: var(--font-size-sm); font-weight: var(--font-weight-semibold); margin: 0; }
.product-digest__bars { display: flex; flex-direction: column; gap: var(--spacing-2); }
.product-digest__bar { display: flex; align-items: center; gap: var(--spacing-2); }
.product-digest__bar-label { width: 100px; font-size: var(--font-size-sm); }
.product-digest__bar-track { flex: 1; height: 8px; background: var(--surface-tertiary); border-radius: var(--radius-full); overflow: hidden; }
.product-digest__bar-fill { height: 100%; background: var(--color-primary); border-radius: var(--radius-full); transition: width 0.3s; }
.product-digest__bar-value { width: 40px; text-align: right; font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); }
.product-digest__nps-breakdown { display: flex; gap: var(--spacing-4); }
.product-digest__nps-item { display: flex; flex-direction: column; gap: var(--spacing-1); }
.product-digest__nps-label { font-size: var(--font-size-sm); }
.product-digest__nps-value { font-size: var(--font-size-lg); font-weight: var(--font-weight-bold); }
</style>
