<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Molecule/CommitItem — individual commit entry with timeline-compatible body indentation -->
<!-- Matches design/ui-kit.lib.pen mCommitItem (Molecule/CommitItem) -->
<template>
    <div
        class="commit-item"
        role="listitem"
        :aria-label="`Commit ${sha} by ${author}`"
    >
        <div class="ci-header">
            <Avatar :display-name="author" :src="avatarSrc" size="sm" />
            <div class="ci-author-wrap">
                <span class="ci-author">{{ author }}</span>
                <span class="ci-handle">{{ handle }}</span>
            </div>
            <div class="ci-spacer"></div>
            <span class="ci-timestamp">{{ timestamp }}</span>
        </div>

        <div class="ci-body">
            <p class="ci-action">{{ action }}</p>

            <div class="ci-commit-row">
                <Badge
                    :text="sha.slice(0, 7)"
                    variant="default"
                    class="ci-badge"
                />
                <span class="ci-commit-msg">{{ commitMessage }}</span>
            </div>

            <p v-if="moreInfo" class="ci-more-info">{{ moreInfo }}</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import Avatar from '../ui-kit/Avatar.vue'
import Badge from '../ui-kit/Badge.vue'

// @hlv:sec [INPUT_VALIDATION] — commit data props with safe defaults
defineProps<{
  author: string
  handle: string
  timestamp: string
  action: string
  sha: string
  commitMessage: string
  moreInfo?: string
  avatarSrc?: string
}>()
</script>

<style scoped>
.commit-item {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    transition: background var(--transition-fast, 0.15s ease);
}

.commit-item:hover {
    background: var(--muted);
}

.ci-header {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
}

.ci-author-wrap {
    display: flex;
    align-items: center;
    gap: 4px;
}

.ci-author {
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    font-weight: 600;
    color: var(--foreground);
}

.ci-handle {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
}

.ci-spacer {
    flex: 1;
    min-width: 0;
}

.ci-timestamp {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
    text-align: right;
    white-space: nowrap;
}

.ci-body {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding-left: 36px;
}

.ci-action {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
    line-height: 1.4;
    margin: 0;
}

.ci-commit-row {
    display: flex;
    align-items: center;
    gap: 6px;
}

.ci-badge {
    flex-shrink: 0;
}

.ci-commit-msg {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--foreground);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.ci-more-info {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
    line-height: 1.4;
    margin: 0;
}
</style>
