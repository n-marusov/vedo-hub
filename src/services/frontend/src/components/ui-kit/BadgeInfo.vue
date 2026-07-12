<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit BadgeInfo, Tab, Divider, ExpandableSection, TableColumn, Spacer -->

<!-- BadgeInfo.vue -->
<template>
  <span class="badge-info" :aria-label="text">
    <span class="badge-info__icon" aria-hidden="true">ℹ</span>
    {{ text }}
  </span>
</template>
<script setup lang="ts">
defineProps<{ text: string }>()
</script>
<style scoped>
.badge-info {
  display: inline-flex; align-items: center; gap: var(--spacing-1);
  padding: var(--spacing-1) var(--spacing-2); font-size: var(--font-size-xs);
  background-color: var(--status-info-bg); color: var(--status-info);
  border-radius: var(--radius-md);
}
.badge-info__icon { font-size: var(--font-size-sm); }
</style>
