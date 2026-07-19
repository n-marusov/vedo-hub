<!-- @m4 — Create Individual Dialog -->
<!-- @hlv:artifact create-individual-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Create Individual" size="md" :modal="true" @close="$emit('close')">
    <div class="create-individual-form">
      <div class="form-group">
        <label class="form-label">Individual Name</label>
        <input
          v-model="individualName"
          type="text"
          class="form-input"
          placeholder="e.g. JohnDoe"
          :disabled="submitting"
          @keyup.enter="submit"
        />
        <p v-if="validationError?.field === 'name'" class="form-error">{{ validationError.message }}</p>
      </div>

      <div class="form-group">
        <label class="form-label">Class</label>
        <select v-model="selectedClass" class="form-select" :disabled="submitting">
          <option value="">— Select class —</option>
          <option value="owl:Thing">owl:Thing</option>
          <option value="Person">Person</option>
          <option value="Organization">Organization</option>
        </select>
      </div>

      <div class="form-group">
        <label class="form-label">Properties</label>
        <div v-for="(pv, i) in propertyValues" :key="i" class="pv-row">
          <input v-model="pv.property" class="form-input form-input--sm" placeholder="Property" :disabled="submitting" />
          <input v-model="pv.value" class="form-input form-input--sm" placeholder="Value" :disabled="submitting" />
          <GhostButton :disabled="submitting" @click="propertyValues.splice(i, 1)">✕</GhostButton>
        </div>
        <GhostButton :disabled="submitting" @click="propertyValues.push({ property: '', value: '' })">+ Add property</GhostButton>
      </div>
    </div>

    <template #footer>
      <GhostButton :disabled="submitting" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="submitting" @click="submit">{{ submitting ? 'Creating...' : 'Create' }}</PrimaryButton>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import Dialog from "@/components/ui-kit/Dialog.vue";
import GhostButton from "@/components/ui-kit/GhostButton.vue";
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { reactive, ref } from "vue";

defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: []; created: [individualName: string] }>();

const { addError } = useErrorPresentation();

const individualName = ref("");
const selectedClass = ref("");
const propertyValues = reactive<Array<{ property: string; value: string }>>([]);
const submitting = ref(false);
const validationError = ref<{ field: string; message: string } | null>(null);

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "CreateIndividual.submitted",
			indName: individualName.value,
			className: selectedClass.value,
			ts: new Date().toISOString(),
		}),
	);

	validationError.value = null;

	if (!individualName.value.trim()) {
		validationError.value = {
			field: "name",
			message: "Individual name is required",
		};
		return;
	}

	submitting.value = true;
	try {
		await new Promise((resolve) => setTimeout(resolve, 500));
		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "CreateIndividual.success",
				indName: individualName.value,
				ts: new Date().toISOString(),
			}),
		);
		emit("created", individualName.value);
		reset();
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("CREATE-INDIVIDUAL-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "CreateIndividual.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		submitting.value = false;
	}
}

function reset(): void {
	individualName.value = "";
	selectedClass.value = "";
	propertyValues.length = 0;
	validationError.value = null;
}
</script>

<style scoped>
.create-individual-form {
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
  flex: 1;
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

.pv-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
