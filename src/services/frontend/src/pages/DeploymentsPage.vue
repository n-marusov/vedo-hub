<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Deployments page — aligned to design/pages/deployments.pen -->
<!-- Matches design/pages/deployments.pen frame deployPage -->
<template>
    <div class="dp-page" role="main" aria-label="Deployments page">
        <div class="dp-title-col">
            <div class="dp-breadcrumbs">
                <span class="dp-crumb">Workspace</span>
                <ChevronRight :size="12" class="dp-crumb-sep" />
                <span class="dp-crumb">Deployments</span>
            </div>
            <h1 class="dp-page-title">Deployments</h1>
        </div>

        <Deployments class="dp-section" />
    </div>
</template>

<script setup lang="ts">
import Deployments from "@/components/organisms/Deployments.vue";
import { ChevronRight } from "lucide-vue-next";
</script>

<style scoped>
.dp-page {
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.dp-title-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.dp-breadcrumbs {
    display: flex;
    align-items: center;
    gap: 8px;
}

.dp-crumb {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 500;
    color: var(--muted-foreground);
}

.dp-crumb-sep {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.dp-page-title {
    margin: 0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 24px;
    font-weight: 600;
    color: var(--foreground);
}

.dp-section {
    width: 100%;
}
</style>
