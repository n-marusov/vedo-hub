<!-- @m4 — Create Class Dialog -->
<!-- @hlv:artifact create-class-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Create Class" :modal="true" @close="$emit('close')">
    <div class="create-class-form">
      <div class="form-group">
        <label class="form-label">Class Name</label>
        <input
          v-model="className"
          type="text"
          class="form-input"
          placeholder="e.g. Person"
          :disabled="submitting"
          @keyup.enter="submit"
        />
        <p v-if="validationError?.field === 'name'" class="form-error">{{ validationError.message }}</p>
      </div>

      <div class="form-group">
        <label class="form-label">Parent Class</label>
        <select v-model="parentClass" class="form-select" :disabled="submitting">
          <option value="">— None (top-level class) —</option>
          <option value="owl:Thing">owl:Thing</option>
        </select>
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
        <label class="form-label">Annotations</label>
        <div v-for="(ann, i) in annotations" :key="i" class="annotation-row">
          <input v-model="ann.key" class="form-input form-input--sm" placeholder="Key" :disabled="submitting" />
          <input v-model="ann.value" class="form-input form-input--sm" placeholder="Value" :disabled="submitting" />
          <GhostButton :disabled="submitting" @click="annotations.splice(i, 1)">✕</GhostButton>
        </div>
        <GhostButton :disabled="submitting" @click="annotations.push({ key: '', value: '' })">+ Add annotation</GhostButton>
      </div>
    </div>

    <template #footer>
      <GhostButton :disabled="submitting" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="submitting" @click="submit">{{ submitting ? 'Creating...' : 'Create' }}</PrimaryButton>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
// @aif — Migrated from Apollo GraphQL `CREATE_CLASS_MUTATION` to REST
// `POST /api/v1/ontologies/:id/classes` per ADR-DES.API.rest-graphql-mutation-boundary.md.
import { createClass } from "@/api/ontology";
import Dialog from "@/components/ui-kit/Dialog.vue";
import GhostButton from "@/components/ui-kit/GhostButton.vue";
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { reactive, ref } from "vue";

const props = defineProps<{ open: boolean; ontologyId: string }>();
const emit = defineEmits<{ close: []; created: [className: string] }>();

const { addError } = useErrorPresentation();

const className = ref("");
const parentClass = ref("");
const description = ref("");
const annotations = reactive<Array<{ key: string; value: string }>>([]);
const submitting = ref(false);
const validationError = ref<{ field: string; message: string } | null>(null);

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "CreateClass.submitted",
			className: className.value,
			ontologyId: props.ontologyId,
			ts: new Date().toISOString(),
		}),
	);

	validationError.value = null;

	if (!className.value.trim()) {
		validationError.value = {
			field: "name",
			message: "Class name is required",
		};
		return;
	}

	submitting.value = true;
	try {
		const result = await createClass({
			ontologyId: props.ontologyId,
			label: className.value.trim(),
			parentId: parentClass.value || undefined,
			description: description.value.trim() || undefined,
			annotations:
				annotations.length > 0
					? annotations
							.filter((a) => a.key)
							.map((a) => ({
								propertyIri: a.key,
								value: a.value,
							}))
					: undefined,
		});

		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "CreateClass.success",
				className: className.value,
				classId: result.id,
				ts: new Date().toISOString(),
			}),
		);
		emit("created", className.value);
		reset();
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("CREATE-CLASS-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "CreateClass.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		submitting.value = false;
	}
}

function reset(): void {
	className.value = "";
	parentClass.value = "";
	description.value = "";
	annotations.length = 0;
	validationError.value = null;
}
</script>

<style scoped>
.create-class-form {
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

.form-input--sm {
  flex: 1;
}

.form-error {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--danger);
  margin: 0;
}

.annotation-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
