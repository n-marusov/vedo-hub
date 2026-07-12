<!-- SHACLRuleBuilder.vue -->
<template>
  <div class="shacl-builder" role="region" :aria-label="'SHACL rule builder'">
    <div class="shacl-builder__toolbar">
      <PrimaryButton @click="$emit('run-validation')">Run Validation</PrimaryButton>
    </div>
    <div class="shacl-builder__tree">
      <ExpandableSection v-for="rule in rules" :key="rule.id" :title="rule.name" :badge="rule.severity">
        <dl class="shacl-builder__rule">
          <dt>Target</dt>
          <dd>{{ rule.target }}</dd>
          <dt>Condition</dt>
          <dd><code>{{ rule.condition }}</code></dd>
          <dt>Action</dt>
          <dd>{{ rule.action }}</dd>
        </dl>
      </ExpandableSection>
    </div>
  </div>
</template>

<script setup lang="ts">
import ExpandableSection from '../ui-kit/ExpandableSection.vue'
import PrimaryButton from '../ui-kit/PrimaryButton.vue'

defineProps<{
  rules: Array<{
    id: string
    name: string
    severity: string
    target: string
    condition: string
    action: string
  }>
}>()
defineEmits<{ 'run-validation': [] }>()
</script>

<style scoped>
.shacl-builder { padding: var(--spacing-4); }
.shacl-builder__toolbar { margin-bottom: var(--spacing-4); }
.shacl-builder__tree { display: flex; flex-direction: column; gap: var(--spacing-2); }
.shacl-builder__rule { display: flex; flex-direction: column; gap: var(--spacing-1); padding: var(--spacing-2) 0; font-size: var(--font-size-sm); }
.shacl-builder__rule dt { font-size: var(--font-size-xs); color: var(--text-muted); }
.shacl-builder__rule dd code { font-family: var(--font-family-mono); background: var(--surface-secondary); padding: var(--spacing-1) var(--spacing-2); border-radius: var(--radius-sm); }
</style>
