<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Dialog, BadgeInfo, Tab, Divider, ExpandableSection, TableColumn, Spacer -->

<!-- Dialog.vue -->
<template>
  <Teleport to="body">
    <div v-if="open" class="dialog-overlay" @click.self="onOverlayClick" role="dialog" :aria-modal="modal" :aria-labelledby="titleId">
      <div class="dialog" :class="`dialog--${size}`">
        <div class="dialog__header">
          <div class="dialog__heading">
            <h2 :id="titleId" class="dialog__title">{{ title }}</h2>
            <p v-if="description" class="dialog__description">{{ description }}</p>
          </div>
          <button class="dialog__close" aria-label="Close dialog" @click="$emit('close')">✕</button>
        </div>
        <div class="dialog__body"><slot /></div>
        <div v-if="$slots.footer" class="dialog__footer"><slot name="footer" /></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from "vue";
const props = defineProps<{
	open: boolean;
	title: string;
	description?: string;
	size?: "sm" | "md" | "form" | "lg" | "xl";
	modal?: boolean;
}>();
const emit = defineEmits<{ close: [] }>();
const titleId = computed(
	() => `dialog-title-${props.title.toLowerCase().replace(/\s+/g, "-")}`,
);
function onOverlayClick() {
	if (props.modal) return;
	emit("close");
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  z-index: var(--z-modal);
}
.dialog {
  width: min(calc(100vw - 32px), 600px);
  background: var(--surface);
  color: var(--foreground);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  max-height: 90vh;
  overflow: auto;
}
.dialog--sm { width: min(calc(100vw - 32px), 400px); }
.dialog--md { width: min(calc(100vw - 32px), 600px); }
.dialog--form { width: min(calc(100vw - 32px), 620px); }
.dialog--lg { width: min(calc(100vw - 32px), 800px); }
.dialog--xl { width: min(calc(100vw - 32px), 960px); }
.dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid var(--border-default);
}
.dialog__heading {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
}
.dialog__title {
  margin: 0;
  color: var(--foreground);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}
.dialog__description {
  margin: 0;
  white-space: pre-line;
  color: var(--muted-foreground);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-normal);
}
.dialog__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 0;
  border-radius: var(--radius-md);
  background: transparent;
  font-size: var(--font-size-lg);
  color: var(--text-muted);
  cursor: pointer;
}
.dialog__close:hover {
  background: var(--surface-variant);
  color: var(--foreground);
}
.dialog__body { padding: var(--space-6); }
.dialog__footer {
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--border-default);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}
</style>
