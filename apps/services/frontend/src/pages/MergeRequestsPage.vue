<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Merge Requests page — aligned to design/pages/merge-requests.pen -->
<!-- @m4 — Wired to LIST_MERGE_REQUESTS_QUERY via Apollo GraphQL, filtered by tab -->
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

        <!-- @m4 Loading state -->
        <div v-if="loading" class="mr-section">
            <div class="skeleton-card" v-for="n in 3" :key="n">
                <div class="skeleton-line skeleton-line--wide"></div>
                <div class="skeleton-line skeleton-line--med"></div>
            </div>
        </div>

        <!-- @m4 Error state -->
        <div v-else-if="error" class="mr-section">
            <div class="error-state">
                <p>Failed to load merge requests.</p>
                <button class="retry-btn" type="button" @click="fetchMRs">Retry</button>
            </div>
        </div>

        <!-- @m4 Data state — pass filtered MRs to organism -->
        <MergeRequests v-else class="mr-section" :mergeRequests="filteredMRs" />
    </div>
</template>

<script setup lang="ts">
import { listMergeRequests } from "@/api/merge-requests";
import MergeRequests from "@/components/organisms/MergeRequests.vue";
import Tab from "@/components/ui-kit/Tab.vue";
import { ChevronDown, ChevronRight } from "lucide-vue-next";
import { computed, onMounted, ref, watch } from "vue";

const tabs = [
	{ value: "active", label: "Active" },
	{ value: "merged", label: "Merged" },
	{ value: "all", label: "Search" },
];

const activeTab = ref("active");

const loading = ref(false);
const error = ref<string | null>(null);
const mrData = ref<any[]>([]);

async function fetchMRs() {
	loading.value = true;
	error.value = null;
	try {
		mrData.value = await listMergeRequests(
			activeTab.value === "all" ? undefined : activeTab.value,
		);
	} catch (e: any) {
		error.value = e.message ?? String(e);
	} finally {
		loading.value = false;
	}
}

onMounted(() => { fetchMRs(); });

watch(activeTab, () => { fetchMRs(); });

interface MREntry {
	id: string;
	title: string;
	description: string;
	sourceBranch: string;
	targetBranch: string;
	authorName: string;
	status: string;
	mergeStatus: string;
	createdAt: string;
	commentCount: number;
}

const filteredMRs = computed<MREntry[]>(() => {
	const mrs = mrData.value;
	if (!mrs) return [];
	if (activeTab.value === "all") return mrs;
	return mrs.filter((mr: MREntry) => mr.status === activeTab.value);
});

// ── Logging ─────────────────────────────────────────────────────────────────

watch(filteredMRs, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "merge_requests.list.loaded",
			count: val.length,
			ts: new Date().toISOString(),
		}),
	);
});

watch(error, (err) => {
	if (err) {
		console.error(
			JSON.stringify({
				level: "error",
				msg: "merge_requests.query.error",
				error: String(err),
				ts: new Date().toISOString(),
			}),
		);
	}
});
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

/* @m4 Skeleton loading */
.skeleton-card {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    gap: 8px;
    animation: pulse 1.5s ease-in-out infinite;
}
.skeleton-line {
    height: 12px;
    border-radius: 4px;
    background: var(--muted);
}
.skeleton-line--wide { width: 70%; }
.skeleton-line--med { width: 40%; }
@keyframes pulse {
    0%, 100% { opacity: 0.4; }
    50% { opacity: 0.8; }
}

/* @m4 Error state */
.error-state {
    padding: 32px;
    text-align: center;
    font-family: 'IBM Plex Mono', monospace;
    font-size: 14px;
    color: var(--muted-foreground);
}
.retry-btn {
    margin-top: 12px;
    height: 32px;
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 0 12px;
    font-family: 'IBM Plex Mono', monospace;
    font-size: 12px;
    background: var(--card);
    color: var(--foreground);
    cursor: pointer;
}
.retry-btn:hover { background: var(--muted); }
</style>
