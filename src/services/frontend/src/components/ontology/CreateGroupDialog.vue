<!-- @m4 — Create Group Dialog -->
<!-- @hlv:artifact create-group-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Create Group" :modal="true" @close="$emit('close')">
    <div class="create-group-form">
      <div class="form-group">
        <label class="form-label">Group Name</label>
        <input
          v-model="groupName"
          type="text"
          class="form-input"
          placeholder="e.g. Research Team"
          :disabled="submitting"
          @keyup.enter="submit"
        />
        <p v-if="validationError?.field === 'name'" class="form-error">{{ validationError.message }}</p>
      </div>

      <div class="form-group">
        <label class="form-label">Description</label>
        <textarea
          v-model="description"
          class="form-input form-textarea"
          placeholder="Optional description"
          rows="3"
          :disabled="submitting"
        />
      </div>

      <div class="form-group">
        <label class="form-label">Visibility</label>
        <select v-model="visibility" class="form-select" :disabled="submitting">
          <option value="private">Private</option>
          <option value="public">Public</option>
        </select>
      </div>
    </div>

    <template #footer>
      <GhostButton :disabled="submitting" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="submitting" @click="submit">{{ submitting ? 'Creating...' : 'Create' }}</PrimaryButton>
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

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: []; created: [groupName: string] }>();

const { addError } = useErrorPresentation();

const groupName = ref("");
const description = ref("");
const visibility = ref("private");
const submitting = ref(false);
const validationError = ref<{ field: string; message: string } | null>(null);

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "CreateGroup.submitted",
			groupName: groupName.value,
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
	description.value = "";
	visibility.value = "private";
	validationError.value = null;
}
</script>

<style scoped>
.create-group-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-label {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 600;
  color: var(--foreground);
}

.form-input {
  height: 36px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--foreground);
  padding: 0 10px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  outline: none;
}

.form-input:focus {
  border-color: var(--primary);
}

.form-textarea {
  height: auto;
  padding: 8px 10px;
  resize: vertical;
}

.form-select {
  height: 36px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--foreground);
  padding: 0 10px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  outline: none;
}

.form-error {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--danger);
  margin: 0;
}
</style>
