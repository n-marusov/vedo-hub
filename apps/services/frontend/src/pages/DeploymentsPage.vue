<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Deployments page — aligned to design/pages/deployments.pen -->
<!-- @m4 — Wired to LIST_DEPLOYMENTS_QUERY via Apollo GraphQL -->
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

        <!-- @m4 Loading state -->
        <div v-if="loading" class="dp-section">
            <div class="skeleton-card" v-for="n in 3" :key="n">
                <div class="skeleton-line skeleton-line--wide"></div>
                <div class="skeleton-line skeleton-line--med"></div>
            </div>
        </div>

        <!-- @m4 Error state -->
        <div v-else-if="error" class="dp-section">
            <div class="error-state">
                <p>Failed to load deployments.</p>
                <button class="retry-btn" type="button" @click="fetchDeployments()">Retry</button>
            </div>
        </div>

        <!-- @m4 Data state — pass API deployments to organism -->
        <Deployments v-else class="dp-section" :deployments="resolvedDeployments" @delete="handleDelete" />
    </div>
</template>

<script setup lang="ts">
import { type DeploymentInfo, listDeployments } from "@/api/deployments";
import Deployments from "@/components/organisms/Deployments.vue";
import { ChevronRight } from "@lucide/vue";
import { computed, onMounted, ref } from "vue";

// @m4 — Wire deployments via REST client
const loading = ref(false);
const error = ref<string | null>(null);
const depsData = ref<DeploymentInfo[]>([]);

async function fetchDeployments() {
	loading.value = true;
	error.value = null;
	try {
		depsData.value = await listDeployments(true);
	} catch (e: unknown) {
		error.value = e instanceof Error ? e.message : "Failed to load deployments";
	} finally {
		loading.value = false;
	}
}

onMounted(fetchDeployments);

interface DeploymentEntry {
	id: string;
	status: string;
	url: string;
	version: string;
	ontologyId: string;
	ontologyName: string;
	deployedAt: string;
	deployedBy: string;
	stopped?: boolean;
}

const resolvedDeployments = computed<DeploymentEntry[]>(() => {
	return depsData.value.map((d) => ({
		...d,
		stopped: d.status === "stopped",
	}));
});

function handleDelete(deploymentId: string): void {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "deployments.delete",
			deploymentId,
			ts: new Date().toISOString(),
		}),
	);
	// In a full implementation, this would call a DELETE endpoint
	// For now, refetch the list to reflect the deletion
	fetchDeployments();
}
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
