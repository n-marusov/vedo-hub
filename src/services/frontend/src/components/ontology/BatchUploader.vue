<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism BatchUploader — multi-file upload with per-file progress and conflict resolution -->
<template>
  <div class="batch-uploader" role="region" aria-label="Batch document upload">
    <!-- File drop zone -->
    <div
      :class="['batch-uploader__dropzone', { 'batch-uploader__dropzone--active': isDragOver, 'batch-uploader__dropzone--disabled': isProcessing }]"
      @dragenter.prevent="onDragEnter"
      @dragover.prevent="onDragOver"
      @dragleave.prevent="onDragLeave"
      @drop.prevent="onDrop"
    >
      <Upload :size="28" class="batch-uploader__drop-icon" />
      <p class="batch-uploader__drop-text">
        Drop documents here or <button type="button" class="batch-uploader__browse-link" @click="onBrowseClick" :disabled="isProcessing">browse</button>
      </p>
      <p class="batch-uploader__drop-hint">
        Up to 10 files ({{ ALLOWED_FORMATS.join(', ') }})
      </p>
      <input
        ref="fileInputRef"
        type="file"
        multiple
        :accept="ALLOWED_FORMATS.join(',')"
        class="batch-uploader__input"
        @change="onFileSelected"
      />
    </div>

    <!-- File list -->
    <div v-if="files.length > 0" class="batch-uploader__files">
      <div
        v-for="fileState in files"
        :key="fileState.fileName"
        :class="['batch-uploader__file', `batch-uploader__file--${fileState.status}`]"
      >
        <FileText :size="16" class="batch-uploader__file-icon" />
        <span class="batch-uploader__file-name">{{ fileState.fileName }}</span>

        <!-- Progress bar (uploading) -->
        <div v-if="fileState.status === 'uploading'" class="batch-uploader__file-progress">
          <div class="batch-uploader__file-progress-track">
            <div class="batch-uploader__file-progress-fill" :style="{ width: `${fileState.progress}%` }"></div>
          </div>
          <span class="batch-uploader__file-progress-text">{{ fileState.progress }}%</span>
        </div>

        <!-- Status indicators -->
        <span v-if="fileState.status === 'pending'" class="batch-uploader__file-status batch-uploader__file-status--pending">Pending</span>
        <span v-if="fileState.status === 'success'" class="batch-uploader__file-status batch-uploader__file-status--success">
          {{ fileState.preview?.steps.length }} steps
        </span>
        <span v-if="fileState.status === 'error'" class="batch-uploader__file-status batch-uploader__file-status--error" :title="fileState.error">
          Failed
        </span>

        <!-- Remove button -->
        <button
          v-if="fileState.status === 'pending' || fileState.status === 'error'"
          class="batch-uploader__file-remove"
          type="button"
          :aria-label="`Remove ${fileState.fileName}`"
          @click="removeFile(fileState.fileName)"
        >
          <X :size="14" />
        </button>
      </div>
    </div>

    <!-- Upload controls -->
    <div v-if="files.length > 0 && !allDone" class="batch-uploader__actions">
      <span class="batch-uploader__action-count">
        {{ pendingCount }} pending · {{ successCount }} success · {{ failedCount }} failed
      </span>
      <div class="batch-uploader__action-buttons">
        <button
          v-if="isProcessing"
          class="batch-uploader__cancel-btn"
          type="button"
          @click="cancelUpload"
        >
          Cancel
        </button>
        <button
          v-else
          class="batch-uploader__upload-btn"
          type="button"
          :disabled="pendingCount === 0 && failedCount === 0"
          @click="startUpload"
        >
          Upload {{ pendingCount > 0 ? `(${pendingCount})` : '' }}
        </button>
      </div>
    </div>

    <!-- Retry failed -->
    <div v-if="hasFailedFiles && !isProcessing" class="batch-uploader__retry">
      <button class="batch-uploader__retry-btn" type="button" @click="retryFailed">
        Retry {{ failedCount }} failed {{ failedCount === 1 ? 'file' : 'files' }}
      </button>
    </div>

    <!-- Merged preview (after all uploads complete) -->
    <div v-if="allDone && mergedSteps.length > 0" class="batch-uploader__preview">
      <SequencePreview
        :steps="mergedSteps"
        :ontology-id="ontologyId"
        :source-files="sourceFiles"
        @apply="onApplyMerged"
        @cancel="$emit('reset')"
      />
    </div>

    <!-- Conflict resolver modal -->
    <ConflictResolver
      :visible="showConflicts"
      :conflicts="conflicts"
      @resolve="onResolveConflict"
      @close="showConflicts = false"
      @apply="onConflictsApplied"
    />
  </div>
</template>

<script setup lang="ts">
import { FileText, Upload, X } from "lucide-vue-next";
import { computed, ref } from "vue";
import { ALLOWED_FORMATS } from "../../api/extraction";
import { useBatchUpload } from "../../composables/useBatchUpload";
import type { ConflictResolution, SequenceStep } from "../../types/extraction";
import ConflictResolver from "./ConflictResolver.vue";
import SequencePreview from "./SequencePreview.vue";

const props = defineProps<{
	ontologyId: string;
}>();

const emit = defineEmits<{
	"batch-complete": [result: { steps: SequenceStep[] }];
	"batch-error": [error: string];
	reset: [];
}>();

// ── Batch composable ────────────────────────────────────────────────────────

const {
	files,
	mergedSteps,
	conflicts,
	isProcessing,
	hasConflicts,
	allDone,
	pendingCount,
	successCount,
	failedCount,
	hasFailedFiles,
	addFiles,
	removeFile,
	uploadAll,
	cancelUpload: cancelBatchUpload,
	resolveConflict,
} = useBatchUpload();

// ── Component state ─────────────────────────────────────────────────────────

const isDragOver = ref(false);
const showConflicts = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);
let dragCounter = 0;

// ── Computed ────────────────────────────────────────────────────────────────

const sourceFiles = computed(() =>
	files.value.filter((f) => f.status === "success").map((f) => f.fileName),
);

// ── Drag & drop ─────────────────────────────────────────────────────────────

function onDragEnter() {
	if (isProcessing.value) return;
	dragCounter++;
	isDragOver.value = true;
}

function onDragOver() {
	if (isProcessing.value) return;
	isDragOver.value = true;
}

function onDragLeave() {
	dragCounter--;
	if (dragCounter <= 0) {
		dragCounter = 0;
		isDragOver.value = false;
	}
}

function onDrop(event: DragEvent) {
	dragCounter = 0;
	isDragOver.value = false;
	if (isProcessing.value) return;

	const droppedFiles = event.dataTransfer?.files;
	if (!droppedFiles || droppedFiles.length === 0) return;

	handleFiles(Array.from(droppedFiles));
}

// ── Browse ──────────────────────────────────────────────────────────────────

function onBrowseClick() {
	if (isProcessing.value) return;
	fileInputRef.value?.click();
}

function onFileSelected(event: Event) {
	const input = event.target as HTMLInputElement;
	const selectedFiles = input.files;
	if (selectedFiles) {
		handleFiles(Array.from(selectedFiles));
	}
	input.value = "";
}

// ── File handling ───────────────────────────────────────────────────────────

function handleFiles(newFiles: File[]) {
	// Validate total count
	const totalAfterAdd = files.value.length + newFiles.length;
	if (totalAfterAdd > 10) {
		console.warn("[BatchUploader] too many files, max 10");
		return;
	}

	addFiles(newFiles);
}

async function startUpload() {
	console.info("[BatchUploader] starting upload", {
		pendingCount: pendingCount.value,
	});

	await uploadAll(props.ontologyId);

	if (hasConflicts.value) {
		showConflicts.value = true;
	} else if (allDone.value && mergedSteps.value.length > 0) {
		console.info("[BatchUploader] all files processed", {
			totalSteps: mergedSteps.value.length,
		});
		emit("batch-complete", { steps: mergedSteps.value });
	}
}

function cancelUpload() {
	cancelBatchUpload();
}

function retryFailed() {
	startUpload();
}

// ── Conflict resolution ─────────────────────────────────────────────────────

function onResolveConflict(
	index: number,
	resolution: ConflictResolution["resolution"],
	customValue?: string,
) {
	resolveConflict(index, resolution, customValue);
}

function onConflictsApplied() {
	showConflicts.value = false;

	console.info("[BatchUploader] conflicts resolved, emitting merged steps", {
		totalSteps: mergedSteps.value.length,
	});

	emit("batch-complete", { steps: mergedSteps.value });
}

// ── Apply merged steps ──────────────────────────────────────────────────────

async function onApplyMerged(steps: SequenceStep[]) {
	console.info("[BatchUploader] apply merged steps", {
		stepCount: steps.length,
	});
	emit("batch-complete", { steps });
}
</script>

<style scoped>
.batch-uploader {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4, 16px);
}

.batch-uploader__dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius-lg, 12px);
  padding: var(--spacing-6, 24px);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-2, 8px);
  cursor: pointer;
  transition: all var(--transition-fast, 0.15s);
  background: var(--surface-primary, #101010);
}

.batch-uploader__dropzone--active {
  border-color: var(--primary, #10b981);
  background: rgba(16, 185, 129, 0.05);
}

.batch-uploader__dropzone--disabled {
  opacity: 0.5;
  pointer-events: none;
}

.batch-uploader__drop-icon {
  color: var(--text-muted, #64748b);
}

.batch-uploader__dropzone--active .batch-uploader__drop-icon {
  color: var(--primary, #10b981);
}

.batch-uploader__drop-text {
  font-size: var(--font-size-base, 14px);
  color: var(--text-primary, #fafafa);
  margin: 0;
}

.batch-uploader__browse-link {
  background: none;
  border: none;
  color: var(--primary, #10b981);
  cursor: pointer;
  text-decoration: underline;
  font-size: inherit;
  padding: 0;
}

.batch-uploader__drop-hint {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #64748b);
  margin: 0;
}

.batch-uploader__input {
  display: none;
}

.batch-uploader__files {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1, 4px);
}

.batch-uploader__file {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 8px);
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  font-size: var(--font-size-sm, 13px);
}

.batch-uploader__file--success {
  border-color: rgba(16, 185, 129, 0.3);
}

.batch-uploader__file--error {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.03);
}

.batch-uploader__file-icon {
  color: var(--text-muted, #64748b);
  flex-shrink: 0;
}

.batch-uploader__file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-primary, #fafafa);
}

.batch-uploader__file-progress {
  display: flex;
  align-items: center;
  gap: var(--spacing-1, 4px);
  width: 140px;
}

.batch-uploader__file-progress-track {
  flex: 1;
  height: 6px;
  background: var(--surface-tertiary, #1a1a1a);
  border-radius: var(--radius-full, 999px);
  overflow: hidden;
}

.batch-uploader__file-progress-fill {
  height: 100%;
  background: var(--primary, #10b981);
  border-radius: var(--radius-full, 999px);
  transition: width 0.3s ease;
}

.batch-uploader__file-progress-text {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #64748b);
  width: 36px;
  text-align: right;
}

.batch-uploader__file-status {
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-medium, 500);
}

.batch-uploader__file-status--pending {
  color: var(--text-muted, #64748b);
}

.batch-uploader__file-status--success {
  color: var(--success, #10b981);
}

.batch-uploader__file-status--error {
  color: var(--status-error, #ef4444);
}

.batch-uploader__file-remove {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--text-muted, #64748b);
  cursor: pointer;
  border-radius: var(--radius-sm, 6px);
  flex-shrink: 0;
}

.batch-uploader__file-remove:hover {
  background: var(--surface-secondary, #141414);
  color: var(--status-error, #ef4444);
}

.batch-uploader__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.batch-uploader__action-count {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #64748b);
}

.batch-uploader__action-buttons {
  display: flex;
  gap: var(--spacing-2, 8px);
}

.batch-uploader__upload-btn,
.batch-uploader__cancel-btn {
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  border-radius: var(--radius-md, 8px);
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  cursor: pointer;
  border: none;
}

.batch-uploader__upload-btn {
  background: var(--primary, #10b981);
  color: var(--text-inverse, #0a0a0a);
}

.batch-uploader__upload-btn:hover:not(:disabled) {
  background: var(--primary-hover, #34d399);
}

.batch-uploader__upload-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.batch-uploader__cancel-btn {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text-secondary, #6b7280);
}

.batch-uploader__cancel-btn:hover {
  background: var(--surface-secondary, #141414);
}

.batch-uploader__retry {
  display: flex;
  justify-content: center;
}

.batch-uploader__retry-btn {
  padding: var(--spacing-1, 4px) var(--spacing-3, 12px);
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--warning, #f59e0b);
  background: transparent;
  color: var(--warning, #f59e0b);
  font-size: var(--font-size-sm, 13px);
  cursor: pointer;
}

.batch-uploader__retry-btn:hover {
  background: rgba(245, 158, 11, 0.1);
}

.batch-uploader__preview {
  border-top: 1px solid var(--border);
  padding-top: var(--spacing-4, 16px);
}
</style>
