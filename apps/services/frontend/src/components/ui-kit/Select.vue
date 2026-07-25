<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Select component — accessible dropdown select -->
<template>
  <div :class="['select-group', { 'select-group--error': error }]">
    <label :for="selectId" class="select-group__label">{{ label }}</label>
    <select
      :id="selectId"
      :value="modelValue"
      :disabled="disabled"
      :aria-invalid="!!error"
      class="select"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-for="option in options" :key="option.value" :value="option.value">
        {{ option.label }}
      </option>
    </select>
    <span v-if="error" class="select-group__error" role="alert">{{ error }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue?: string
  label: string
  options: Array<{ value: string; label: string }>
  disabled?: boolean
  error?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const selectId = computed(() => `select-${props.label.toLowerCase().replace(/\s+/g, '-')}`)
</script>

<style scoped>
.select-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.select-group__label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
}

.select {
  padding: var(--spacing-2) var(--spacing-3);
  font-size: var(--font-size-base);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background-color: var(--surface-primary);
  color: var(--text-primary);
  cursor: pointer;
}

.select:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 2px var(--primary-muted);
}

.select-group--error .select {
  border-color: var(--border-error);
}

.select-group__error {
  font-size: var(--font-size-xs);
  color: var(--status-error);
}
</style>
