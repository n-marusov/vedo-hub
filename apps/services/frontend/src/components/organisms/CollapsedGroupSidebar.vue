<!-- CollapsedGroupSidebar.vue -->
<template>
  <aside class="collapsed-sidebar" :aria-label="'Collapsed group sidebar'">
    <Avatar v-for="m in members.slice(0, 3)" :key="m.id" :display-name="m.name" size="sm" />
    <span v-if="members.length > 3" class="collapsed-sidebar__more">+{{ members.length - 3 }}</span>
  </aside>
</template>
<script setup lang="ts">
import Avatar from '../ui-kit/Avatar.vue'
defineProps<{ members: Array<{ id: string; name: string }> }>()
</script>
<style scoped>
.collapsed-sidebar { display: flex; flex-direction: column; align-items: center; gap: var(--spacing-2); padding: var(--spacing-4); background: var(--surface-primary); border-right: 1px solid var(--border-default); }
.collapsed-sidebar__more { font-size: var(--font-size-xs); color: var(--text-muted); }
</style>
