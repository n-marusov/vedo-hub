<!-- CommitHistory.vue -->
<template>
  <div class="commit-history" role="region" :aria-label="'Commit history'">
    <table class="commit-history__table">
      <thead>
        <tr>
          <th scope="col">Author</th>
          <th scope="col">Message</th>
          <th scope="col">SHA</th>
          <th scope="col">Date</th>
          <th scope="col">Branch</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="commit in commits" :key="commit.sha">
          <td><Avatar :display-name="commit.author" size="sm" /></td>
          <td>{{ commit.message }}</td>
          <td><code class="commit-history__sha">{{ commit.sha.slice(0, 7) }}</code></td>
          <td>{{ formatDate(commit.date) }}</td>
          <td><Badge :text="commit.branch" variant="default" /></td>
        </tr>
      </tbody>
    </table>
    <div v-if="!commits.length" class="commit-history__empty">
      No commits yet. Start by creating a class.
    </div>
  </div>
</template>

<script setup lang="ts">
import Avatar from '../ui-kit/Avatar.vue'
import Badge from '../ui-kit/Badge.vue'

defineProps<{
  commits: Array<{ author: string; message: string; sha: string; date: string; branch: string }>
}>()

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString()
}
</script>

<style scoped>
.commit-history { padding: var(--spacing-4); }
.commit-history__table { font-size: var(--font-size-sm); }
.commit-history__table th, .commit-history__table td { padding: var(--spacing-2); border-bottom: 1px solid var(--border-default); text-align: left; }
.commit-history__sha { font-family: var(--font-family-mono); font-size: var(--font-size-xs); background: var(--surface-secondary); padding: var(--spacing-1) var(--spacing-2); border-radius: var(--radius-sm); }
.commit-history__empty { padding: var(--spacing-8); text-align: center; color: var(--text-muted); }
</style>
