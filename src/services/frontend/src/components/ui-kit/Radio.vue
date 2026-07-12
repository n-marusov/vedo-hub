<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Radio component -->
<template>
  <label :class="['radio', { 'radio--disabled': disabled }]" :for="radioId">
    <input
      :id="radioId"
      type="radio"
      :name="name"
      :value="value"
      :checked="modelValue === value"
      :disabled="disabled"
      @change="$emit('update:modelValue', value)"
    />
    <span class="radio__dot" aria-hidden="true"></span>
    <span class="radio__label">{{ label }}</span>
  </label>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue?: string
  name: string
  value: string
  label: string
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const radioId = computed(() => `radio-${props.name}-${props.value}`)
</script>

<style scoped>
.radio {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  cursor: pointer;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.radio--disabled { opacity: 0.5; cursor: not-allowed; }

.radio input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.radio__dot {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-default);
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.radio input:checked + .radio__dot {
  border-color: var(--primary);
}

.radio input:checked + .radio__dot::after {
  content: '';
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
  background-color: var(--primary);
}

.radio input:focus-visible + .radio__dot {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
}
</style>
