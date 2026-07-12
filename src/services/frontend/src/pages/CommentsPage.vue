<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Comments page — full-page view aligned to design/pages/comments.pen -->
<!-- Matches design/pages/comments.pen frame commentsPage -->
<template>
    <div class="comments-page" role="main" aria-label="Comments page">
        <div class="cm-title-col">
            <div class="cm-breadcrumbs">
                <span class="cm-crumb">Workspace</span>
                <ChevronRight :size="12" class="cm-crumb-sep" />
                <span class="cm-crumb">Comments</span>
            </div>
            <h1 class="cm-page-title">Comments</h1>
        </div>

        <Comments class="cm-section" />
    </div>
</template>

<script setup lang="ts">
import Comments from "@/components/organisms/Comments.vue";
import { ChevronRight } from "lucide-vue-next";
</script>

<style scoped>
.comments-page {
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.cm-title-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.cm-breadcrumbs {
    display: flex;
    align-items: center;
    gap: 8px;
}

.cm-crumb {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 500;
    color: var(--muted-foreground);
}

.cm-crumb-sep {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.cm-page-title {
    margin: 0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 24px;
    font-weight: 600;
    color: var(--foreground);
}

.cm-section {
    width: 100%;
}
</style>
