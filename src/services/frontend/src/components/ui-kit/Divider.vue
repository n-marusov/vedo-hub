<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Divider, ExpandableSection, TableColumn, Spacer -->

<!-- Divider.vue -->
<template>
  <hr :class="['divider', `divider--${orientation}`]" role="separator" :aria-orientation="orientation" />
</template>
<script setup lang="ts">
defineProps<{ orientation?: 'horizontal' | 'vertical' }>()
</script>
<style scoped>
.divider { border: none; }
.divider--horizontal { border-top: 1px solid var(--border-default); margin: var(--spacing-4) 0; }
.divider--vertical { border-left: 1px solid var(--border-default); height: 100%; margin: 0 var(--spacing-2); }
</style>
