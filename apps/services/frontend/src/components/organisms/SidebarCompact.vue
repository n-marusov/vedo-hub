<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Remaining Organism components — SidebarCompact, GroupSidebar, CollapsedGroupSidebar, SPARQLQueryEditor, MetricsDashboard, OntologyMetadata, MembersPanel, CommitHistory, DiffView, BranchList, TagList, RepositoryGraph, CommentsThread, SHACLRuleBuilder, ValidationReport -->

<!-- SidebarCompact.vue -->
<template>
  <nav class="sidebar-compact" :aria-label="'Compact navigation'">
    <ul role="list">
      <li v-for="item in items" :key="item.route">
        <router-link :to="item.route" :aria-label="item.label" class="sidebar-compact__link">
          <span aria-hidden="true">{{ item.icon }}</span>
        </router-link>
      </li>
    </ul>
  </nav>
</template>
<script setup lang="ts">
defineProps<{ items: Array<{ route: string; icon: string; label: string }> }>()
</script>
<style scoped>
.sidebar-compact { width: var(--sidebar-compact-width); height: 100%; background: var(--surface-primary); border-right: 1px solid var(--border-default); }
.sidebar-compact__link { display: flex; align-items: center; justify-content: center; width: 100%; height: 48px; font-size: var(--font-size-xl); color: var(--text-secondary); text-decoration: none; }
.sidebar-compact__link:hover { background: var(--surface-secondary); }
</style>
