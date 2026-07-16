<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism ApplySequenceButton — apply button with confirmation modal and progress -->
<template>
  <div class="apply-flow">
    <!-- Trigger button -->
    <button
      class="apply-btn"
      type="button"
      :disabled="disabled || stepCount === 0"
      :title="disabled ? 'Upload a document first' : `Import ${stepCount} entities`"
      @click="onApplyClick"
    >
      <Upload :size="16" />
      {{ buttonLabel }}
    </button>

    <!-- Progress modal -->
    <ApplyProgressModal
      :visible="modalVisible"
      :status="progress.status"
      :total-steps="progress.totalSteps"
      :completed-steps="progress.completedSteps"
      :percentage="percentage"
      :current-step-label="progress.currentStep"
      :error-message="progress.errorMessage"
      :commit-message="commitMessage"
      :steps="selectedSteps"
      :apply-result="applyResult"
      @confirm="onConfirm"
      @cancel="onCancel"
      @close="onClose"
      @retry="onRetry"
      @update:commit-message="commitMessage = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { Upload } from "lucide-vue-next";
import { computed, ref } from "vue";
import { useApplySequence } from "../../composables/useApplySequence";
import type { SequenceStep } from "../../types/extraction";
import ApplyProgressModal from "./ApplyProgressModal.vue";

const props = defineProps<{
	ontologyId: string;
	steps: SequenceStep[];
	disabled?: boolean;
}>();

const emit = defineEmits<{
	"apply-success": [result: { commitId?: string; commitUrl?: string }];
	"apply-error": [error: string];
}>();

// ── Apply composable ────────────────────────────────────────────────────────

const {
	progress,
	applyResult,
	commitMessage,
	isApplying,
	isSuccess,
	isError,
	percentage,
	requestConfirm,
	executeApply,
	cancelApply,
	reset,
} = useApplySequence();

// ── Local state ─────────────────────────────────────────────────────────────

const modalVisible = ref(false);
const selectedSteps = ref<SequenceStep[]>([]);

// ── Computed ────────────────────────────────────────────────────────────────

const stepCount = computed(() => props.steps.filter((s) => s.included).length);
const buttonLabel = computed(() => {
	if (stepCount.value === 0) return "Apply Import";
	return `Apply Import (${stepCount.value})`;
});

// ── Handlers ────────────────────────────────────────────────────────────────

function onApplyClick() {
	selectedSteps.value = props.steps;
	requestConfirm(props.steps);
	modalVisible.value = true;
}

async function onConfirm() {
	await executeApply(
		props.ontologyId,
		selectedSteps.value,
		commitMessage.value,
	);

	if (isSuccess.value && applyResult.value) {
		console.info("[ApplySequenceButton] apply success", {
			commitId: applyResult.value.commitId,
		});
		emit("apply-success", {
			commitId: applyResult.value.commitId,
			commitUrl: applyResult.value.commitUrl,
		});
	} else if (isError.value) {
		const errMsg = progress.value.errorMessage || "Unknown error";
		console.error("[ApplySequenceButton] apply error", { error: errMsg });
		emit("apply-error", errMsg);
	}
}

function onCancel() {
	if (isApplying.value) {
		cancelApply();
	}
	modalVisible.value = false;
	reset();
}

function onClose() {
	modalVisible.value = false;
	reset();
}

function onRetry() {
	// Go back to confirming state
	requestConfirm(selectedSteps.value);
}
</script>

<style scoped>
.apply-flow {
  display: inline-flex;
}

.apply-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-4);
  border-radius: var(--radius-md);
  border: 1px solid var(--primary);
  background: var(--primary);
  color: var(--text-inverse);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.apply-btn:hover:not(:disabled) {
  background: var(--primary-hover);
}

.apply-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
