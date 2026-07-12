<!-- ValidationReport.vue -->
<template>
  <div class="validation-report" role="region" :aria-label="'Validation report'">
    <div class="validation-report__summary">
      <Badge :text="`Passed: ${summary.passed}`" variant="success" />
      <Badge :text="`Failed: ${summary.failed}`" variant="error" />
      <Badge :text="`Warnings: ${summary.warnings}`" variant="warning" />
      <span class="validation-report__total">Total: {{ summary.total_rules }}</span>
    </div>
    <table class="validation-report__table">
      <thead>
        <tr>
          <th scope="col">Rule</th>
          <th scope="col">Severity</th>
          <th scope="col">Focus</th>
          <th scope="col">Message</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="result in results" :key="result.rule_id">
          <td>{{ result.rule_name }}</td>
          <td><Badge :text="result.severity" :variant="result.severity === 'error' ? 'error' : 'warning'" /></td>
          <td><code>{{ result.focus_node }}</code></td>
          <td>{{ result.message }}</td>
        </tr>
      </tbody>
    </table>
    <div v-if="!results.length" class="validation-report__empty">
      All rules passed
    </div>
  </div>
</template>

<script setup lang="ts">
import Badge from '../ui-kit/Badge.vue'

defineProps<{
  summary: { total_rules: number; passed: number; failed: number; warnings: number }
  results: Array<{
    rule_id: string
    rule_name: string
    severity: string
    focus_node: string
    message: string
  }>
}>()
</script>

<style scoped>
.validation-report { padding: var(--spacing-4); }
.validation-report__summary { display: flex; align-items: center; gap: var(--spacing-2); margin-bottom: var(--spacing-4); }
.validation-report__total { font-size: var(--font-size-sm); color: var(--text-secondary); }
.validation-report__table { font-size: var(--font-size-sm); }
.validation-report__table th, .validation-report__table td { padding: var(--spacing-2); border-bottom: 1px solid var(--border-default); text-align: left; }
.validation-report__empty { padding: var(--spacing-8); text-align: center; color: var(--status-success); font-weight: var(--font-weight-medium); }
</style>
