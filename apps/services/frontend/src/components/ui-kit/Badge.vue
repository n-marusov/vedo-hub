<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Badge component — status/role indicator -->
<template>
  <span
    :class="['badge', `badge--${variant}`]"
    :aria-label="label ? `${label}: ${text}` : text"
  >
    <span v-if="dot" class="badge__dot" :aria-hidden="true"></span>
    {{ text }}
  </span>
</template>

<script setup lang="ts">
defineProps<{
  text: string
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info'
  dot?: boolean
  label?: string
}>()
</script>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  padding: var(--spacing-1) var(--spacing-2);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-full);
  line-height: 1;
}

.badge--default {
  background-color: var(--surface-tertiary);
  color: var(--text-secondary);
}

.badge--success {
  background-color: var(--status-success-bg);
  color: var(--status-success);
}

.badge--warning {
  background-color: var(--status-warning-bg);
  color: var(--status-warning);
}

.badge--error {
  background-color: var(--status-error-bg);
  color: var(--status-error);
}

.badge--info {
  background-color: var(--status-info-bg);
  color: var(--status-info);
}

.badge__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: currentColor;
}
</style>
