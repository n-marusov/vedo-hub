<!-- @ctx: UI-Kit ProgressBar component — determinate progress indicator -->
<template>
  <div
    class="progress"
    role="progressbar"
    :aria-valuenow="value"
    aria-valuemin="0"
    aria-valuemax="100"
    :aria-label="label || 'Progress'"
  >
    <div class="progress__track">
      <div
        class="progress__fill"
        :class="[`progress__fill--${variant}`]"
        :style="{ width: `${clampedValue}%` }"
      ></div>
    </div>
    <span v-if="showLabel" class="progress__label">{{ clampedValue }}%</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  value: number
  variant?: 'primary' | 'success' | 'warning' | 'error'
  showLabel?: boolean
  label?: string
}>()

const clampedValue = computed(() => Math.max(0, Math.min(100, props.value)))
</script>

<style scoped>
.progress {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  width: 100%;
}

.progress__track {
  flex: 1;
  height: 8px;
  background: var(--surface-tertiary, #1a1a1a);
  border-radius: var(--radius-full, 999px);
  overflow: hidden;
}

.progress__fill {
  height: 100%;
  border-radius: var(--radius-full, 999px);
  transition: width 0.3s ease;
}

.progress__fill--primary {
  background: var(--primary, #10b981);
}

.progress__fill--success {
  background: var(--success, #10b981);
}

.progress__fill--warning {
  background: var(--warning, #f59e0b);
}

.progress__fill--error {
  background: var(--status-error, #ef4444);
}

.progress__label {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #64748b);
  font-family: 'IBM Plex Mono', monospace;
  white-space: nowrap;
  min-width: 32px;
  text-align: right;
}
</style>
