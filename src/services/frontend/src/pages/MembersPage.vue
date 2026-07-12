<!-- @ctx: Members page strictly mirrored from design/frontend.pen frame memPan -->
<template>
  <div class="members-page" role="main" aria-label="Members content">
    <section class="members-title-row">
      <div class="members-title-wrap">
        <Users :size="20" class="primary" />
        <h1 class="members-title">Members</h1>
      </div>
      <span class="members-count">4</span>
    </section>

    <section class="members-context">
      <Folder :size="14" class="muted" />
      <span class="context-label">Managing access for:</span>
      <span class="context-badge context-badge--primary">ProductOntology</span>
      <Shield :size="14" class="muted" />
      <span class="context-note">Only owners can manage members</span>
    </section>

    <section class="members-card">
      <div class="table-head">
        <span class="member-col">Member</span>
        <span class="role-col">Role</span>
        <span class="mail-col">Email</span>
        <span class="actions-col">Actions</span>
      </div>
      <div class="sep"></div>
      <div v-for="member in members" :key="member.name + member.mail" class="table-row">
        <span class="member-col member-name">{{ member.name }}</span>
        <span class="role-col"><span class="role-pill">{{ member.role }}</span></span>
        <span class="mail-col muted">{{ member.mail }}</span>
        <span class="actions-col row-actions">
          <button class="icon-btn" type="button" aria-label="Edit member"><Pencil :size="14" /></button>
          <button class="icon-btn danger" type="button" aria-label="Remove member"><Trash2 :size="14" /></button>
        </span>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Folder, Pencil, Shield, Trash2, Users } from 'lucide-vue-next'

const members = [
  { name: 'Nikolay Marusov', role: 'Owner', mail: 'nikolay@vedo.local' },
  { name: 'Alice Smith', role: 'Maintainer', mail: 'alice@vedo.local' },
  { name: 'Bob Johnson', role: 'Editor', mail: 'bob@vedo.local' },
  { name: 'Anna Petrova', role: 'Viewer', mail: 'anna@vedo.local' }
]
</script>

<style scoped>
.members-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.members-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
}

.members-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.members-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

.members-count {
  min-width: 22px;
  height: 22px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.members-context {
  border-radius: 6px;
  background: rgba(20, 20, 20, 0.3);
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.context-label,
.context-note {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.context-note { color: var(--info); font-size: 11px; }

.context-badge {
  border-radius: 999px;
  border: 1px solid var(--border);
  background: #000;
  padding: 2px 8px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.context-badge--primary {
  border-color: var(--primary);
  color: var(--primary);
}

.members-card {
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card);
  padding: 16px;
}

.table-head,
.table-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 8px;
  font-family: 'IBM Plex Mono', monospace;
}

.table-head {
  color: var(--muted-foreground);
  font-size: 11px;
  font-weight: 600;
}

.table-row {
  border-radius: 6px;
  font-size: 13px;
}

.table-row:hover { background: rgba(20, 20, 20, 0.25); }

.member-col { flex: 1.2; }
.role-col { width: 180px; }
.mail-col { flex: 1; }
.actions-col { width: 140px; }

.member-name { color: var(--foreground); font-weight: 500; }

.role-pill {
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--muted);
  padding: 2px 8px;
  font-size: 11px;
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.icon-btn {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  color: var(--muted-foreground);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.icon-btn.danger { color: var(--destructive); }

.sep { width: 100%; height: 1px; background: var(--border); }

.primary { color: var(--primary); }
.muted { color: var(--muted-foreground); }

@media (max-width: 1024px) {
  .table-head { display: none; }
  .table-row {
    display: grid;
    grid-template-columns: 1fr;
    gap: 8px;
    padding: 12px;
  }
  .member-col,
  .role-col,
  .mail-col,
  .actions-col { width: 100%; }
}

@media (max-width: 768px) {
  .members-page { padding: 16px; }
}
</style>
