<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit GhostButton component -->
<template>
  <button
    :class="[
      'btn',
      'btn--ghost',
      { 'btn--disabled': disabled }
    ]"
    :disabled="disabled"
    :aria-disabled="disabled"
    @click="$emit('click', $event)"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
defineProps<{
  disabled?: boolean
}>()

defineEmits<{
  click: [event: MouseEvent]
}>()
</script>

<style scoped>
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-md);
  border: 0;
  font-family: var(--font-family-mono);
  transition: all var(--transition-fast);
  cursor: pointer;
}

.btn--ghost {
  background-color: transparent;
  color: var(--text-secondary);
}

.btn--ghost:hover:not(.btn--disabled) {
  background-color: var(--surface-variant);
  color: var(--text-primary);
}

.btn--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
