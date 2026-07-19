<!-- @m4 — Publish Snapshot Dialog -->
<!-- @hlv:artifact publish-snapshot-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Publish Snapshot" size="md" :modal="true" @close="$emit('close')">
    <div class="publish-form">
      <div class="form-group">
        <label class="form-label">Version / Tag Name</label>
        <input
          v-model="versionName"
          type="text"
          class="form-input"
          placeholder="e.g. v1.0.0"
          :disabled="publishing"
        />
      </div>

      <div class="form-group">
        <label class="form-label">Visibility</label>
        <select v-model="visibility" class="form-select" :disabled="publishing">
          <option value="public">Public</option>
          <option value="private">Private</option>
        </select>
      </div>

      <div class="form-group">
        <label class="form-label">Description</label>
        <textarea
          v-model="description"
          class="form-input form-textarea"
          placeholder="Snapshot description (optional)"
          rows="3"
          :disabled="publishing"
        />
      </div>

      <div class="form-group">
        <label class="form-label">Published URL</label>
        <div class="url-preview-row">
          <input
            :value="generatedUrl"
            class="form-input"
            readonly
          />
          <GhostButton @click="copyUrl">Copy</GhostButton>
        </div>
      </div>
    </div>

    <template #footer>
      <GhostButton :disabled="publishing" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="publishing" :disabled="!versionName.trim()" @click="publish">
        {{ publishing ? 'Publishing...' : 'Publish' }}
      </PrimaryButton>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import Dialog from "@/components/ui-kit/Dialog.vue";
import GhostButton from "@/components/ui-kit/GhostButton.vue";
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { computed, ref } from "vue";

const props = defineProps<{
	open: boolean;
	ontologyId?: string;
	ontologyName?: string;
}>();
const emit = defineEmits<{ close: []; published: [versionName: string] }>();

const { addError } = useErrorPresentation();

const versionName = ref("");
const visibility = ref("public");
const description = ref("");
const publishing = ref(false);
const copied = ref(false);

const generatedUrl = computed(() => {
	const slug = versionName.value.trim() || "latest";
	return `https://vedo.app/public/${props.ontologyId || "{ontology}"}/v/${slug}`;
});

async function publish(): Promise<void> {
	if (!versionName.value.trim()) return;

	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "PublishSnapshot.submitted",
			version: versionName.value,
			visibility: visibility.value,
			ts: new Date().toISOString(),
		}),
	);

	publishing.value = true;
	try {
		// [bookmark] Full publish flow not in M4 scope — mock publish
		await new Promise((resolve) => setTimeout(resolve, 800));
		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "PublishSnapshot.success",
				version: versionName.value,
				url: generatedUrl.value,
				ts: new Date().toISOString(),
			}),
		);
		emit("published", versionName.value);
		reset();
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("PUBLISH-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "PublishSnapshot.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		publishing.value = false;
	}
}

async function copyUrl(): Promise<void> {
	try {
		await navigator.clipboard.writeText(generatedUrl.value);
		copied.value = true;
		setTimeout(() => {
			copied.value = false;
		}, 2000);
	} catch {
		addError("CLIPBOARD-FAILED", "Failed to copy URL to clipboard");
	}
}

function reset(): void {
	versionName.value = "";
	visibility.value = "public";
	description.value = "";
	copied.value = false;
}
</script>

<style scoped>
.publish-form {
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

.form-input[readonly] {
  opacity: 0.7;
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

.url-preview-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.url-preview-row .form-input {
  flex: 1;
}
</style>
