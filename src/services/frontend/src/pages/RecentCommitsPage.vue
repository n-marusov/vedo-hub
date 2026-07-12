<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Recent Commits page — full-page view aligned to design/pages/recent-commits.pen -->
<!-- Matches design/pages/recent-commits.pen frame rcPage -->
<template>
    <div class="rc-page" role="main" aria-label="Recent commits page">
        <div class="rc-title-col">
            <div class="rc-breadcrumbs">
                <span class="rc-crumb">Workspace</span>
                <ChevronRight :size="12" class="rc-crumb-sep" />
                <span class="rc-crumb">Commits</span>
            </div>
            <h1 class="rc-page-title">Commits</h1>
        </div>

        <RecentCommits class="rc-recent-section" />
    </div>
</template>

<script setup lang="ts">
import RecentCommits from "@/components/organisms/RecentCommits.vue";
import { ChevronRight } from "lucide-vue-next";
</script>

<style scoped>
.rc-page {
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.rc-title-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.rc-breadcrumbs {
    display: flex;
    align-items: center;
    gap: 8px;
}

.rc-crumb {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 500;
    color: var(--muted-foreground);
}

.rc-crumb-sep {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.rc-page-title {
    margin: 0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 24px;
    font-weight: 600;
    color: var(--foreground);
}

.rc-recent-section {
    width: 100%;
}
</style>
