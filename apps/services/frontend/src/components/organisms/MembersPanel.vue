<!-- MembersPanel.vue -->
<template>
  <div class="members-panel" role="region" :aria-label="'Members panel'">
    <div class="members-panel__header">
      <h3>Members</h3>
      <Badge :text="String(totalCount)" variant="info" label="Member count" />
    </div>
    <table class="members-panel__table">
      <thead>
        <tr>
          <th scope="col">Avatar</th>
          <th scope="col">Name</th>
          <th scope="col">Role</th>
          <th scope="col">Joined</th>
          <th scope="col">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="member in members" :key="member.user_id">
          <td><Avatar :src="member.avatar_url" :display-name="member.display_name" size="sm" /></td>
          <td>{{ member.display_name }}</td>
          <td>
            <Select
              v-if="member.can_edit"
              :model-value="member.role"
              :options="roleOptions"
              label="Role"
              @update:model-value="$emit('role-change', member.user_id, $event)"
            />
            <span v-else>{{ member.role }}</span>
          </td>
          <td>{{ formatDate(member.joined_at) }}</td>
          <td>
            <GhostButton v-if="member.can_edit" @click="$emit('remove', member.user_id)">Remove</GhostButton>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import Avatar from '../ui-kit/Avatar.vue'
import Badge from '../ui-kit/Badge.vue'
import GhostButton from '../ui-kit/GhostButton.vue'
import Select from '../ui-kit/Select.vue'

defineProps<{
  members: Array<{
    user_id: string
    avatar_url?: string
    display_name: string
    role: string
    joined_at: string
    can_edit: boolean
  }>
  totalCount: number
}>()

defineEmits<{
  'role-change': [userId: string, role: string]
  remove: [userId: string]
}>()

const roleOptions = [
  { value: 'Viewer', label: 'Viewer' },
  { value: 'Editor', label: 'Editor' },
  { value: 'Maintainer', label: 'Maintainer' },
  { value: 'Owner', label: 'Owner' }
]

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString()
}
</script>

<style scoped>
.members-panel { padding: var(--spacing-4); }
.members-panel__header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--spacing-4); }
.members-panel__table { font-size: var(--font-size-sm); }
.members-panel__table th, .members-panel__table td { padding: var(--spacing-2); border-bottom: 1px solid var(--border-default); text-align: left; vertical-align: middle; }
</style>
