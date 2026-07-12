<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit IconButton, Checkbox, Radio, Dialog, BadgeInfo, Tab, Divider, ExpandableSection, TableColumn, Spacer -->

<!-- IconButton.vue -->
<template>
  <button
    :class="['icon-btn', { 'icon-btn--disabled': disabled }]"
    :disabled="disabled"
    :aria-label="label"
    :aria-disabled="disabled"
    @click="$emit('click', $event)"
  >
    <slot />
  </button>
</template>
<script setup lang="ts">
defineProps<{ label: string; disabled?: boolean }>()
defineEmits<{ click: [event: MouseEvent] }>()
</script>
<style scoped>
.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}
.icon-btn:hover:not(.icon-btn--disabled) {
  background-color: var(--surface-secondary);
  color: var(--text-primary);
}
.icon-btn--disabled { opacity: 0.5; cursor: not-allowed; }
</style>
