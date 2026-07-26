<!-- @m4 — Annotation Dialog -->
<!-- @hlv:artifact annotation-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Annotation" size="sm" :modal="true" @close="$emit('close')">
    <div class="annotation-form">
      <div class="form-group">
        <label class="form-label">Annotation Property</label>
        <select v-model="annotationProperty" class="form-select" :disabled="submitting">
          <option value="">— Select —</option>
          <option value="rdfs:label">rdfs:label</option>
          <option value="rdfs:comment">rdfs:comment</option>
          <option value="rdfs:seeAlso">rdfs:seeAlso</option>
          <option value="skos:definition">skos:definition</option>
          <option value="skos:note">skos:note</option>
        </select>
      </div>

      <div class="form-group">
        <label class="form-label">Value</label>
        <input
          v-model="value"
          type="text"
          class="form-input"
          placeholder="Annotation value"
          :disabled="submitting"
        />
      </div>

      <div class="form-group">
        <label class="form-label">Language</label>
        <input
          v-model="language"
          type="text"
          class="form-input form-input--sm"
          placeholder="e.g. en (optional)"
          maxlength="5"
          :disabled="submitting"
        />
      </div>
    </div>

    <template #footer>
      <GhostButton :disabled="submitting" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="submitting" @click="submit">Save</PrimaryButton>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import Dialog from "@/components/ui-kit/Dialog.vue";
import GhostButton from "@/components/ui-kit/GhostButton.vue";
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { ref } from "vue";

defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: []; saved: [] }>();

const { addError } = useErrorPresentation();

const annotationProperty = ref("");
const value = ref("");
const language = ref("");
const submitting = ref(false);

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "Annotation.submitted",
			property: annotationProperty.value,
			ts: new Date().toISOString(),
		}),
	);

	if (!annotationProperty.value || !value.value.trim()) {
		addError("ANNOTATION-VALIDATION", "Property and value are required");
		return;
	}

	submitting.value = true;
	try {
		await new Promise((resolve) => setTimeout(resolve, 500));
		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "Annotation.success",
				property: annotationProperty.value,
				ts: new Date().toISOString(),
			}),
		);
		emit("saved");
		reset();
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("ANNOTATION-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "Annotation.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		submitting.value = false;
	}
}

function reset(): void {
	annotationProperty.value = "";
	value.value = "";
	language.value = "";
}
</script>

<style scoped>
.annotation-form {
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

.form-input--sm {
  width: 120px;
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
</style>
