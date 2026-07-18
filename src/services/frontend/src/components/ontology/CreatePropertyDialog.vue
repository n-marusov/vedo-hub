<!-- @m2.5 — Create Property Dialog -->
<!-- @hlv:artifact create-property-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Create Property" size="md" :modal="true" @close="$emit('close')">
    <div class="create-property-form">
      <div class="form-tabs">
        <button
          :class="['tab', { 'tab--active': activeTab === 'config' }]"
          @click="activeTab = 'config'"
        >
          Config
        </button>
        <button
          :class="['tab', { 'tab--active': activeTab === 'preview' }]"
          @click="activeTab = 'preview'"
        >
          Preview
        </button>
      </div>

      <div v-if="activeTab === 'config'" class="tab-content">
        <div class="form-group">
          <label class="form-label">Property Name</label>
          <input
            v-model="propertyName"
            type="text"
            class="form-input"
            placeholder="e.g. hasName"
            :disabled="submitting"
            @keyup.enter="submit"
          />
          <p v-if="validationError?.field === 'name'" class="form-error">{{ validationError.message }}</p>
        </div>

        <div class="form-group">
          <label class="form-label">Property Type</label>
          <select v-model="propertyType" class="form-select" :disabled="submitting">
            <option value="object">Object Property</option>
            <option value="datatype">Datatype Property</option>
            <option value="annotation">Annotation Property</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">Domain</label>
          <input
            v-model="domain"
            type="text"
            class="form-input"
            placeholder="e.g. Person"
            :disabled="submitting"
          />
        </div>

        <div class="form-group">
          <label class="form-label">Range</label>
          <input
            v-model="range"
            type="text"
            class="form-input"
            :placeholder="propertyType === 'datatype' ? 'e.g. xsd:string' : 'e.g. Organization'"
            :disabled="submitting"
          />
        </div>
      </div>

      <div v-if="activeTab === 'preview'" class="tab-content">
        <pre class="turtle-preview">:{{ propertyName || 'propertyName' }} a {{ typeToRdf(propertyType) }} ;
    rdfs:domain :{{ domain || 'DomainClass' }} ;
    rdfs:range {{ propertyType === 'datatype' ? (range || 'xsd:string') : ':' + (range || 'RangeClass') }} .</pre>
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
import { ref } from "vue";

defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: []; created: [propertyName: string] }>();

const { addError } = useErrorPresentation();

const activeTab = ref<"config" | "preview">("config");
const propertyName = ref("");
const propertyType = ref("object");
const domain = ref("");
const range = ref("");
const submitting = ref(false);
const validationError = ref<{ field: string; message: string } | null>(null);

function typeToRdf(type: string): string {
	switch (type) {
		case "object":
			return "owl:ObjectProperty";
		case "datatype":
			return "owl:DatatypeProperty";
		case "annotation":
			return "owl:AnnotationProperty";
		default:
			return "rdf:Property";
	}
}

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "CreateProperty.submitted",
			propName: propertyName.value,
			propType: propertyType.value,
			ts: new Date().toISOString(),
		}),
	);

	validationError.value = null;

	if (!propertyName.value.trim()) {
		validationError.value = {
			field: "name",
			message: "Property name is required",
		};
		return;
	}

	submitting.value = true;
	try {
		await new Promise((resolve) => setTimeout(resolve, 500));
		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "CreateProperty.success",
				propName: propertyName.value,
				ts: new Date().toISOString(),
			}),
		);
		emit("created", propertyName.value);
		reset();
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("CREATE-PROPERTY-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "CreateProperty.failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		submitting.value = false;
	}
}

function reset(): void {
	propertyName.value = "";
	propertyType.value = "object";
	domain.value = "";
	range.value = "";
	activeTab.value = "config";
	validationError.value = null;
}
</script>

<style scoped>
.create-property-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border);
}

.tab {
  padding: 8px 16px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  background: none;
  border: none;
  color: var(--muted-foreground);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.15s;
}

.tab--active {
  color: var(--foreground);
  border-bottom-color: var(--primary);
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
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

.turtle-preview {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 12px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--foreground);
  overflow-x: auto;
  white-space: pre-wrap;
}
</style>
