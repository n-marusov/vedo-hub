<!-- RepositoryGraph.vue -->
<template>
  <div class="repo-graph" role="img" :aria-label="'Repository DAG graph'">
    <div class="repo-graph__canvas">
      <div
        v-for="node in nodes"
        :key="node.sha"
        :class="['repo-graph__node', `repo-graph__node--${node.type}`]"
        :style="{ left: `${node.x}px`, top: `${node.y}px` }"
        :title="`${node.message} (${node.sha.slice(0, 7)})`"
      >
        <span class="repo-graph__dot"></span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  nodes: Array<{
    sha: string
    message: string
    type: 'commit' | 'branch' | 'tag'
    x: number
    y: number
  }>
}>()
</script>

<style scoped>
.repo-graph { height: 400px; position: relative; background: var(--surface-secondary); border-radius: var(--radius-lg); }
.repo-graph__canvas { position: relative; width: 100%; height: 100%; }
.repo-graph__node { position: absolute; display: flex; align-items: center; }
.repo-graph__dot { width: 12px; height: 12px; border-radius: 50%; border: 2px solid var(--primary); background: var(--surface-primary); }
.repo-graph__node--branch .repo-graph__dot { border-color: var(--status-success); background: var(--status-success-bg); }
.repo-graph__node--tag .repo-graph__dot { border-color: var(--accent); background: var(--primary-muted); }
</style>
