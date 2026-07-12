<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Dialog, BadgeInfo, Tab, Divider, ExpandableSection, TableColumn, Spacer -->

<!-- Dialog.vue -->
<template>
  <Teleport to="body">
    <div v-if="open" class="dialog-overlay" @click.self="onOverlayClick" role="dialog" :aria-modal="modal" :aria-labelledby="titleId">
      <div class="dialog" :class="`dialog--${size}`">
        <div class="dialog__header">
          <h2 :id="titleId" class="dialog__title">{{ title }}</h2>
          <button class="dialog__close" aria-label="Close dialog" @click="$emit('close')">✕</button>
        </div>
        <div class="dialog__body"><slot /></div>
        <div v-if="$slots.footer" class="dialog__footer"><slot name="footer" /></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{
  open: boolean
  title: string
  size?: 'sm' | 'md' | 'lg'
  modal?: boolean
}>()
const emit = defineEmits<{ close: [] }>()
const titleId = computed(() => `dialog-title-${props.title.toLowerCase().replace(/\s+/g, '-')}`)
function onOverlayClick() {
  if (props.modal) return
  emit('close')
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed; inset: 0; background: var(--surface-overlay);
  display: flex; align-items: center; justify-content: center; z-index: var(--z-modal);
}
.dialog {
  background: var(--surface-primary); border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl); max-height: 90vh; overflow: auto;
}
.dialog--sm { width: 400px; } .dialog--md { width: 600px; } .dialog--lg { width: 800px; }
.dialog__header { display: flex; align-items: center; justify-content: space-between; padding: var(--spacing-4) var(--spacing-6); border-bottom: 1px solid var(--border-default); }
.dialog__title { font-size: var(--font-size-lg); font-weight: var(--font-weight-semibold); }
.dialog__close { font-size: var(--font-size-lg); color: var(--text-muted); cursor: pointer; }
.dialog__body { padding: var(--spacing-6); }
.dialog__footer { padding: var(--spacing-4) var(--spacing-6); border-top: 1px solid var(--border-default); display: flex; justify-content: flex-end; gap: var(--spacing-2); }
</style>
