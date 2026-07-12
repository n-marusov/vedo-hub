<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit Tabs — individual pill-style tabs matching Component/Tabs in design -->
<template>
    <div class="tabs" role="tablist" :aria-label="label">
        <button
            v-for="tab in tabs"
            :key="tab.value"
            role="tab"
            :class="['tab', { 'tab--active': modelValue === tab.value }]"
            :aria-selected="modelValue === tab.value"
            @click="$emit('update:modelValue', tab.value)"
        >
            {{ tab.label }}
        </button>
    </div>
</template>

<script setup lang="ts">
defineProps<{
    modelValue?: string;
    tabs: Array<{ value: string; label: string }>;
    label?: string;
}>();

defineEmits<{
    "update:modelValue": [value: string];
}>();
</script>

<style scoped>
.tabs {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
}

.tab {
    padding: 4px 8px;
    font-family: "IBM Plex Mono", monospace;
    font-size: 10px;
    font-weight: 500;
    color: var(--muted-foreground);
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 0;
    cursor: pointer;
    transition:
        background 0.15s,
        color 0.15s;
}

.tab:hover {
    background: var(--muted);
}

.tab--active {
    background: var(--secondary);
    color: var(--primary);
    font-weight: 500;
}
</style>
