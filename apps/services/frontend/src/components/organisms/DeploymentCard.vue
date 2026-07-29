<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Molecule/DeploymentCard — individual deployment entry with status, URL, classes/individuals, metadata, and delete action -->
<!-- Matches design/ui-kit.lib.pen mDeploymentCard (Molecule/DeploymentCard) -->
<template>
    <div
        class="deployment-card"
        role="listitem"
        :aria-label="`Deployment ${url}`"
    >
        <div class="dc-header">
            <span :class="['dc-status', { 'dc-status--stopped': stopped }]">{{
                status
            }}</span>
            <span class="dc-url">{{ url }}</span>
        </div>

        <div class="dc-divider"></div>

        <div class="dc-meta-row">
            <span class="dc-meta">{{ classes }}</span>
            <span class="dc-meta-dot">·</span>
            <span class="dc-meta">{{ individuals }}</span>
        </div>

        <div class="dc-footer">
            <span class="dc-meta">{{ created }}</span>
            <span class="dc-meta-dot">·</span>
            <span class="dc-meta">{{ updated }}</span>
            <span class="dc-meta-dot">·</span>
            <span class="dc-meta dc-meta--expiry">{{ expiry }}</span>
            <span class="dc-footer-spacer"></span>
            <button
                class="dc-delete-btn"
                type="button"
                :aria-label="`Delete deployment ${url}`"
                @click="$emit('delete')"
            >
                <Trash2 :size="14" />
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { Trash2 } from "@lucide/vue";

defineProps<{
	status: string;
	url: string;
	classes: string;
	individuals: string;
	created: string;
	updated: string;
	expiry: string;
	stopped?: boolean;
}>();

defineEmits<{
	delete: [];
}>();
</script>

<style scoped>
.deployment-card {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    transition: background var(--transition-fast, 0.15s ease);
}

.deployment-card:hover {
    background: var(--muted);
}

.dc-header {
    display: flex;
    align-items: center;
    gap: 8px;
}

.dc-status {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    color: var(--status-done);
}

.dc-status--stopped {
    color: var(--muted-foreground);
}

.dc-url {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--primary);
}

.dc-divider {
    width: 100%;
    height: 1px;
    background: var(--border);
}

.dc-meta-row {
    display: flex;
    align-items: center;
    gap: 4px;
}

.dc-footer {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-wrap: wrap;
}

.dc-footer-spacer {
    flex: 1;
    min-width: 0;
}

.dc-meta {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
}

.dc-meta-dot {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
}

.dc-delete-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border: none;
    background: transparent;
    color: var(--muted-foreground);
    cursor: pointer;
    border-radius: 4px;
    transition:
        color 0.15s,
        background 0.15s;
    flex-shrink: 0;
}

.dc-delete-btn:hover {
    color: var(--destructive, #ef4444);
    background: var(--muted);
}
</style>
