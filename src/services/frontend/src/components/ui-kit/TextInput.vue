<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit TextInput component — accessible form input with label -->
<template>
  <div :class="['input-group', { 'input-group--error': error }]">
    <label :for="inputId" class="input-group__label">{{ label }}</label>
    <input
      :id="inputId"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :aria-invalid="!!error"
      :aria-describedby="error ? errorId : undefined"
      class="input"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <span v-if="error" :id="errorId" class="input-group__error" role="alert">
      {{ error }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue?: string
  label: string
  type?: string
  placeholder?: string
  disabled?: boolean
  error?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const inputId = computed(() => `input-${props.label.toLowerCase().replace(/\s+/g, '-')}`)
const errorId = computed(() => `${inputId.value}-error`)
</script>

<style scoped>
.input-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.input-group__label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
}

.input {
  padding: var(--spacing-2) var(--spacing-3);
  font-size: var(--font-size-base);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background-color: var(--surface-primary);
  color: var(--text-primary);
  transition: border-color var(--transition-fast);
}

.input:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 2px var(--primary-muted);
}

.input-group--error .input {
  border-color: var(--border-error);
}

.input-group__error {
  font-size: var(--font-size-xs);
  color: var(--status-error);
}
</style>
