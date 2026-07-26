<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism/Deployments — list of deployment cards with show stopped checkbox and delete -->
<!-- @m4 — Wired: accepts deployments as prop from DeploymentsPage -->
<template>
    <div class="deployments" role="region" aria-label="Deployments">
        <label class="dp-show-stopped">
            <input type="checkbox" v-model="showStopped" class="dp-checkbox" />
            <span class="dp-show-stopped-text">Show stopped deployments</span>
        </label>

        <div v-if="filteredDeployments.length === 0" class="dp-empty">
            <p>No deployments to display.</p>
        </div>

        <div v-else class="dp-list">
            <DeploymentCard
                v-for="dep in filteredDeployments"
                :key="dep.id"
                :status="dep.status === 'active' ? 'Active' : 'Stopped'"
                :url="dep.url"
                :classes="dep.ontologyName + ' ontology'"
                :individuals="'v' + (dep.version || '—')"
                :created="'Deployed ' + formatDate(dep.deployedAt)"
                :updated="'By ' + (dep.deployedBy || 'unknown')"
                :expiry="dep.status === 'stopped' ? 'Expired' : 'Active'"
                :stopped="dep.status === 'stopped'"
                @delete="$emit('delete', dep.id)"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import DeploymentCard from "./DeploymentCard.vue";

// @m4 — Deployment entry from LIST_DEPLOYMENTS_QUERY
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

const props = defineProps<{
	deployments: DeploymentEntry[];
}>();

defineEmits<{
	delete: [deploymentId: string];
}>();

const showStopped = ref(true);

const filteredDeployments = computed(() => {
	if (showStopped.value) return props.deployments;
	return props.deployments.filter((d) => d.status !== "stopped");
});

function formatDate(dateStr: string): string {
	if (!dateStr) return "unknown";
	const date = new Date(dateStr);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
	if (diffDays === 0) return "today";
	if (diffDays === 1) return "yesterday";
	if (diffDays < 30) return `${diffDays} days ago`;
	if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
	return `${Math.floor(diffDays / 365)} years ago`;
}
</script>

<style scoped>
.deployments {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 0;
}

.dp-show-stopped {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 0;
    cursor: pointer;
}

.dp-checkbox {
    width: 14px;
    height: 14px;
    accent-color: var(--primary);
}

.dp-show-stopped-text {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
}

.dp-list {
    display: flex;
    flex-direction: column;
    gap: 0;
}

/* @m4 Empty state */
.dp-empty {
    padding: 32px;
    text-align: center;
    font-family: 'IBM Plex Mono', monospace;
    font-size: 13px;
    color: var(--muted-foreground);
}
</style>
