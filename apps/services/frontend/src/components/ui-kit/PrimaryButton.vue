<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit PrimaryButton component -->
<template>
  <button
    :class="[
      'btn',
      'btn--primary',
      { 'btn--disabled': disabled, 'btn--loading': loading }
    ]"
    :disabled="disabled || loading"
    :aria-disabled="disabled || loading"
    :aria-busy="loading"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" class="btn__spinner" aria-hidden="true"></span>
    <slot />
  </button>
</template>

<script setup lang="ts">
defineProps<{
	disabled?: boolean;
	loading?: boolean;
}>();

defineEmits<{
	click: [event: MouseEvent];
}>();
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

.btn--primary {
  background-color: var(--primary);
  color: var(--text-inverse);
}

.btn--primary:hover:not(.btn--disabled) {
  background-color: var(--primary-hover);
}

.btn--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn__spinner {
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
