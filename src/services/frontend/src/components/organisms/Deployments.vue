<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism/Deployments — list of deployment cards with show stopped checkbox and delete -->
<!-- Matches design/ui-kit.lib.pen orgDeployments (Organism/Deployments) -->
<template>
    <div class="deployments" role="region" aria-label="Deployments">
        <label class="dp-show-stopped">
            <input type="checkbox" v-model="showStopped" class="dp-checkbox" />
            <span class="dp-show-stopped-text">Show stopped deployments</span>
        </label>

        <div class="dp-list">
            <DeploymentCard
                v-for="dep in filteredDeployments"
                :key="dep.url"
                :status="dep.status"
                :url="dep.url"
                :classes="dep.classes"
                :individuals="dep.individuals"
                :created="dep.created"
                :updated="dep.updated"
                :expiry="dep.expiry"
                :stopped="dep.stopped"
                @delete="
                    deployments = deployments.filter((d) => d.url !== dep.url)
                "
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import DeploymentCard from "./DeploymentCard.vue";

interface DeploymentEntry {
    status: string;
    url: string;
    classes: string;
    individuals: string;
    created: string;
    updated: string;
    expiry: string;
    stopped?: boolean;
}

const showStopped = ref(true);

const deployments = ref<DeploymentEntry[]>([
    {
        status: "Active",
        url: "https://philosophy-aa4ded.gitlab.io",
        classes: "12 classes",
        individuals: "58 individuals",
        created: "Created 5 months ago",
        updated: "Last updated 5 months ago",
        expiry: "Never expires",
    },
    {
        status: "Active",
        url: "https://vedo-core-qa.gitlab.io",
        classes: "35 classes",
        individuals: "142 individuals",
        created: "Created 2 days ago",
        updated: "Last updated 2 days ago",
        expiry: "Expires in 28 days",
    },
    {
        status: "Active",
        url: "https://ontology-staging.gitlab.io",
        classes: "8 classes",
        individuals: "23 individuals",
        created: "Created 1 week ago",
        updated: "Last updated 6 days ago",
        expiry: "Expires in 21 days",
    },
    {
        status: "Active",
        url: "https://shacl-validator.gitlab.io",
        classes: "42 classes",
        individuals: "197 individuals",
        created: "Created 3 weeks ago",
        updated: "Last updated 3 weeks ago",
        expiry: "Expires in 7 days",
    },
    {
        status: "Stopped",
        url: "https://legacy-ontology.gitlab.io",
        classes: "18 classes",
        individuals: "76 individuals",
        created: "Created 1 year ago",
        updated: "Last updated 8 months ago",
        expiry: "Expired",
        stopped: true,
    },
]);

const filteredDeployments = computed(() => {
    if (showStopped.value) return deployments.value;
    return deployments.value.filter((d) => !d.stopped);
});
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
</style>
