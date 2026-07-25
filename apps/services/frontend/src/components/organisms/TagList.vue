<!-- TagList.vue -->
<template>
  <div class="tag-list" role="region" :aria-label="'Tag list'">
    <table class="tag-list__table">
      <thead>
        <tr>
          <th scope="col">Tag</th>
          <th scope="col">Commit</th>
          <th scope="col">Description</th>
          <th scope="col">Updated</th>
          <th scope="col">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="tag in tags" :key="tag.name">
          <td><Badge :text="tag.name" variant="info" /></td>
          <td><code>{{ tag.commit.slice(0, 7) }}</code></td>
          <td>{{ tag.description }}</td>
          <td>{{ formatDate(tag.updated) }}</td>
          <td><GhostButton @click="$emit('delete', tag.name)">Delete</GhostButton></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import Badge from '../ui-kit/Badge.vue'
import GhostButton from '../ui-kit/GhostButton.vue'

defineProps<{
  tags: Array<{ name: string; commit: string; description: string; updated: string }>
}>()
defineEmits<{ delete: [name: string] }>()

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString()
}
</script>

<style scoped>
.tag-list { padding: var(--spacing-4); }
.tag-list__table { font-size: var(--font-size-sm); }
.tag-list__table th, .tag-list__table td { padding: var(--spacing-2); border-bottom: 1px solid var(--border-default); text-align: left; }
</style>
