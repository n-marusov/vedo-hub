<!-- @ctx: Members page strictly mirrored from design/frontend.pen frame memPan -->
<template>
  <div class="members-page" role="main" :aria-label="t('members.page_label')">
    <section class="members-title-row">
      <div class="members-title-wrap">
        <Users :size="20" class="primary" />
        <h1 class="members-title">{{ t('members.title') }}</h1>
      </div>
      <span class="members-count">{{ members.length }}</span>
    </section>

    <!-- Loading state -->
    <div v-if="loading" class="members-loading">
      <div class="skeleton skeleton--row" v-for="n in 3" :key="n"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="members-error" role="alert">
      <span>{{ t('members.load_error') }}</span>
      <button class="retry-btn" type="button" @click="fetchMembers">{{ t('common.retry') }}</button>
    </div>

    <!-- Empty state -->
    <div v-else-if="members.length === 0" class="members-empty">
      <span>{{ t('members.no_members') }}</span>
    </div>

    <!-- Data state -->
    <template v-else>
      <section class="members-context">
        <Folder :size="14" class="muted" />
        <span class="context-label">{{ t('members.managing_access_for') }}</span>
        <span class="context-badge context-badge--primary">ProductOntology</span>
        <Shield :size="14" class="muted" />
        <span class="context-note">{{ t('members.only_owners_note') }}</span>
      </section>

      <section class="members-card">
        <div class="table-head">
          <span class="member-col">{{ t('members.col_member') }}</span>
          <span class="role-col">{{ t('members.col_role') }}</span>
          <span class="mail-col">{{ t('members.col_email') }}</span>
          <span class="actions-col">{{ t('members.col_actions') }}</span>
        </div>
        <div class="sep"></div>
        <div v-for="member in members" :key="member.name + member.mail" class="table-row">
          <span class="member-col member-name">{{ member.name }}</span>
          <span class="role-col">
            <!-- @m4 Inline role edit: select when editing, pill when not -->
            <select
              v-if="editingMemberName === member.name"
              :value="member.role"
              class="role-select"
              :aria-label="t('members.select_role')"
              @change="onRoleChange(member, ($event.target as HTMLSelectElement).value)"
              @blur="onRoleBlur(member)"
            >
              <option value="Owner">Owner</option>
              <option value="Editor">Editor</option>
              <option value="Viewer">Viewer</option>
            </select>
            <span v-else class="role-pill">{{ member.role }}</span>
          </span>
          <span class="mail-col muted">{{ member.mail }}</span>
          <span class="actions-col row-actions">
            <button
              class="icon-btn"
              type="button"
              :aria-label="t('members.edit_member')"
              @click="startEdit(member)"
            ><Pencil :size="14" /></button>
            <button
              class="icon-btn danger"
              type="button"
              :aria-label="t('members.remove_member')"
              :disabled="isLastOwner(member.name)"
              :title="isLastOwner(member.name) ? t('members.cannot_remove_last_owner') : ''"
              @click="confirmRemove(member)"
            ><Trash2 :size="14" /></button>
          </span>
        </div>
      </section>
    </template>

    <!-- @m4 Remove confirmation dialog -->
    <Dialog
      :open="removeDialogOpen"
      :title="t('members.remove_member')"
      size="sm"
      @close="removeDialogOpen = false"
    >
      <p>{{ t('members.confirm_remove', { name: removingMemberName }) }}</p>
      <template #footer>
        <button class="dialog-cancel-btn" type="button" @click="removeDialogOpen = false">{{ t('common.cancel') }}</button>
        <button class="dialog-confirm-btn" type="button" @click="doRemoveMember">{{ t('common.confirm') }}</button>
      </template>
    </Dialog>

    <!-- @m4 Notification toast -->
    <div v-if="notifyMessage" class="members-notify" role="status">{{ notifyMessage }}</div>
  </div>
</template>

<script setup lang="ts">
import { type MemberInfo, listMembers } from "@/api/org";
import Dialog from "@/components/ui-kit/Dialog.vue";
import { useI18n } from "@/composables/useI18n";
import { Folder, Pencil, Shield, Trash2, Users } from "@lucide/vue";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();
const projectId = computed(() => String(route.params.id || ""));

interface MemberRow {
	name: string;
	role: string;
	mail: string;
}

const loading = ref(false);
const error = ref<string | null>(null);
const rawMembers = ref<MemberInfo[]>([]);

function mapMember(m: MemberInfo): MemberRow {
	return {
		name: String(m.username || m.userId || ""),
		role: String(m.role || "Viewer"),
		mail: `${String(m.username || "").toLowerCase()}@vedo.local`,
	};
}

const members = computed<MemberRow[]>(() => rawMembers.value.map(mapMember));

async function fetchMembers() {
	if (!projectId.value) return;
	loading.value = true;
	error.value = null;
	try {
		rawMembers.value = await listMembers(projectId.value);
	} catch (e: unknown) {
		error.value = e instanceof Error ? e.message : String(e);
	} finally {
		loading.value = false;
	}
}

onMounted(() => {
	fetchMembers();
});

watch(projectId, () => {
	fetchMembers();
});

// ── @m4 Inline Edit State ─────────────────────────────────────────────────────

const editingMemberName = ref<string | null>(null);
const notifyMessage = ref<string | null>(null);

function startEdit(member: MemberRow): void {
	editingMemberName.value = member.name;
	notifyMessage.value = null;
}

function onRoleChange(member: MemberRow, newRole: string): void {
	member.role = newRole;
	editingMemberName.value = null;
	notifyMessage.value = t("members.role_updated");
	setTimeout(() => {
		notifyMessage.value = null;
	}, 3000);
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "Members.edit",
			member: member.name,
			newRole,
			ts: new Date().toISOString(),
		}),
	);
}

function onRoleBlur(_member: MemberRow): void {
	editingMemberName.value = null;
}

// ── @m4 Remove Member ─────────────────────────────────────────────────────────

const removeDialogOpen = ref(false);
const removingMemberName = ref("");

function isLastOwner(name: string): boolean {
	const owners = members.value.filter((m) => m.role.toLowerCase() === "owner");
	return owners.length <= 1 && owners.some((m) => m.name === name);
}

function confirmRemove(member: MemberRow): void {
	if (isLastOwner(member.name)) {
		notifyMessage.value = t("members.cannot_remove_last_owner");
		setTimeout(() => {
			notifyMessage.value = null;
		}, 3000);
		return;
	}
	removingMemberName.value = member.name;
	removeDialogOpen.value = true;
}

function doRemoveMember(): void {
	const name = removingMemberName.value;
	const idx = rawMembers.value.findIndex(
		(m: MemberInfo) => mapMember(m).name === name,
	);
	if (idx !== -1) {
		rawMembers.value.splice(idx, 1);
	}
	removeDialogOpen.value = false;
	removingMemberName.value = "";
	notifyMessage.value = null;
}

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

/* @m4 Dialog buttons */
.dialog-cancel-btn {
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

.dialog-confirm-btn {
  height: 32px;
  border-radius: 6px;
  border: 1px solid var(--destructive);
  background: var(--destructive);
  color: white;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  padding: 0 12px;
  cursor: pointer;
}

/* @m4 Role select */
.role-select {
  height: 28px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  padding: 0 8px;
  cursor: pointer;
}

/* @m4 Notification toast */
.members-notify {
  position: fixed;
  bottom: 24px;
  right: 24px;
  background: var(--primary);
  color: var(--primary-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  padding: 12px 20px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  z-index: 100;
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
