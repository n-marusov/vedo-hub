<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Metadata preview — auto-captured system metadata displayed before submission -->
<!-- @hlv:sec [SECRET_HANDLING] — metadata must not contain secrets, tokens, credentials, or PII -->

<template>
  <details class="metadata-preview">
    <summary class="metadata-preview__summary">
      <span>Auto-captured metadata</span>
      <span class="metadata-preview__badge">{{ itemCount }} fields</span>
    </summary>
    <div class="metadata-preview__content">
      <dl class="metadata-preview__list">
        <div v-for="item in metadataItems" :key="item.label" class="metadata-preview__item">
          <dt class="metadata-preview__label">{{ item.label }}</dt>
          <dd class="metadata-preview__value">{{ item.value || '—' }}</dd>
        </div>
      </dl>
    </div>
  </details>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// @ctx: metadata capture — version, environment, user_agent, trace_id, page_url
// @hlv metadata auto-capture

const props = defineProps<{
  version?: string
  environment?: string
  userAgent?: string
  traceId?: string | null
  pageUrl?: string | null
}>()

const metadataItems = computed(() => [
  { label: 'VEDO Version', value: props.version || 'unknown' },
  { label: 'Environment', value: props.environment || 'unknown' },
  { label: 'User Agent', value: props.userAgent || 'unknown' },
  { label: 'Trace ID', value: props.traceId || 'not available' },
  { label: 'Page URL', value: props.pageUrl || 'not available' }
])

const itemCount = computed(() => metadataItems.value.length)
</script>

<style scoped>
.metadata-preview {
  border: 1px solid var(--border-default); border-radius: var(--radius-md);
  background: var(--surface-secondary);
}
.metadata-preview__summary {
  padding: var(--spacing-3) var(--spacing-4); cursor: pointer;
  font-size: var(--font-size-sm); font-weight: var(--font-weight-medium);
  display: flex; align-items: center; justify-content: space-between;
  user-select: none;
}
.metadata-preview__summary:hover { background: var(--surface-hover); }
.metadata-preview__badge {
  font-size: var(--font-size-xs); background: var(--color-info);
  color: var(--font-on-info); padding: 1px var(--spacing-2);
  border-radius: var(--radius-full);
}
.metadata-preview__content { padding: 0 var(--spacing-4) var(--spacing-4); }
.metadata-preview__list { display: grid; gap: var(--spacing-2); }
.metadata-preview__item { display: flex; justify-content: space-between; }
.metadata-preview__label { font-size: var(--font-size-xs); color: var(--text-muted); font-weight: var(--font-weight-medium); }
.metadata-preview__value { font-size: var(--font-size-xs); color: var(--font-primary); font-family: var(--font-mono); }
</style>
