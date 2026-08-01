<!-- @m4 — Create Class Dialog -->
<!-- @hlv:artifact create-class-dialog implements GUI-OW-001 -->
<template>
  <Teleport to="body">
    <div v-if="open" class="dialog-overlay" role="dialog" aria-modal="true" aria-labelledby="create-class-dialog-title" @click.self="onOverlayClick">
      <div class="dialog-card">
        <!-- Close button — absolute top-right -->
        <button class="dialog-close" aria-label="Close" @click="$emit('close')">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 4l8 8M12 4l-8 8" />
          </svg>
        </button>

        <!-- Header -->
        <h2 id="create-class-dialog-title" class="dialog-header">Create Class</h2>

        <!-- Form body -->
        <div class="dialog-form">
          <!-- Class IRI — readonly, auto-generated -->
          <div class="form-field">
            <label class="ff-label">Class IRI</label>
            <div class="ff-input ff-input--disabled">
              <span class="ff-placeholder">{{ classIriPreview || 'Will be auto-generated from Label' }}</span>
            </div>
            <span class="ff-hint">Unique identifier for the class</span>
          </div>

          <!-- rdfs:label * -->
          <div class="form-field">
            <label class="ff-label">rdfs:label *</label>
            <input
              v-model="className"
              class="ff-input form-input"
              type="text"
              placeholder="Enter class name"
              :disabled="submitting"
              @keyup.enter="submit"
            />
            <span v-if="validationError?.field === 'name'" class="ff-error">{{ validationError.message }}</span>
          </div>

          <!-- rdfs:comment -->
          <div class="form-field">
            <label class="ff-label">rdfs:comment</label>
            <textarea
              v-model="description"
              class="ff-input ff-textarea"
              placeholder="Enter description"
              :disabled="submitting"
            ></textarea>
          </div>

          <!-- Subclass of -->
          <div class="form-field">
            <label class="ff-label">Subclass of</label>
            <div class="ff-select-wrapper">
              <select v-model="parentClass" class="ff-select" :disabled="submitting">
                <option value="">— None (top-level class) —</option>
                <option value="owl:Thing">owl:Thing</option>
              </select>
              <svg class="ff-select-chevron" width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M4 6l3 3 3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </div>
          </div>

          <!-- Flags -->
          <div class="form-checks">
            <label class="check-item" :class="{ 'check-item--disabled': submitting }">
              <span class="check-box" aria-hidden="true">
                <svg v-if="isAbstract" width="10" height="10" viewBox="0 0 10 10" fill="none">
                  <path d="M2 5l2 2 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>
              <input v-model="isAbstract" type="checkbox" class="check-input" :disabled="submitting" />
              <span class="check-label">Abstract</span>
            </label>
            <label class="check-item" :class="{ 'check-item--disabled': submitting }">
              <span class="check-box" aria-hidden="true">
                <svg v-if="isDeprecated" width="10" height="10" viewBox="0 0 10 10" fill="none">
                  <path d="M2 5l2 2 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>
              <input v-model="isDeprecated" type="checkbox" class="check-input" :disabled="submitting" />
              <span class="check-label">Deprecated</span>
            </label>
          </div>
        </div>

        <!-- Footer actions -->
        <div class="dialog-actions">
          <button class="btn btn--ghost" :disabled="submitting" @click="$emit('close')">Cancel</button>
          <button class="btn btn--primary" :disabled="submitting" @click="submit">
            <span v-if="submitting" class="btn-spinner"></span>
            {{ submitting ? 'Creating...' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
// @aif — Migrated from Apollo GraphQL `CREATE_CLASS_MUTATION` to REST
// `POST /api/v1/ontologies/:id/classes` per ADR-DES.API.rest-graphql-mutation-boundary.md.
import { createClass } from "@/api/ontology";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { computed, ref } from "vue";

const props = defineProps<{ open: boolean; ontologyId: string }>();
const emit = defineEmits<{ close: []; created: [className: string] }>();

const { addError } = useErrorPresentation();

const className = ref("");
const parentClass = ref("owl:Thing");
const description = ref("");
const isAbstract = ref(false);
const isDeprecated = ref(false);
const submitting = ref(false);
const validationError = ref<{ field: string; message: string } | null>(null);

/** Preview of the auto-generated IRI based on the class name. */
const classIriPreview = computed(() => {
	const trimmed = className.value.trim();
	return trimmed
		? `https://vedo-hub.ru/ontologies/${props.ontologyId}/${trimmed.replace(/\s+/g, "_")}`
		: "";
});

function onOverlayClick() {
	// modal — do not close on overlay click
}

async function submit(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "CreateClass.submitted",
			className: className.value,
			ontologyId: props.ontologyId,
			isAbstract: isAbstract.value,
			isDeprecated: isDeprecated.value,
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
	parentClass.value = "owl:Thing";
	description.value = "";
	isAbstract.value = false;
	isDeprecated.value = false;
	validationError.value = null;
}
</script>

<style scoped>
/* ── Overlay ──────────────────────────────────────────────────────────────── */

.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  z-index: var(--z-modal);
}

/* ── Card ─────────────────────────────────────────────────────────────────── */

.dialog-card {
  position: relative;
  width: min(calc(100vw - 32px), 550px);
  max-height: 90vh;
  overflow-y: auto;
  background: var(--surface);
  color: var(--foreground);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  font-family: var(--font-family-mono);
}

/* ── Close button ─────────────────────────────────────────────────────────── */

.dialog-close {
  position: absolute;
  top: 12px;
  right: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.dialog-close:hover {
  background: var(--surface-variant);
  color: var(--foreground);
}

/* ── Header ───────────────────────────────────────────────────────────────── */

.dialog-header {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  font-family: var(--font-family-mono);
  color: var(--foreground);
  line-height: var(--line-height-tight);
}

/* ── Form ─────────────────────────────────────────────────────────────────── */

.dialog-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

/* ── Field (label + input + hint/error) ───────────────────────────────────── */

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ff-label {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--foreground);
}

.ff-input {
  display: flex;
  align-items: center;
  height: 36px;
  padding: 4px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface);
  color: var(--foreground);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  line-height: 1.4;
  outline: none;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.ff-input::placeholder {
  color: var(--text-muted);
}

.ff-input:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-muted);
}

.ff-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ff-input--disabled {
  background: var(--surface-variant);
  color: var(--text-muted);
}

.ff-placeholder {
  color: var(--text-muted);
  font-size: var(--font-size-sm);
}

/* ── Textarea ─────────────────────────────────────────────────────────────── */

.ff-textarea {
  height: 80px;
  padding: 10px 12px;
  resize: vertical;
  align-items: flex-start;
}

/* ── Hint / Error text ────────────────────────────────────────────────────── */

.ff-hint {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  line-height: var(--line-height-tight);
}

.ff-error {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--error);
  line-height: var(--line-height-tight);
}

/* ── Select ───────────────────────────────────────────────────────────────── */

.ff-select-wrapper {
  position: relative;
  display: flex;
}

.ff-select {
  width: 100%;
  height: 36px;
  padding: 4px 32px 4px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface);
  color: var(--foreground);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  outline: none;
  appearance: none;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.ff-select:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-muted);
}

.ff-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ff-select-chevron {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  color: var(--text-muted);
}

/* ── Checkbox group ───────────────────────────────────────────────────────── */

.form-checks {
  display: flex;
  gap: var(--space-6);
}

.check-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.check-item--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.check-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.check-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface);
  color: var(--foreground);
  flex-shrink: 0;
}

.check-input:checked + .check-box,
.check-item:has(.check-input:checked) .check-box {
  background: var(--primary);
  border-color: var(--primary);
  color: var(--primary-foreground);
}

.check-label {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-normal);
  color: var(--foreground);
}

/* ── Actions ──────────────────────────────────────────────────────────────── */

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  padding-top: var(--space-6);
}

/* ── Buttons ──────────────────────────────────────────────────────────────── */

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: var(--space-2) var(--space-4);
  border: none;
  border-radius: var(--radius-md);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  line-height: var(--line-height-normal);
  cursor: pointer;
  transition: background var(--transition-fast), opacity var(--transition-fast);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--ghost {
  background: transparent;
  color: var(--foreground);
  border: 1px solid var(--border-default);
}

.btn--ghost:hover:not(:disabled) {
  background: var(--surface-variant);
}

.btn--primary {
  background: var(--primary);
  color: var(--primary-foreground);
}

.btn--primary:hover:not(:disabled) {
  background: var(--primary-hover);
}

/* ── Spinner ──────────────────────────────────────────────────────────────── */

.btn-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
