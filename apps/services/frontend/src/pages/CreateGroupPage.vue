<template>
  <div class="create-group-page" role="main" aria-label="Create group page">
    <!-- Breadcrumbs -->
    <div class="cgp-breadcrumbs">
      <span class="cgp-breadcrumb-text">{{ t('groups.breadcrumb_workspace') }}</span>
      <ChevronRight :size="12" class="cgp-breadcrumb-sep" />
      <router-link to="/dashboard/groups" class="cgp-breadcrumb-link">{{ t('groups.breadcrumb_groups') }}</router-link>
      <ChevronRight :size="12" class="cgp-breadcrumb-sep" />
      <span class="cgp-breadcrumb-current">{{ t('groups.breadcrumb_new') }}</span>
    </div>

    <!-- Page Title -->
    <h1 class="cgp-title">{{ t('groups.create_title') }}</h1>
    <p class="cgp-desc">{{ t('groups.create_description') }}</p>

    <!-- Parent Group Info (when creating subgroup) -->
    <div v-if="parentName" class="cgp-parent-info">
      <FolderTree :size="16" />
      <span>{{ t('groups.parent_group') }}: {{ parentName }}</span>
    </div>

    <!-- Form Card -->
    <div class="cgp-card">
      <!-- Group Name -->
      <div class="cgp-field">
        <label class="cgp-label">{{ t('groups.group_name') }}</label>
        <input
          v-model="groupName"
          type="text"
          class="cgp-input"
          :placeholder="t('groups.name_placeholder')"
          aria-label="Group name"
          @input="onNameInput"
        />
        <p v-if="nameError" class="cgp-error-text">{{ nameError }}</p>
        <p class="cgp-help-text">{{ t('groups.name_help') }}</p>
      </div>

      <!-- Group URL / Slug (read-only) -->
      <div class="cgp-field">
        <label class="cgp-label">{{ t('groups.group_url') }}</label>
        <div class="cgp-slug-wrap">
          <span class="cgp-slug-prefix">vedo-core.local/</span>
          <span class="cgp-slug-value">{{ slugPreview || t('groups.slug_placeholder') }}</span>
        </div>
      </div>

      <!-- Visibility -->
      <div class="cgp-field">
        <label class="cgp-label">{{ t('groups.visibility') }}</label>
        <p class="cgp-help-text">{{ t('groups.visibility_help') }}</p>

        <!-- Inherited visibility (readonly for subgroup) -->
        <div v-if="isSubgroup" class="cgp-inherited-vis">
          <span>{{ t('groups.visibility_inherited') }}</span>
          <span class="cgp-inherited-value">{{ visibilityOptions[parentVisibility] }}</span>
        </div>

        <!-- Visibility selector (top-level) -->
        <div v-else class="cgp-vis-options">
          <label
            v-for="opt in visOptions"
            :key="opt.value"
            :class="['cgp-vis-option', { 'cgp-vis-option--selected': selectedVisibility === opt.value }]"
          >
            <input
              type="radio"
              :value="opt.value"
              v-model="selectedVisibility"
              class="cgp-vis-radio"
            />
            <div class="cgp-vis-body">
              <span class="cgp-vis-name">{{ t(opt.labelKey) }}</span>
              <span class="cgp-vis-desc">{{ t(opt.descKey) }}</span>
            </div>
            <component :is="opt.icon" :size="16" class="cgp-vis-icon" />
          </label>
        </div>
      </div>

      <!-- Actions -->
      <div class="cgp-actions">
        <router-link to="/dashboard/groups" class="cgp-btn-cancel">
          {{ t('groups.cancel') }}
        </router-link>
        <button
          class="cgp-btn-create"
          type="button"
          :disabled="creating"
          @click="handleCreate"
        >
          <Spinner v-if="creating" :size="14" />
          {{ creating ? t('groups.creating') : t('groups.create_button') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { createGroup } from "@/api/org";
import { useI18n } from "@/composables/useI18n";
import { useToast } from "@/composables/useToast";
import {
	ChevronRight,
	FolderTree,
	Globe,
	Lock,
	Loader as Spinner,
	Users,
} from "@lucide/vue";
import type { Component } from "vue";
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

const { t } = useI18n();
const { showToast } = useToast();
const route = useRoute();
const router = useRouter();

const groupName = ref("");
const nameError = ref("");
const creating = ref(false);
const parentName = ref("");
const parentVisibility = ref("private");

const parentId = computed(() => (route.query.parent_id as string) || "");
const isSubgroup = computed(() => !!parentId.value);

interface VisOption {
	value: string;
	labelKey: string;
	descKey: string;
	icon: Component;
}

const visOptions: VisOption[] = [
	{
		value: "private",
		labelKey: "groups.visibility_private",
		descKey: "groups.visibility_private_desc",
		icon: Lock,
	},
	{
		value: "internal",
		labelKey: "groups.visibility_internal",
		descKey: "groups.visibility_internal_desc",
		icon: Users,
	},
	{
		value: "public",
		labelKey: "groups.visibility_public",
		descKey: "groups.visibility_public_desc",
		icon: Globe,
	},
];

const visibilityOptions: Record<string, string> = {
	private: "Private",
	internal: "Internal",
	public: "Public",
};

const selectedVisibility = ref(
	isSubgroup.value ? parentVisibility.value : "private",
);

const slugPreview = computed(() => {
	if (!groupName.value) return "";
	return groupName.value
		.toLowerCase()
		.replace(/\s+/g, "-")
		.replace(/[^a-z0-9-]/g, "")
		.substring(0, 64);
});

function onNameInput() {
	nameError.value = "";
}

async function handleCreate() {
	const trimmed = groupName.value.trim();
	if (!trimmed) {
		nameError.value = t("groups.name_required");
		return;
	}

	creating.value = true;
	nameError.value = "";

	console.info(
		JSON.stringify({
			event: "CreateGroupPage.submit",
			name: trimmed,
			visibility: selectedVisibility.value,
			parent_id: parentId.value || null,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const result = await createGroup({
			name: trimmed,
			slug: slugPreview.value || undefined,
			visibility: selectedVisibility.value,
			parent_id: parentId.value || undefined,
		});

		console.info(
			JSON.stringify({
				event: "CreateGroupPage.success",
				group_id: result.id,
				name: trimmed,
				ts: new Date().toISOString(),
			}),
		);

		showToast(t("groups.create_success", { name: trimmed }));
		router.push(`/dashboard/groups/${result.id}`);
	} catch (err: unknown) {
		const msg = err instanceof Error ? err.message : String(err);
		console.error(
			JSON.stringify({
				event: "CreateGroupPage.error",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		showToast(t("groups.create_error"), "error");
	} finally {
		creating.value = false;
	}
}

onMounted(() => {
	console.info(
		JSON.stringify({
			event: "CreateGroupPage.mounted",
			parent_id: parentId.value || null,
			ts: new Date().toISOString(),
		}),
	);
});
</script>

<style scoped>
.create-group-page {
  padding: 24px 32px;
  max-width: 720px;
}

.cgp-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.cgp-breadcrumb-text,
.cgp-breadcrumb-link {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
}

.cgp-breadcrumb-link {
  text-decoration: none;
}

.cgp-breadcrumb-link:hover {
  color: var(--foreground);
}

.cgp-breadcrumb-current {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--foreground);
  font-weight: 600;
}

.cgp-breadcrumb-sep {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.cgp-title {
  margin: 0 0 8px 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 24px;
  font-weight: 600;
}

.cgp-desc {
  margin: 0 0 24px 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--muted-foreground);
}

.cgp-parent-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  margin-bottom: 24px;
  border-radius: 6px;
  background: var(--muted);
  border: 1px solid var(--border);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--foreground);
}

.cgp-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.cgp-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cgp-label {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  color: var(--foreground);
}

.cgp-input {
  height: 36px;
  padding: 0 12px;
  border-radius: 6px;
  border: 1px solid var(--input, var(--border));
  background: var(--card);
  color: var(--foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  opacity: 0.8;
}

.cgp-input::placeholder {
  color: var(--muted-foreground);
}

.cgp-help-text {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.cgp-error-text {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--destructive, #ef4444);
}

.cgp-slug-wrap {
  display: flex;
  align-items: center;
  height: 36px;
  padding: 0 12px;
  border-radius: 6px;
  border: 1px solid var(--input, var(--border));
  background: var(--card);
  gap: 0;
}

.cgp-slug-prefix {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--muted-foreground);
}

.cgp-slug-value {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--foreground);
  opacity: 0.5;
}

.cgp-inherited-vis {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  border-radius: 6px;
  background: var(--muted);
  border: 1px solid var(--border);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--muted-foreground);
}

.cgp-inherited-value {
  font-weight: 600;
  color: var(--foreground);
}

.cgp-vis-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cgp-vis-option {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--card);
  cursor: pointer;
  transition: background 0.15s;
}

.cgp-vis-option--selected {
  background: var(--secondary, #1a1a1a);
  border-color: var(--primary, #6366f1);
}

.cgp-vis-radio {
  margin-top: 2px;
  accent-color: var(--primary, #6366f1);
}

.cgp-vis-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.cgp-vis-name {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
  color: var(--foreground);
}

.cgp-vis-desc {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.cgp-vis-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.cgp-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 8px;
}

.cgp-btn-cancel {
  display: inline-flex;
  align-items: center;
  height: 36px;
  padding: 0 16px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
}

.cgp-btn-cancel:hover {
  background: var(--secondary);
}

.cgp-btn-create {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 16px;
  border-radius: 6px;
  border: none;
  background: var(--primary, #6366f1);
  color: var(--primary-foreground, #fff);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.cgp-btn-create:hover {
  background: var(--primary-hover, #4f46e5);
}

.cgp-btn-create:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
