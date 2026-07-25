<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Molecule/CommentItem — individual comment entry with body text aligned under the author name -->
<!-- Matches design/ui-kit.lib.pen mCommentItem (Molecule/CommentItem) -->
<template>
  <div class="comment-item" role="listitem" :aria-label="`Comment by ${author}`">
    <div class="cm-header">
      <Avatar :display-name="author" :src="avatarSrc" size="sm" />
      <div class="cm-author-wrap">
        <span class="cm-author">{{ author }}</span>
        <span class="cm-handle">{{ handle }}</span>
      </div>
      <div class="cm-spacer"></div>
      <span class="cm-timestamp">{{ timestamp }}</span>
    </div>

    <div class="cm-body-group">
      <p class="cm-action">{{ action }}</p>
      <p class="cm-text">{{ text }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import Avatar from '../ui-kit/Avatar.vue'

// @hlv:sec [INPUT_VALIDATION] — comment data props with safe defaults
defineProps<{
  author: string
  handle: string
  timestamp: string
  action: string
  text: string
  avatarSrc?: string
}>()
</script>

<style scoped>
.comment-item {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  transition: background var(--transition-fast, 0.15s ease);
}

.comment-item:hover {
  background: var(--muted);
}

.cm-header {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.cm-author-wrap {
  display: flex;
  align-items: center;
  gap: 4px;
}

.cm-author {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 600;
  color: var(--foreground);
}

.cm-handle {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
}

.cm-spacer {
  flex: 1;
  min-width: 0;
}

.cm-timestamp {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  text-align: right;
  white-space: nowrap;
}

.cm-body-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-left: 36px;
}

.cm-action {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  line-height: 1.4;
  margin: 0;
}

.cm-text {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--foreground);
  line-height: 1.4;
  margin: 0;
}
</style>
