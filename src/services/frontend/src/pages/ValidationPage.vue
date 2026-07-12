<!-- @ctx: Validation page strictly mirrored from design/frontend.pen frame valRep -->
<template>
  <div class="validation-page" role="main" aria-label="Validation Report content">
    <section class="validation-title-row">
      <div class="validation-title-wrap">
        <Shield :size="20" class="warning" />
        <h1 class="validation-title">Validation Report</h1>
      </div>
    </section>

    <section class="validation-context">
      <Folder :size="14" class="muted" />
      <span class="context-label">Validating ontology:</span>
      <span class="context-badge context-badge--primary">ProductOntology</span>
      <GitBranch :size="14" class="muted" />
      <span class="context-badge">main</span>
      <Calendar :size="14" class="muted" />
      <span class="context-time">Last validation: May 15, 2026 14:32:15</span>
    </section>

    <section class="validation-actions">
      <button class="run-btn" type="button" @click="runValidation">
        <Play :size="14" />
        Run validation
      </button>
    </section>

    <section class="validation-card">
      <ValidationReport :summary="summary" :results="results" />
    </section>
  </div>
</template>

<script setup lang="ts">
import ValidationReport from '@/components/organisms/ValidationReport.vue'
import { Calendar, Folder, GitBranch, Play, Shield } from 'lucide-vue-next'
import { ref } from 'vue'

const summary = ref({ total_rules: 24, passed: 19, failed: 3, warnings: 2 })
const results = ref([
  {
    rule_id: 'VAL-001',
    rule_name: 'INN length check',
    severity: 'error',
    focus_node: 'ex:Person/Alice_Johnson',
    message: "INN value '12345' does not match pattern ^\\d{12}$"
  },
  {
    rule_id: 'VAL-014',
    rule_name: 'Email format',
    severity: 'warning',
    focus_node: 'ex:Organization/Acme_Corp',
    message: 'Missing contactEmail property'
  }
])

function runValidation(): void {
  summary.value = { total_rules: 24, passed: 20, failed: 2, warnings: 2 }
}
</script>

<style scoped>
.validation-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.validation-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
}

.validation-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.validation-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

.validation-context {
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

.validation-actions {
  display: flex;
  justify-content: flex-end;
  padding-right: 8px;
}

.run-btn {
  height: 36px;
  border-radius: 8px;
  background: var(--primary);
  color: var(--primary-foreground);
  border: 1px solid var(--primary);
  padding: 0 14px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
}

.validation-card {
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card);
  overflow: hidden;
}

.warning { color: var(--warning); }
.muted { color: var(--muted-foreground); }

@media (max-width: 768px) {
  .validation-page { padding: 16px; }
  .validation-actions { justify-content: flex-start; padding-right: 0; }
}
</style>
