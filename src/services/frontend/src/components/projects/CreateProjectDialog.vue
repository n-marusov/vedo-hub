<!-- @m4 — Create Project Dialog -->
<!-- @hlv:artifact create-project-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Create project" size="lg" :modal="true" @close="$emit('close')">
    <form class="create-project-form" @submit.prevent="submit">
      <div class="form-group">
        <label class="form-label" for="project-name">Project name</label>
        <input
          id="project-name"
          v-model="projectName"
          type="text"
          class="form-input"
          placeholder="My project"
          :disabled="submitting"
          @input="syncSlugFromName"
        />
        <p class="form-help">Start with a letter, digit, emoji, or underscore.</p>
        <p v-if="validationError?.field === 'name'" class="form-error">{{ validationError.message }}</p>
      </div>

      <div class="form-group">
        <label class="form-label" for="project-slug">Project URL</label>
        <div class="project-url-row">
          <span class="url-host">https://vedo-hub.ru/</span>
          <select v-model="namespace" class="namespace-select" :disabled="submitting" aria-label="Project namespace">
            <option value="workspace">workspace</option>
            <option value="team">team</option>
          </select>
          <span class="url-separator">/</span>
          <input
            id="project-slug"
            v-model="projectSlug"
            type="text"
            class="form-input slug-input"
            placeholder="project-slug"
            :disabled="submitting"
          />
        </div>
      </div>

      <fieldset class="form-group radio-fieldset">
        <legend class="form-label">Visibility Level</legend>
        <label v-for="option in visibilityOptions" :key="option.value" class="visibility-option">
          <input v-model="visibility" type="radio" name="project-visibility" :value="option.value" :disabled="submitting" />
          <span class="visibility-copy">
            <span class="visibility-title">{{ option.label }}</span>
            <span class="visibility-description">{{ option.description }}</span>
          </span>
        </label>
      </fieldset>
    </form>

    <template #footer>
      <GhostButton :disabled="submitting" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="submitting" @click="submit">{{ submitting ? 'Creating...' : 'Create project' }}</PrimaryButton>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import { createProject } from "@/api/org";
import Dialog from "@/components/ui-kit/Dialog.vue";
import GhostButton from "@/components/ui-kit/GhostButton.vue";
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { ref } from "vue";

defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: []; created: [projectName: string] }>();

const { addError } = useErrorPresentation();

const projectName = ref("");
const projectSlug = ref("project-slug");
const namespace = ref("workspace");
const visibility = ref("private");
const submitting = ref(false);
const validationError = ref<{ field: string; message: string } | null>(null);

const visibilityOptions = [
	{
		value: "private",
		label: "Private",
		description:
			"Project access must be granted explicitly to each user. If this project is part of a group, access is granted to members of the group.",
	},
	{
		value: "internal",
		label: "Internal",
		description:
			"The project can be accessed by any logged in user except external users.",
	},
	{
		value: "public",
		label: "Public",
		description: "The project can be accessed without any authentication.",
	},
];

function slugify(value: string): string {
	return value
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9_\s-]/g, "")
		.replace(/[\s_]+/g, "-")
		.replace(/-+/g, "-")
		.replace(/^-|-$/g, "");
}

function syncSlugFromName(): void {
	const slug = slugify(projectName.value);
	projectSlug.value = slug || projectSlug.value;
}

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "CreateProject.submitted",
			projectName: projectName.value,
			namespace: namespace.value,
			projectSlug: projectSlug.value,
			visibility: visibility.value,
			ts: new Date().toISOString(),
		}),
	);

	validationError.value = null;

	if (!projectName.value.trim()) {
		validationError.value = {
			field: "name",
			message: "Project name is required",
		};
		return;
	}

	submitting.value = true;
	try {
		await createProject({
			name: projectName.value.trim(),
			description: undefined,
			groupId: undefined,
		});

		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "CreateProject.success",
				projectName: projectName.value,
				ts: new Date().toISOString(),
			}),
		);
		emit("created", projectName.value);
		reset();
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("CREATE-PROJECT-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "CreateProject.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		submitting.value = false;
	}
}

function reset(): void {
	projectName.value = "";
	projectSlug.value = "project-slug";
	namespace.value = "workspace";
	visibility.value = "private";
	validationError.value = null;
}
</script>

<style scoped>
.create-project-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--foreground);
}

.form-input,
.namespace-select {
  height: 40px;
  border: 1px solid var(--border-default);
  background: var(--surface);
  color: var(--foreground);
  padding: 0 var(--space-3);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  outline: none;
}

.form-input {
  width: 100%;
  border-radius: var(--radius-md);
}

.form-input:focus,
.namespace-select:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-muted);
}

.form-help,
.form-error {
  margin: 0;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-normal);
}

.form-help {
  color: var(--text-secondary);
}

.form-error {
  color: var(--error);
}

.project-url-row {
  display: grid;
  grid-template-columns: max-content minmax(120px, 180px) max-content minmax(160px, 1fr);
  align-items: center;
  overflow: hidden;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface);
}

.url-host,
.url-separator {
  height: 40px;
  display: inline-flex;
  align-items: center;
  background: var(--surface-variant);
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
}

.url-host {
  padding: 0 var(--space-3);
  border-right: 1px solid var(--border-default);
}

.url-separator {
  justify-content: center;
  padding: 0 var(--space-2);
  border-left: 1px solid var(--border-default);
  border-right: 1px solid var(--border-default);
}

.namespace-select,
.slug-input {
  border: 0;
  border-radius: 0;
}

.slug-input:focus,
.namespace-select:focus {
  box-shadow: none;
}

.radio-fieldset {
  margin: 0;
  padding: 0;
  border: 0;
}

.visibility-option {
  display: flex;
  gap: var(--space-3);
  align-items: flex-start;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  padding: var(--space-3);
  background: var(--surface);
  color: var(--foreground);
  cursor: pointer;
}

.visibility-option:has(input:checked) {
  border-color: var(--primary);
  background: var(--primary-muted);
}

.visibility-copy {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.visibility-title {
  color: var(--foreground);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.visibility-description {
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-normal);
}

@media (max-width: 760px) {
  .project-url-row {
    grid-template-columns: 1fr;
  }

  .url-host,
  .url-separator {
    border-right: 0;
    border-left: 0;
    border-bottom: 1px solid var(--border-default);
    justify-content: flex-start;
    padding: 0 var(--space-3);
  }
}
</style>
