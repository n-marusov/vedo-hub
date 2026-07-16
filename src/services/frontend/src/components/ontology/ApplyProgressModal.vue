<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism ApplyProgressModal — workflow: Confirm → Progress → Success/Error -->
<template>
  <Teleport to="body">
    <div v-if="visible" class="modal-overlay" @click.self="onOverlayClick" role="dialog" aria-modal="true" aria-label="Apply import">
      <div class="modal">
        <!-- Confirm step -->
        <div v-if="status === 'confirming'" class="modal__body">
          <h3 class="modal__title">Confirm Import</h3>
          <p class="modal__desc">
            This will import <strong>{{ totalSteps }}</strong> ontology {{ totalSteps === 1 ? 'entity' : 'entities' }}
            from the document extraction.
          </p>

          <div class="modal__summary">
            <div class="modal__summary-row">
              <span>Classes</span>
              <span class="modal__summary-value">{{ classCount }}</span>
            </div>
            <div class="modal__summary-row">
              <span>Properties</span>
              <span class="modal__summary-value">{{ propertyCount }}</span>
            </div>
            <div class="modal__summary-row">
              <span>Individuals</span>
              <span class="modal__summary-value">{{ individualCount }}</span>
            </div>
            <div class="modal__summary-row">
              <span>Updates / other</span>
              <span class="modal__summary-value">{{ otherCount }}</span>
            </div>
          </div>

          <div class="modal__field">
            <label class="modal__label" for="commit-msg">Commit message</label>
            <input
              id="commit-msg"
              :value="commitMessage"
              class="modal__input"
              placeholder="Describe the import..."
              @input="$emit('update:commitMessage', ($event.target as HTMLInputElement).value)"
            />
          </div>

          <div class="modal__footer">
            <button class="modal__btn modal__btn--secondary" type="button" @click="$emit('cancel')">
              Cancel
            </button>
            <button class="modal__btn modal__btn--primary" type="button" :disabled="!commitMessage.trim()" @click="$emit('confirm')">
              Import {{ totalSteps }} {{ totalSteps === 1 ? 'entity' : 'entities' }}
            </button>
          </div>
        </div>

        <!-- Applying step -->
        <div v-if="status === 'applying'" class="modal__body">
          <h3 class="modal__title">Importing...</h3>
          <div class="modal__progress">
            <div class="modal__progress-track">
              <div
                class="modal__progress-fill"
                :style="{ width: `${percentage}%` }"
                role="progressbar"
                :aria-valuenow="completedSteps"
                aria-valuemin="0"
                :aria-valuemax="totalSteps"
              ></div>
            </div>
            <span class="modal__progress-text">
              {{ completedSteps }} of {{ totalSteps }} steps
            </span>
          </div>
          <p class="modal__step-label">{{ currentStepLabel }}</p>
        </div>

        <!-- Success step -->
        <div v-if="status === 'success'" class="modal__body">
          <div class="modal__success-icon">
            <CheckCircle :size="48" />
          </div>
          <h3 class="modal__title modal__title--success">Import Successful</h3>
          <p class="modal__desc">
            Imported <strong>{{ applyResult?.appliedCount || totalSteps }}</strong> {{ (applyResult?.appliedCount || totalSteps) === 1 ? 'entity' : 'entities' }}.
          </p>
          <div v-if="applyResult?.commitId" class="modal__commit-link">
            <a
              :href="applyResult.commitUrl || '#'"
              class="modal__commit-hash"
              target="_blank"
              rel="noopener noreferrer"
            >
              <GitCommit :size="14" />
              Commit {{ applyResult.commitId.substring(0, 7) }}
            </a>
          </div>
          <div class="modal__footer modal__footer--center">
            <button class="modal__btn modal__btn--primary" type="button" @click="$emit('close')">
              Done
            </button>
          </div>
        </div>

        <!-- Error step -->
        <div v-if="status === 'error'" class="modal__body">
          <div class="modal__error-icon">
            <AlertCircle :size="48" />
          </div>
          <h3 class="modal__title modal__title--error">Import Failed</h3>
          <p class="modal__desc">{{ errorMessage || 'An unknown error occurred during import.' }}</p>
          <div v-if="applyResult?.appliedCount && applyResult.appliedCount > 0" class="modal__partial">
            <p class="modal__partial-text">
              Partially applied: {{ applyResult.appliedCount }} of {{ totalSteps }} steps completed.
            </p>
          </div>
          <div class="modal__footer modal__footer--center">
            <button class="modal__btn modal__btn--secondary" type="button" @click="$emit('close')">
              Close
            </button>
            <button class="modal__btn modal__btn--primary" type="button" @click="$emit('retry')">
              Retry
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { AlertCircle, CheckCircle, GitCommit } from "lucide-vue-next";
import { computed } from "vue";
import type {
	ApplyProgress,
	ApplyResult,
	SequenceStep,
} from "../../types/extraction";

const props = defineProps<{
	visible: boolean;
	status: ApplyProgress["status"];
	totalSteps: number;
	completedSteps: number;
	percentage: number;
	currentStepLabel: string;
	errorMessage?: string;
	commitMessage: string;
	steps?: SequenceStep[];
	applyResult?: ApplyResult | null;
}>();

const emit = defineEmits<{
	confirm: [];
	cancel: [];
	close: [];
	retry: [];
	"update:commitMessage": [value: string];
}>();

// ── Summary computation ─────────────────────────────────────────────────────

const classCount = computed(
	() =>
		props.steps?.filter((s) => s.included && s.operation === "CREATE_CLASS")
			.length ?? 0,
);
const propertyCount = computed(
	() =>
		props.steps?.filter((s) => s.included && s.operation === "CREATE_PROPERTY")
			.length ?? 0,
);
const individualCount = computed(
	() =>
		props.steps?.filter(
			(s) => s.included && s.operation === "CREATE_INDIVIDUAL",
		).length ?? 0,
);
const otherCount = computed(
	() =>
		props.steps?.filter(
			(s) =>
				s.included &&
				!["CREATE_CLASS", "CREATE_PROPERTY", "CREATE_INDIVIDUAL"].includes(
					s.operation,
				),
		).length ?? 0,
);

function onOverlayClick() {
	// Only allow close on idle/success/error states — not during confirming/applying
	if (props.status === "success" || props.status === "error") {
		emit("close");
	}
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal, 1000);
}

.modal {
  width: 480px;
  max-width: 90vw;
  background: var(--surface-primary, #101010);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl, 12px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

.modal__body {
  padding: var(--spacing-6, 24px);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4, 16px);
}

.modal__title {
  font-size: var(--font-size-lg, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #fafafa);
  margin: 0;
  text-align: center;
}

.modal__title--success {
  color: var(--success, #10b981);
}

.modal__title--error {
  color: var(--status-error, #ef4444);
}

.modal__desc {
  font-size: var(--font-size-sm, 13px);
  color: var(--text-secondary, #6b7280);
  margin: 0;
  text-align: center;
}

.modal__summary {
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: var(--spacing-3, 12px);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2, 8px);
}

.modal__summary-row {
  display: flex;
  justify-content: space-between;
  font-size: var(--font-size-sm, 13px);
  color: var(--text-secondary, #6b7280);
}

.modal__summary-value {
  font-weight: var(--font-weight-medium, 500);
  color: var(--text-primary, #fafafa);
}

.modal__field {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1, 4px);
}

.modal__label {
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  color: var(--text-secondary, #6b7280);
}

.modal__input {
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  font-size: var(--font-size-sm, 13px);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-primary, #101010);
  color: var(--text-primary, #fafafa);
  outline: none;
}

.modal__input:focus {
  border-color: var(--primary, #10b981);
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
}

.modal__progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-2, 8px);
}

.modal__progress-track {
  width: 100%;
  height: 10px;
  background: var(--surface-tertiary, #1a1a1a);
  border-radius: var(--radius-full, 999px);
  overflow: hidden;
}

.modal__progress-fill {
  height: 100%;
  background: var(--primary, #10b981);
  border-radius: var(--radius-full, 999px);
  transition: width 0.3s ease;
}

.modal__progress-text {
  font-size: var(--font-size-sm, 13px);
  color: var(--text-muted, #64748b);
}

.modal__step-label {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #64748b);
  text-align: center;
  margin: 0;
}

.modal__success-icon,
.modal__error-icon {
  display: flex;
  justify-content: center;
  color: var(--success, #10b981);
}

.modal__error-icon {
  color: var(--status-error, #ef4444);
}

.modal__commit-link {
  display: flex;
  justify-content: center;
}

.modal__commit-hash {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1, 4px);
  padding: var(--spacing-1, 4px) var(--spacing-3, 12px);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  font-family: 'IBM Plex Mono', monospace;
  font-size: var(--font-size-sm, 13px);
  color: var(--primary, #10b981);
  text-decoration: none;
}

.modal__commit-hash:hover {
  background: rgba(16, 185, 129, 0.1);
}

.modal__partial {
  border: 1px solid var(--warning);
  border-radius: var(--radius-md, 8px);
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  background: rgba(245, 158, 11, 0.05);
}

.modal__partial-text {
  font-size: var(--font-size-sm, 13px);
  color: var(--warning, #f59e0b);
  margin: 0;
  text-align: center;
}

.modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-2, 8px);
  margin-top: var(--spacing-2, 8px);
}

.modal__footer--center {
  justify-content: center;
}

.modal__btn {
  padding: var(--spacing-2, 8px) var(--spacing-4, 16px);
  border-radius: var(--radius-md, 8px);
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  cursor: pointer;
  transition: all var(--transition-fast, 0.15s);
  border: none;
}

.modal__btn--primary {
  background: var(--primary, #10b981);
  color: var(--text-inverse, #0a0a0a);
}

.modal__btn--primary:hover:not(:disabled) {
  background: var(--primary-hover, #34d399);
}

.modal__btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal__btn--secondary {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text-secondary, #6b7280);
}

.modal__btn--secondary:hover {
  background: var(--surface-secondary, #141414);
}
</style>
