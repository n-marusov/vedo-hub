<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism/MergeRequests — collapsible section cards, no tabs -->
<!-- @m4 — Wired: accepts mergeRequests as prop from MergeRequestsPage -->
<template>
    <div class="merge-requests" role="region" aria-label="Merge requests">
        <MergeRequestCard
            v-for="section in sections"
            :key="section.title"
            :title="section.title"
            :open="section.open"
            :count="section.count"
            :empty-text="section.emptyText"
            @toggle="section.open = !section.open"
        />

        <p class="mr-excluded-hint">
            Items below are excluded from the active count
        </p>

        <MergeRequestCard
            v-for="section in secondarySections"
            :key="section.title"
            :title="section.title"
            :open="section.open"
            :count="0"
            :empty-text="section.emptyText"
            @toggle="section.open = !section.open"
        />
    </div>
</template>

<script setup lang="ts">
import type { MergeRequestInfo } from "@/api/merge-requests";
import { computed, reactive } from "vue";
import MergeRequestCard from "./MergeRequestCard.vue";

const props = defineProps<{
	mergeRequests: MergeRequestInfo[];
}>();

interface MRSection {
	title: string;
	open: boolean;
	count?: number;
	emptyText: string;
}

// @m4 — Compute section counts from API data
const openCount = computed(
	() => props.mergeRequests.filter((mr) => mr.status === "open").length,
);

const sections: MRSection[] = reactive([
	{
		title: "Returned to you",
		open: true,
		count: 0,
		emptyText: "No merge requests match this list.",
	},
	{
		title: "Review requested",
		open: true,
		count: openCount,
		emptyText: "No merge requests match this list.",
	},
	{
		title: "Your merge requests",
		open: true,
		count: 0,
		emptyText: "No merge requests match this list.",
	},
]);

const secondarySections: MRSection[] = reactive([
	{
		title: "Waiting for author or assignee",
		open: true,
		emptyText: "No merge requests match this list.",
	},
	{
		title: "Waiting for approvals",
		open: true,
		emptyText: "No merge requests match this list.",
	},
	{
		title: "Approved by you",
		open: true,
		emptyText: "No merge requests match this list.",
	},
	{
		title: "Approved by others",
		open: true,
		emptyText: "No merge requests match this list.",
	},
]);
</script>

<style scoped>
.merge-requests {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.mr-excluded-hint {
    margin: 8px 0 0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 500;
    color: var(--muted-foreground);
}
</style>
