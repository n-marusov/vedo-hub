<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Spacer component — flexible spacing utility -->
<template>
  <div :class="['spacer', `spacer--${size}`]" :style="customStyle" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  custom?: string
}>()

const customStyle = computed(() => {
  if (props.custom) return { height: props.custom }
  return {}
})
</script>

<style scoped>
.spacer { width: 100%; }
.spacer--xs { height: var(--spacing-1); }
.spacer--sm { height: var(--spacing-2); }
.spacer--md { height: var(--spacing-4); }
.spacer--lg { height: var(--spacing-8); }
.spacer--xl { height: var(--spacing-16); }
</style>
