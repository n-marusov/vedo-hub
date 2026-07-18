<!-- @ctx: Members page strictly mirrored from design/frontend.pen frame memPan -->
<template>
  <div class="members-page" role="main" aria-label="Members content">
    <section class="members-title-row">
      <div class="members-title-wrap">
        <Users :size="20" class="primary" />
        <h1 class="members-title">Members</h1>
      </div>
      <span class="members-count">{{ members.length }}</span>
    </section>

    <!-- Loading state -->
    <div v-if="loading" class="members-loading">
      <div class="skeleton skeleton--row" v-for="n in 3" :key="n"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="members-error" role="alert">
      <span>Failed to load members</span>
      <button class="retry-btn" type="button" @click="refetch">Retry</button>
    </div>

    <!-- Empty state -->
    <div v-else-if="members.length === 0" class="members-empty">
      <span>No members found.</span>
    </div>

    <!-- Data state -->
    <template v-else>
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
    </template>
  </div>
</template>

<script setup lang="ts">
import { LIST_MEMBERS_QUERY } from "@/apollo/queries";
import { useQuery } from "@vue/apollo-composable";
import { Folder, Pencil, Shield, Trash2, Users } from "lucide-vue-next";
import { computed, watch } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const ontologyId = computed(() => String(route.params.id || ""));

interface MemberRow {
	name: string;
	role: string;
	mail: string;
}

const { result, loading, error, refetch } = useQuery(
	LIST_MEMBERS_QUERY,
	() => ({
		ontologyId: ontologyId.value,
	}),
	{
		fetchPolicy: "cache-and-network",
		enabled: computed(() => !!ontologyId.value),
	},
);

const members = computed<MemberRow[]>(() => {
	const items = result.value?.members;
	if (!items || items.length === 0) {
		return [];
	}
	return items.map((m: Record<string, unknown>) => ({
		name: String(m.username || m.userId || ""),
		role: String(m.role || "Viewer"),
		mail: `${String(m.username || "").toLowerCase()}@vedo.local`,
	}));
});

// ── Logging ─────────────────────────────────────────────────────────────────

watch(members, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "members.list.loaded",
			count: val.length,
			ts: new Date().toISOString(),
		}),
	);
});

watch(error, (err) => {
	if (err) {
		console.error(
			JSON.stringify({
				level: "error",
				msg: "members.query.error",
				error: String(err),
				ts: new Date().toISOString(),
			}),
		);
	}
});
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

/* Skeleton loading */
.members-loading {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton {
  background: var(--muted);
  border-radius: 4px;
}

.skeleton--row {
  height: 48px;
  width: 100%;
}

/* Error state */
.members-error,
.members-empty {
  padding: 32px;
  text-align: center;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card);
}

.retry-btn {
  margin-top: 12px;
  height: 32px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  padding: 0 12px;
  cursor: pointer;
}

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
