<template>
  <div class="groups-page" role="main" aria-label="Groups page">
    <section class="gp-top">
      <div class="gp-title-col">
        <div class="gp-breadcrumbs">
          <span class="gp-breadcrumb-text">Workspace</span>
          <ChevronRight :size="12" class="gp-breadcrumb-sep" />
          <span class="gp-breadcrumb-text">Groups</span>
        </div>
        <h1 class="gp-title">Groups</h1>
      </div>
      <button class="gp-new-btn" type="button" @click="showCreateDialog = true"><Plus :size="14" />New group</button>
    </section>

    <section class="gp-toolbar">
      <div class="gp-search-wrap">
        <Search :size="14" class="gp-search-icon" />
        <input class="gp-search-input" type="text" v-model="searchQuery" placeholder="Search groups" aria-label="Search groups" />
      </div>
      <div class="gp-sort-wrap">
        <span class="gp-sort-label">Name</span>
        <ChevronDown :size="12" class="gp-sort-chevron" />
        <span class="gp-sort-divider"></span>
        <span class="gp-sort-label">Ascending</span>
        <ChevronDown :size="12" class="gp-sort-chevron" />
      </div>
    </section>

    <!-- Loading state -->
    <div v-if="loading" class="gp-list">
      <div v-for="n in 3" :key="n" class="gp-row gp-skeleton-row">
        <div class="gp-row-body">
          <div class="gp-row-body-top">
            <div class="skeleton skeleton--circle"></div>
            <div class="skeleton skeleton--text skeleton--name"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="gp-error" role="alert">
      <span>Failed to load groups</span>
      <button class="retry-btn" type="button" @click="fetchGroups">Retry</button>
    </div>

    <!-- Empty state -->
    <div v-else-if="groupRows.length === 0" class="gp-empty">
      <span>No groups found.</span>
    </div>

    <!-- Data state -->
    <section v-else class="gp-list">
      <div v-for="(row, i) in groupRows" :key="row.name + i" :class="['gp-row', { 'group-child-row': row.isChild } ]">
        <div class="gp-row-body">
          <div class="gp-row-body-top">
            <div class="gp-row-indent" :style="{ width: row.indent + 'px' }"></div>
            <component
              v-if="row.type === 'group'"
              :is="row.chevronIcon"
              :size="12"
              class="gp-row-chevron"
              @click="toggleExpand(row.name)"
            />
            <FolderTree v-if="row.type === 'group'" :size="20" class="gp-row-folder-icon" />
            <Folder v-else :size="20" class="gp-row-folder-icon" />
            <span class="gp-row-logo" :style="{ background: row.logoBg }">{{ row.logoLetter }}</span>
            <span :class="['gp-row-name', { 'gp-row-name-active': row.active, 'gp-row-name--project': row.type === 'project' }]">{{ row.name }}</span>
            <Globe v-if="row.visibility === 'public'" :size="12" class="gp-row-vis-icon" />
            <Lock v-if="row.visibility === 'private'" :size="12" class="gp-row-vis-icon" />
          </div>
          <span
            class="gp-row-desc"
            :style="{ paddingLeft: (row.indent + (row.type === 'group' ? 86 : 68)) + 'px' }"
          >{{ row.description }}</span>
        </div>
        <div class="gp-row-actions">
          <div class="gp-row-counters">
            <div v-if="row.type === 'group'" class="gp-counter">
              <FolderTree :size="14" /><span>{{ row.subgroups }}</span>
            </div>
            <div v-if="row.type === 'group'" class="gp-counter">
              <Folder :size="14" /><span>{{ row.projects }}</span>
            </div>
            <div v-if="row.type === 'group'" class="gp-counter">
              <Users :size="14" /><span>{{ row.members }}</span>
            </div>
            <div v-if="row.type === 'project'" class="gp-counter">
              <Star :size="14" /><span>{{ row.stars }}</span>
            </div>
          </div>
          <span class="gp-row-created">{{ row.created }}</span>
        </div>
        <div class="gp-row-menu-wrap">
          <MoreVertical :size="16" class="gp-row-menu" />
        </div>
      </div>
    </section>

    <CreateGroupDialog
      :open="showCreateDialog"
      @close="showCreateDialog = false"
      @created="onGroupCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { listGroups } from "@/api/org";
import CreateGroupDialog from "@/components/groups/CreateGroupDialog.vue";
import {
	ChevronDown,
	ChevronRight,
	Folder,
	FolderTree,
	Globe,
	Lock,
	MoreVertical,
	Plus,
	Search,
	Star,
	Users,
} from "lucide-vue-next";
import type { Component } from "vue";
import { computed, onMounted, reactive, ref, watch } from "vue";

const searchQuery = ref("");
const showCreateDialog = ref(false);

// @m4 Reactive expand/collapse map — keyed by group name
type ExpandedMap = Record<string, boolean>;
const expanded = reactive<ExpandedMap>({});

function toggleExpand(name: string): void {
	expanded[name] = !expanded[name];
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "Groups.expand",
			group: name,
			expanded: expanded[name],
			ts: new Date().toISOString(),
		}),
	);
}

type RowType = "group" | "project";

interface GroupRow {
	name: string;
	indent: number;
	chevronIcon: Component;
	logoLetter: string;
	logoBg: string;
	visibility: "public" | "private";
	description: string;
	type: RowType;
	subgroups?: number;
	projects?: number;
	members?: number;
	stars?: number;
	created: string;
	active: boolean;
	isChild: boolean;
}

const loading = ref(false);
const error = ref<string | null>(null);
const groupsData = ref<any[]>([]);

async function fetchGroups() {
	loading.value = true;
	error.value = null;
	try {
		groupsData.value = await listGroups(searchQuery.value || undefined);
	} catch (e: any) {
		error.value = e.message ?? String(e);
	} finally {
		loading.value = false;
	}
}

function onGroupCreated(name: string): void {
	showCreateDialog.value = false;
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "Groups.groupCreated",
			groupName: name,
			ts: new Date().toISOString(),
		}),
	);
	fetchGroups();
}

onMounted(() => {
	fetchGroups();
});

const groupRows = computed<GroupRow[]>(() => {
	const items = groupsData.value;
	if (!items || items.length === 0) {
		return [];
	}
	// Build flat hierarchy from nested API response
	const rows: GroupRow[] = [];
	function walk(
		group: Record<string, unknown>,
		indent: number,
		isChild: boolean,
	): void {
		const groupName = String(group.name || "");
		rows.push({
			name: groupName,
			indent,
			chevronIcon: (group.childGroups as unknown[])?.length
				? ChevronDown
				: ChevronRight,
			logoLetter: String(
				group.name ? (group.name as string)[0] : "?",
			).toUpperCase(),
			logoBg: "#6366f126",
			visibility: (group.visibility as "public" | "private") || "public",
			description: String(group.description || ""),
			type: "group",
			subgroups: Number((group.childGroups as unknown[])?.length || 0),
			projects: Number(group.projectCount || 0),
			members: Number(group.memberCount || 0),
			created: "",
			active: false,
			isChild,
		});
		// Only walk children if this group is expanded
		if (group.childGroups && expanded[groupName]) {
			for (const child of group.childGroups as Record<string, unknown>[]) {
				walk(child, indent + 18, true);
			}
		}
	}
	for (const g of items as Record<string, unknown>[]) {
		walk(g, 0, false);
	}
	return rows;
});

// ── Logging ─────────────────────────────────────────────────────────────────

watch(groupRows, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "groups.list.loaded",
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
				msg: "groups.query.error",
				error: String(err),
				ts: new Date().toISOString(),
			}),
		);
	}
});
</script>

<style scoped>
.groups-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.gp-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.gp-title-col {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.gp-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.gp-breadcrumb-text {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.gp-breadcrumb-sep {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 24px;
  font-weight: 600;
}

.gp-new-btn {
  height: 36px;
  border-radius: 6px;
  padding: 0 14px;
  background: var(--primary);
  color: var(--primary-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.gp-new-btn:hover {
  background: var(--primary-hover);
}

.gp-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.gp-search-wrap {
  flex: 1;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--card);
  padding: 0 10px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.gp-search-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-search-input {
  border: none;
  outline: none;
  background: transparent;
  width: 100%;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--foreground);
  opacity: 0.8;
}

.gp-search-input::placeholder {
  color: var(--muted-foreground);
}

.gp-sort-wrap {
  width: 372px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--card);
  padding: 0 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.gp-sort-label {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.gp-sort-chevron {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-sort-divider {
  width: 1px;
  height: 18px;
  background: var(--border);
  margin: 0 4px;
  flex-shrink: 0;
}

.gp-list {
  display: flex;
  flex-direction: column;
}

.gp-row {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  flex-direction: row;
  align-items: stretch;
  gap: 12px;
}

.gp-row:last-child {
  border-bottom: none;
}

.gp-row-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  justify-content: center;
}

.gp-row-body-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.gp-row-indent {
  flex-shrink: 0;
}

.gp-row-chevron {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-row-folder-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-row-logo {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 16px;
  font-weight: 600;
  color: var(--primary-foreground);
  flex-shrink: 0;
}

.gp-row-name {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  color: var(--foreground);
}

.gp-row-name--project {
  font-weight: 500;
}

.gp-row-name-active {
  color: var(--primary);
}

.gp-row-vis-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gp-row-desc {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.gp-row-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  justify-content: center;
  flex-shrink: 0;
}

.gp-row-counters {
  display: flex;
  flex-direction: row;
  gap: 4px;
}

.gp-row-created {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
  text-align: right;
}

.gp-counter {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: 70px;
  color: #6b7280;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 500;
  flex-shrink: 0;
}

.gp-counter svg {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.gp-row-menu-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 32px;
}

.gp-row-menu {
  color: #6b7280;
  flex-shrink: 0;
}

/* Skeleton loading */
.gp-skeleton-row {
  opacity: 0.6;
}

.skeleton {
  background: var(--muted);
  border-radius: 4px;
}

.skeleton--circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
}

.skeleton--text {
  height: 14px;
  flex: 0 0 200px;
}

/* Error state */
.gp-error,
.gp-empty {
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
</style>
