<template>
  <div class="group-detail-page" role="main" aria-label="Group detail page">
    <!-- Loading -->
    <div v-if="loading" class="gdp-state">
      <Loader :size="24" class="gdp-spinner" />
      <span>{{ t('common.loading') }}</span>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="gdp-state gdp-state--error" role="alert">
      <span>{{ error }}</span>
      <button class="gdp-retry-btn" type="button" @click="fetchGroup">{{ t('groups.retry') }}</button>
    </div>

    <!-- Data -->
    <template v-else-if="group">
      <!-- Breadcrumbs -->
      <div class="gdp-breadcrumbs">
        <router-link to="/dashboard/home" class="gdp-breadcrumb-link">{{ t('groups.breadcrumb_workspace') }}</router-link>
        <ChevronRight :size="12" class="gdp-breadcrumb-sep" />
        <router-link to="/dashboard/groups" class="gdp-breadcrumb-link">{{ t('groups.breadcrumb_groups') }}</router-link>
        <ChevronRight :size="12" class="gdp-breadcrumb-sep" />
        <span class="gdp-breadcrumb-current">{{ group.name }}</span>
      </div>

      <!-- Header -->
      <div class="gdp-header">
        <div class="gdp-name-row">
          <FolderTree :size="24" class="gdp-folder-icon" />
          <h1 class="gdp-title">{{ group.name }}</h1>
          <span :class="['gdp-vis-badge', `gdp-vis-badge--${group.visibility}`]">
            <component :is="visIcon" :size="12" />
            {{ group.visibility }}
          </span>
        </div>
        <p v-if="group.description" class="gdp-desc">{{ group.description }}</p>
      </div>

      <!-- Stats -->
      <div class="gdp-stats">
        <div class="gdp-stat">
          <FolderTree :size="14" />
          <span>{{ group.childGroups?.length ?? 0 }} {{ t('groups.subgroups') }}</span>
        </div>
        <div class="gdp-stat">
          <Folder :size="14" />
          <span>{{ group.projectCount }} {{ t('groups.projects') }}</span>
        </div>
        <div class="gdp-stat">
          <Users :size="14" />
          <span>{{ group.memberCount }} {{ t('groups.members') }}</span>
        </div>
      </div>

      <!-- Back button -->
      <router-link to="/dashboard/groups" class="gdp-back-btn">
        ← {{ t('groups.breadcrumb_groups') }}
      </router-link>
    </template>
  </div>
</template>

<script setup lang="ts">
import { type GroupInfo, getGroup } from "@/api/org";
import { useI18n } from "@/composables/useI18n";
import {
	ChevronRight,
	Folder,
	FolderTree,
	Globe,
	Loader,
	Lock,
	Shield,
	Users,
} from "@lucide/vue";
import type { Component } from "vue";
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();

const loading = ref(false);
const error = ref<string | null>(null);
const group = ref<GroupInfo | null>(null);

const visIconMap: Record<string, Component> = {
	private: Lock,
	internal: Shield,
	public: Globe,
};

const visIcon = computed(
	() => visIconMap[group.value?.visibility || ""] || Globe,
);

async function fetchGroup() {
	const id = route.params.id as string;
	if (!id) return;

	loading.value = true;
	error.value = null;

	console.info(
		JSON.stringify({
			event: "GroupDetailPage.fetch",
			group_id: id,
			ts: new Date().toISOString(),
		}),
	);

	try {
		group.value = await getGroup(id);
		console.info(
			JSON.stringify({
				event: "GroupDetailPage.success",
				group_id: id,
				name: group.value.name,
				ts: new Date().toISOString(),
			}),
		);
	} catch (err: unknown) {
		const msg = err instanceof Error ? err.message : String(err);
		error.value = msg || t("groups.load_error");
		console.error(
			JSON.stringify({
				event: "GroupDetailPage.error",
				group_id: id,
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		loading.value = false;
	}
}

onMounted(() => {
	fetchGroup();
});
</script>

<style scoped>
.group-detail-page {
  padding: 24px 32px;
  max-width: 800px;
}

.gdp-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 24px;
}

.gdp-breadcrumb-link {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  text-decoration: none;
}

.gdp-breadcrumb-link:hover {
  color: var(--foreground);
}

.gdp-breadcrumb-current {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--foreground);
  font-weight: 600;
}

.gdp-breadcrumb-sep {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gdp-header {
  margin-bottom: 24px;
}

.gdp-name-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.gdp-folder-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.gdp-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 28px;
  font-weight: 700;
}

.gdp-vis-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  font-weight: 600;
  text-transform: capitalize;
}

.gdp-vis-badge--private {
  background: #f59e0b1a;
  color: #f59e0b;
}

.gdp-vis-badge--internal {
  background: #3b82f61a;
  color: #3b82f6;
}

.gdp-vis-badge--public {
  background: #10b9811a;
  color: #10b981;
}

.gdp-desc {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
}

.gdp-stats {
  display: flex;
  gap: 24px;
  margin-bottom: 24px;
  padding: 16px;
  border-radius: 8px;
  background: var(--card);
  border: 1px solid var(--border);
}

.gdp-stat {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--muted-foreground);
}

.gdp-back-btn {
  display: inline-flex;
  align-items: center;
  text-decoration: none;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--primary, #6366f1);
}

.gdp-back-btn:hover {
  text-decoration: underline;
}

.gdp-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 64px 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
}

.gdp-state--error {
  color: var(--destructive, #ef4444);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.gdp-spinner {
  animation: spin 1s linear infinite;
  color: var(--muted-foreground);
}

.gdp-retry-btn {
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
