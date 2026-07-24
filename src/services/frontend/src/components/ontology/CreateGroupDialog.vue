<!-- @m4 — Create Group Dialog -->
<!-- @hlv:artifact create-group-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Create group" size="lg" :modal="true" @close="$emit('close')">
    <form class="create-group-form" @submit.prevent="submit">
      <section class="intro-section">
        <p>
          Groups allow you to manage and collaborate across multiple projects. Members of a group have access to all of its projects.
        </p>
        <p>Groups can also be nested by creating subgroups.</p>
      </section>

      <div class="form-group">
        <label class="form-label" for="group-name">Group name</label>
        <input
          id="group-name"
          v-model="groupName"
          type="text"
          class="form-input"
          placeholder="My group"
          :disabled="submitting"
          @input="syncSlugFromName"
        />
        <p class="form-help">Start with a letter, digit, emoji, or underscore.</p>
        <p class="form-help">
          Your group name must not contain a period if you intend to use SCIM integration, as it can lead to errors.
        </p>
        <p v-if="validationError?.field === 'name'" class="form-error">{{ validationError.message }}</p>
      </div>

      <div class="form-group">
        <label class="form-label" for="group-slug">Group URL</label>
        <div class="url-row">
          <span class="url-prefix">https://vedo-hub.ru/</span>
          <input
            id="group-slug"
            v-model="groupSlug"
            type="text"
            class="form-input url-input"
            placeholder="my-awesome-group"
            :disabled="submitting"
          />
        </div>
      </div>

      <fieldset class="form-group radio-fieldset">
        <legend class="form-label">Visibility level</legend>
        <p class="form-help">Who will be able to see this group? View the documentation</p>
        <label v-for="option in visibilityOptions" :key="option.value" class="visibility-option">
          <input v-model="visibility" type="radio" name="group-visibility" :value="option.value" :disabled="submitting" />
          <span class="visibility-copy">
            <span class="visibility-title">{{ option.label }}</span>
            <span class="visibility-description">{{ option.description }}</span>
          </span>
        </label>
      </fieldset>

      <fieldset class="form-group radio-fieldset">
        <legend class="form-label">Who will be using this group?</legend>
        <label class="compact-option">
          <input v-model="usage" type="radio" name="group-usage" value="team" :disabled="submitting" />
          <span>My company or team</span>
        </label>
        <label class="compact-option">
          <input v-model="usage" type="radio" name="group-usage" value="solo" :disabled="submitting" />
          <span>Just me</span>
        </label>
      </fieldset>

      <section class="invite-section">
        <div>
          <h3 class="section-title">Invite Members (optional)</h3>
          <p class="form-help">
            Invited users will be added with developer level permissions. View the documentation to see how to change this later.
          </p>
        </div>
        <div class="form-group">
          <label class="form-label" for="invite-email-1">Email 1</label>
          <input
            id="invite-email-1"
            v-model="inviteEmail"
            type="email"
            class="form-input"
            placeholder="member1@company.com"
            :disabled="submitting"
          />
        </div>
        <button class="invite-more-btn" type="button" :disabled="submitting">
          + Invite another member
        </button>
      </section>
    </form>

    <template #footer>
      <GhostButton :disabled="submitting" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="submitting" @click="submit">{{ submitting ? 'Creating...' : 'Create group' }}</PrimaryButton>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import { createGroup } from "@/api/org";
import Dialog from "@/components/ui-kit/Dialog.vue";
import GhostButton from "@/components/ui-kit/GhostButton.vue";
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { ref } from "vue";

defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: []; created: [groupName: string] }>();

const { addError } = useErrorPresentation();

const groupName = ref("");
const groupSlug = ref("my-awesome-group");
const description = ref("");
const visibility = ref("private");
const usage = ref("team");
const inviteEmail = ref("");
const submitting = ref(false);
const validationError = ref<{ field: string; message: string } | null>(null);

const visibilityOptions = [
	{
		value: "private",
		label: "Private",
		description: "The group and its projects can only be viewed by members.",
	},
	{
		value: "internal",
		label: "Internal",
		description:
			"The group and any internal projects can be viewed by any logged in user except external users.",
	},
	{
		value: "public",
		label: "Public",
		description:
			"The group and any public projects can be viewed without any authentication.",
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
	const slug = slugify(groupName.value);
	groupSlug.value = slug || groupSlug.value;
}

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "CreateGroup.submitted",
			groupName: groupName.value,
			groupSlug: groupSlug.value,
			visibility: visibility.value,
			inviteEmail: inviteEmail.value || undefined,
			ts: new Date().toISOString(),
		}),
	);

	validationError.value = null;

	if (!groupName.value.trim()) {
		validationError.value = {
			field: "name",
			message: "Group name is required",
		};
		return;
	}

	submitting.value = true;
	try {
		await createGroup({
			name: groupName.value.trim(),
			description: description.value.trim() || undefined,
			visibility: visibility.value,
		});

		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "CreateGroup.success",
				groupName: groupName.value,
				ts: new Date().toISOString(),
			}),
		);
		emit("created", groupName.value);
		reset();
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("CREATE-GROUP-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "CreateGroup.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		submitting.value = false;
	}
}

function reset(): void {
	groupName.value = "";
	groupSlug.value = "my-awesome-group";
	description.value = "";
	visibility.value = "private";
	usage.value = "team";
	inviteEmail.value = "";
	validationError.value = null;
}
</script>

<style scoped>
.create-group-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.intro-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-normal);
}

.intro-section p,
.form-help,
.form-error {
  margin: 0;
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

.form-input {
  width: 100%;
  height: 40px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-default);
  background: var(--surface);
  color: var(--foreground);
  padding: 0 var(--space-3);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  outline: none;
}

.form-input:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-muted);
}

.form-help {
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-normal);
}

.form-error {
  color: var(--error);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
}

.url-row {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  align-items: center;
  overflow: hidden;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface);
}

.url-prefix {
  height: 40px;
  display: inline-flex;
  align-items: center;
  padding: 0 var(--space-3);
  border-right: 1px solid var(--border-default);
  background: var(--surface-variant);
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
}

.url-input {
  border: 0;
  border-radius: 0;
}

.url-input:focus {
  box-shadow: none;
}

.radio-fieldset {
  margin: 0;
  padding: 0;
  border: 0;
}

.visibility-option,
.compact-option {
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

.compact-option {
  align-items: center;
}

.visibility-option:has(input:checked),
.compact-option:has(input:checked) {
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

.visibility-description,
.compact-option span {
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-normal);
}

.invite-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-default);
}

.section-title {
  margin: 0 0 var(--space-1);
  color: var(--foreground);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.invite-more-btn {
  align-self: flex-start;
  border: 0;
  background: transparent;
  color: var(--text-link);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
}

.invite-more-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 640px) {
  .url-row {
    grid-template-columns: 1fr;
  }

  .url-prefix {
    border-right: 0;
    border-bottom: 1px solid var(--border-default);
  }
}
</style>
