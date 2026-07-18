<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Merge Requests page — aligned to design/pages/merge-requests.pen -->
<!-- Matches design/pages/merge-requests.pen frame G83Yfe -->
<template>
    <div class="mr-page" role="main" aria-label="Merge requests">
        <div class="mr-title-row">
            <div class="mr-title-left">
                <div class="mr-breadcrumbs">
                    <span class="mr-crumb">Workspace</span>
                    <ChevronRight :size="12" class="mr-crumb-sep" />
                    <span class="mr-crumb">Merge request</span>
                </div>
                <h1 class="mr-title">Merge Requests</h1>
            </div>
            <button class="mr-project-select" type="button">
                <span>Select project to create merge request</span>
                <ChevronDown :size="12" class="mr-select-chevron" />
            </button>
        </div>

        <Tab v-model="activeTab" :tabs="tabs" label="Merge request filter" />

        <MergeRequests class="mr-section" />
    </div>
</template>

<script setup lang="ts">
import MergeRequests from "@/components/organisms/MergeRequests.vue";
import Tab from "@/components/ui-kit/Tab.vue";
import { ChevronDown, ChevronRight } from "lucide-vue-next";
import { ref } from "vue";

const tabs = [
	{ value: "active", label: "Active" },
	{ value: "merged", label: "Merged" },
	{ value: "all", label: "Search" },
];

const activeTab = ref("active");
</script>

<style scoped>
.mr-page {
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.mr-title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.mr-title-left {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.mr-breadcrumbs {
    display: flex;
    align-items: center;
    gap: 8px;
}

.mr-crumb {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 500;
    color: var(--muted-foreground);
}

.mr-crumb-sep {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.mr-title {
    margin: 0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 24px;
    font-weight: 600;
    color: var(--foreground);
}

.mr-project-select {
    height: 36px;
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 0 12px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: "IBM Plex Mono", monospace;
    font-size: 13px;
    color: var(--muted-foreground);
    background: transparent;
}

.mr-select-chevron {
    flex-shrink: 0;
    color: var(--muted-foreground);
}

.mr-section {
    width: 100%;
}
</style>
