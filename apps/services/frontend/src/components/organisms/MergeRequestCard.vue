<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Molecule/MergeRequestCard — collapsible section card matching design/pages/merge-requests.pen -->
<!-- Matches design/ui-kit.lib.pen mMergeRequestCard (Molecule/MergeRequestCard) -->
<template>
  <div class="mr-card" :class="{ 'mr-card--closed': !open }">
    <header class="mrc-header" @click="$emit('toggle')">
      <div class="mrc-left">
        <ChevronDown :size="14" :class="['mrc-chevron', { 'mrc-chevron--closed': !open }]" />
        <span class="mrc-title">{{ title }}</span>
      </div>
      <span v-if="count !== undefined" class="mrc-badge">{{ count }}</span>
    </header>
    <div v-if="open" class="mrc-body">
      <span class="mrc-empty">{{ emptyText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChevronDown } from 'lucide-vue-next'

defineProps<{
  title: string
  open: boolean
  count?: number
  emptyText?: string
}>()

defineEmits<{
  toggle: []
}>()
</script>

<style scoped>
.mr-card {
  width: 100%;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.mrc-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
}

.mrc-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mrc-chevron {
  color: var(--muted-foreground);
  transition: transform 0.15s ease;
  flex-shrink: 0;
}

.mrc-chevron--closed {
  transform: rotate(-90deg);
}

.mrc-title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 500;
  color: var(--foreground);
}

.mrc-badge {
  min-width: 20px;
  height: 20px;
  border-radius: 10px;
  background: var(--muted);
  padding: 1px 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  font-weight: 500;
  color: var(--foreground);
}

.mrc-body {
  padding: 8px 16px 16px;
}

.mrc-empty {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}
</style>
