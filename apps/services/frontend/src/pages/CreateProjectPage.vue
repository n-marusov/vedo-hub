<template>
  <div class="create-project-page" role="main" aria-label="Create project page">
    <!-- Breadcrumbs -->
    <div class="cpp-breadcrumbs">
      <span class="cpp-breadcrumb-text">{{ t('projects.breadcrumb_workspace') }}</span>
      <ChevronRight :size="12" class="cpp-breadcrumb-sep" />
      <router-link to="/dashboard/projects" class="cpp-breadcrumb-link">{{ t('projects.breadcrumb_projects') }}</router-link>
      <ChevronRight :size="12" class="cpp-breadcrumb-sep" />
      <span class="cpp-breadcrumb-current">{{ t('projects.breadcrumb_new') }}</span>
    </div>

    <!-- Page Title -->
    <h1 class="cpp-title">{{ t('projects.create_title') }}</h1>
    <p class="cpp-desc">{{ t('projects.create_description') }}</p>

    <!-- Form Card -->
    <div class="cpp-card">
      <!-- Group Selector -->
      <div class="cpp-field">
        <label class="cpp-label" for="cpp-group-select">{{ t('projects.group') }}</label>
        <div class="cpp-select-wrap">
          <select
            id="cpp-group-select"
            v-model="selectedGroupId"
            class="cpp-select"
            :disabled="loadingGroups || creating"
            aria-label="Select group"
            @change="onGroupChange"
          >
            <option value="" disabled>{{ loadingGroups ? t('common.loading') : t('projects.group_placeholder') }}</option>
            <option v-for="g in groups" :key="g.id" :value="g.id">
              {{ g.name }}
            </option>
          </select>
        </div>
        <p v-if="groupError" class="cpp-error-text">{{ groupError }}</p>
        <p v-if="loadingGroups" class="cpp-help-text">{{ t('common.loading') }}</p>
      </div>

      <!-- Project Name -->
      <div class="cpp-field">
        <label class="cpp-label" for="cpp-project-name">{{ t('projects.name') }}</label>
        <input
          id="cpp-project-name"
          v-model="projectName"
          type="text"
          class="cpp-input"
          :placeholder="t('projects.name_placeholder')"
          aria-label="Project name"
          :disabled="creating"
          @input="onNameInput"
        />
        <p v-if="nameError" class="cpp-error-text">{{ nameError }}</p>
        <p class="cpp-help-text">{{ t('projects.name_hint') }}</p>
      </div>

	      <!-- Project URL / Slug (editable) -->
	      <div class="cpp-field">
	        <label class="cpp-label">{{ t('projects.project_url') }}</label>
	        <InputGroup
	          :id="'cpp-slug-input'"
	          v-model="slug"
	          :prefix="urlPrefix"
	          :placeholder="t('projects.slug_placeholder')"
	          :disabled="creating"
	          :aria-label="t('projects.project_url')"
	          @update:modelValue="onSlugInput"
	        />
	        <p class="cpp-help-text">{{ t('projects.slug_help') }}</p>
	      </div>

      <!-- Description -->
      <div class="cpp-field">
        <label class="cpp-label" for="cpp-project-desc">{{ t('projects.description') }}</label>
        <textarea
          id="cpp-project-desc"
          v-model="projectDescription"
          class="cpp-textarea"
          :placeholder="t('projects.description_placeholder')"
          aria-label="Project description"
          :disabled="creating"
          rows="3"
        ></textarea>
      </div>

      <!-- Visibility -->
      <div class="cpp-field">
        <label class="cpp-label">{{ t('projects.visibility') }}</label>
        <p class="cpp-help-text">{{ t('projects.visibility_help') }}</p>

        <div class="cpp-vis-options">
          <label
            v-for="opt in visOptions"
            :key="opt.value"
            :class="['cpp-vis-option', { 'cpp-vis-option--selected': selectedVisibility === opt.value }]"
          >
            <input
              type="radio"
              :value="opt.value"
              v-model="selectedVisibility"
              class="cpp-vis-radio"
              :disabled="creating"
            />
            <div class="cpp-vis-body">
              <span class="cpp-vis-name">{{ t(opt.labelKey) }}</span>
              <span class="cpp-vis-desc">{{ t(opt.descKey) }}</span>
            </div>
            <component :is="opt.icon" :size="16" class="cpp-vis-icon" />
          </label>
        </div>
      </div>

      <!-- Separator -->
      <div class="cpp-separator"></div>

      <!-- Actions -->
      <div class="cpp-actions">
        <router-link to="/dashboard/projects" class="cpp-btn-cancel">
          {{ t('projects.cancel') }}
        </router-link>
        <button
          class="cpp-btn-create"
          type="button"
          :disabled="creating"
          @click="handleCreate"
        >
          <Spinner v-if="creating" :size="14" />
          {{ creating ? t('projects.creating') : t('projects.create_button') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { createProject, listGroups } from "@/api/org";
import type { GroupInfo } from "@/api/org";
import InputGroup from "@/components/ui-kit/InputGroup.vue";
import { useI18n } from "@/composables/useI18n";
import { useToast } from "@/composables/useToast";
import { getPublicDomain } from "@/config";
import { slugify } from "@/utils/slug";
import {
	ChevronRight,
	Globe,
	Lock,
	Shield,
	Loader as Spinner,
} from "@lucide/vue";
import type { Component } from "vue";
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

const { t } = useI18n();
const { showToast } = useToast();
const router = useRouter();

const groups = ref<GroupInfo[]>([]);
const selectedGroupId = ref("");
const projectName = ref("");
const slug = ref("");
const slugTouched = ref(false);
const projectDescription = ref("");
const nameError = ref("");
const groupError = ref("");
const creating = ref(false);
const loadingGroups = ref(true);
const selectedVisibility = ref("Private");

interface VisOption {
	value: string;
	labelKey: string;
	descKey: string;
	icon: Component;
}

const visOptions: VisOption[] = [
	{
		value: "Private",
		labelKey: "projects.visibility_private",
		descKey: "projects.visibility_private_desc",
		icon: Lock,
	},
	{
		value: "Internal",
		labelKey: "projects.visibility_internal",
		descKey: "projects.visibility_internal_desc",
		icon: Shield,
	},
	{
		value: "Public",
		labelKey: "projects.visibility_public",
		descKey: "projects.visibility_public_desc",
		icon: Globe,
	},
];

const selectedGroupSlug = computed(() => {
	const g = groups.value.find((gr) => gr.id === selectedGroupId.value);
	if (!g) return "";
	// Prefer the server-provided slug; fall back to a slugified name so
	// the URL preview stays correct for legacy/no-slug groups.
	return g.slug || slugify(g.name, 32);
});

// Immutable URL prefix: {domain}/{groupSlug}/ — domain comes from the
// runtime env (VEDO_PUBLIC_DOMAIN).
const urlPrefix = computed(() => {
	const group = selectedGroupSlug.value;
	return group ? `${getPublicDomain()}/${group}/` : `${getPublicDomain()}/`;
});

function onNameInput() {
	nameError.value = "";
	// Auto-fill the slug from the name until the user edits it manually.
	// Transliteration (Cyrillic → Latin) is applied so non-Latin names
	// still produce a valid URL path segment.
	if (!slugTouched.value) {
		slug.value = slugify(projectName.value);
		console.info(
			JSON.stringify({
				level: "info",
				msg: "[FIX] CreateProjectPage.slug_autofill",
				name: projectName.value,
				slug: slug.value,
				transliterated: /[а-яё]/i.test(projectName.value),
				ts: new Date().toISOString(),
			}),
		);
	}
}

function onSlugInput(value: string) {
	slugTouched.value = true;
	// Sanitize what the user types: keep only [a-z0-9-], transliterate
	// any Cyrillic typed directly into the field.
	slug.value = slugify(value);
}

function onGroupChange() {
	groupError.value = "";
}

async function fetchGroups() {
	loadingGroups.value = true;
	try {
		groups.value = await listGroups();
		console.info(
			JSON.stringify({
				event: "CreateProjectPage.groups_loaded",
				count: groups.value.length,
				ts: new Date().toISOString(),
			}),
		);
	} catch (err: unknown) {
		const msg = err instanceof Error ? err.message : "Unknown error";
		console.error(
			JSON.stringify({
				event: "CreateProjectPage.groups_load_failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		showToast(t("projects.load_error"), "error");
	} finally {
		loadingGroups.value = false;
	}
}

async function handleCreate() {
	const trimmed = projectName.value.trim();
	nameError.value = "";
	groupError.value = "";

	if (!selectedGroupId.value) {
		groupError.value = t("projects.group_required");
		return;
	}

	if (!trimmed) {
		nameError.value = t("projects.name_required");
		return;
	}

	creating.value = true;

	console.info(
		JSON.stringify({
			event: "CreateProjectPage.submit",
			name: trimmed,
			group_id: selectedGroupId.value,
			visibility: selectedVisibility.value,
			ts: new Date().toISOString(),
		}),
	);

	try {
		const result = await createProject({
			name: trimmed,
			description: projectDescription.value || undefined,
			groupId: selectedGroupId.value,
			visibility: selectedVisibility.value,
			slug: slug.value || undefined,
		});

		console.info(
			JSON.stringify({
				event: "CreateProjectPage.success",
				project_id: result.id,
				name: trimmed,
				ontology_id: result.ontologyId,
				visibility: selectedVisibility.value,
				ts: new Date().toISOString(),
			}),
		);

		showToast(t("projects.create_success"), "success");
		router.push(`/project/${result.id}/workspace`);
	} catch (err: unknown) {
		const msg = err instanceof Error ? err.message : "Unknown error";
		console.error(
			JSON.stringify({
				event: "CreateProjectPage.error",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
		showToast(t("projects.create_error", { error: msg }), "error");
	} finally {
		creating.value = false;
	}
}

onMounted(() => {
	console.info(
		JSON.stringify({
			event: "CreateProjectPage.mounted",
			ts: new Date().toISOString(),
		}),
	);
	fetchGroups();
});
</script>

<style scoped>
.create-project-page {
  padding: 24px 32px;
  max-width: 720px;
}

.cpp-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.cpp-breadcrumb-text,
.cpp-breadcrumb-link {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
}

.cpp-breadcrumb-link {
  text-decoration: none;
}

.cpp-breadcrumb-link:hover {
  color: var(--foreground);
}

.cpp-breadcrumb-current {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--foreground);
  font-weight: 600;
}

.cpp-breadcrumb-sep {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.cpp-title {
  margin: 0 0 8px 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 24px;
  font-weight: 600;
}

.cpp-desc {
  margin: 0 0 24px 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
}

.cpp-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.cpp-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cpp-label {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  color: var(--foreground);
}

.cpp-input,
.cpp-select {
  height: 36px;
  padding: 0 12px;
  border-radius: 6px;
  border: 1px solid var(--input, var(--border));
  background: var(--card);
  color: var(--foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.cpp-input::placeholder {
  color: var(--muted-foreground);
}

.cpp-select {
  width: 100%;
  cursor: pointer;
  appearance: auto;
}

.cpp-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cpp-select option {
  color: var(--foreground);
  background: var(--card);
}

.cpp-textarea {
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid var(--input, var(--border));
  background: var(--card);
  color: var(--foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  resize: vertical;
  min-height: 60px;
}

.cpp-textarea::placeholder {
  color: var(--muted-foreground);
}

.cpp-help-text {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.cpp-error-text {
	margin: 0;
	font-family: 'IBM Plex Mono', monospace;
	font-size: 12px;
	color: var(--destructive, #ef4444);
}

.cpp-vis-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cpp-vis-option {
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

.cpp-vis-option--selected {
  background: var(--secondary, #1a1a1a);
  border-color: var(--primary, #6366f1);
}

.cpp-vis-radio {
  margin-top: 2px;
  accent-color: var(--primary, #6366f1);
}

.cpp-vis-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.cpp-vis-name {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
  color: var(--foreground);
}

.cpp-vis-desc {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.cpp-vis-icon {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.cpp-separator {
  width: 100%;
  height: 1px;
  background: var(--border);
}

.cpp-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 8px;
}

.cpp-btn-cancel {
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

.cpp-btn-cancel:hover {
  background: var(--secondary);
}

.cpp-btn-create {
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

.cpp-btn-create:hover {
  background: var(--primary-hover, #4f46e5);
}

.cpp-btn-create:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
