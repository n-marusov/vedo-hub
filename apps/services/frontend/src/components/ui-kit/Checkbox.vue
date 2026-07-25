<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Checkbox component — accessible form control -->
<template>
  <label :class="['checkbox', { 'checkbox--disabled': disabled }]" :for="checkboxId">
    <input
      :id="checkboxId"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled"
      :aria-checked="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
    />
    <span class="checkbox__box" aria-hidden="true"></span>
    <span class="checkbox__label">{{ label }}</span>
  </label>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue?: boolean
  label: string
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const checkboxId = computed(() => `checkbox-${props.label.toLowerCase().replace(/\s+/g, '-')}`)
</script>

<style scoped>
.checkbox {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  cursor: pointer;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.checkbox--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.checkbox input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.checkbox__box {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-default);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.checkbox input:checked + .checkbox__box {
  background-color: var(--primary);
  border-color: var(--primary);
}

.checkbox input:checked + .checkbox__box::after {
  content: '✓';
  color: white;
  font-size: 12px;
}

.checkbox input:focus-visible + .checkbox__box {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
}
</style>
