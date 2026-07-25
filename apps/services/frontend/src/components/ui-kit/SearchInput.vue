<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit SearchInput component -->
<template>
  <div class="search-input">
    <span class="search-input__icon" aria-hidden="true">🔍</span>
    <input
      :value="modelValue"
      :placeholder="placeholder"
      :aria-label="label"
      class="search-input__field"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue?: string
  placeholder?: string
  label?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<style scoped>
.search-input {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  background-color: var(--surface-primary);
}

.search-input__field {
  border: none;
  outline: none;
  background: transparent;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  width: 100%;
}

.search-input__icon {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}
</style>
