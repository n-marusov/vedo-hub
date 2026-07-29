<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism DocumentUploader — drag & drop zone with progress, format validation -->
<template>
  <div
    :class="['uploader', {
      'uploader--dragover': isDragOver,
      'uploader--disabled': disabled,
      'uploader--error': localError !== null
    }]"
    role="region"
    aria-label="Document upload zone"
    @dragenter.prevent="onDragEnter"
    @dragover.prevent="onDragOver"
    @dragleave.prevent="onDragLeave"
    @drop.prevent="onDrop"
  >
    <input
      ref="fileInputRef"
      type="file"
      :accept="acceptedExtensions"
      :multiple="mode === 'batch'"
      class="uploader__input"
      aria-hidden="true"
      tabindex="-1"
      @change="onFileSelected"
    />

    <!-- Idle state — drop zone -->
    <div v-if="state === 'idle'" class="uploader__dropzone" @click="onBrowseClick">
      <Upload :size="32" class="uploader__icon" />
      <p class="uploader__title">
        {{ mode === 'batch' ? 'Drop files here' : 'Drop a document here' }}
      </p>
      <p class="uploader__subtitle">
        or <button type="button" class="uploader__browse-link" @click.stop="onBrowseClick">browse</button> to upload
      </p>
      <p class="uploader__formats">
        Supported: {{ formatList }}
        <span v-if="maxSize"> (max {{ maxSize }} MB)</span>
      </p>
    </div>

    <!-- Uploading state — progress bar -->
    <div v-if="state === 'uploading'" class="uploader__progress">
      <div class="uploader__file-info">
        <FileText :size="20" />
        <span class="uploader__file-name">{{ currentFileName }}</span>
        <span class="uploader__file-size">{{ currentFileSize }}</span>
      </div>
      <div class="uploader__progress-bar-track">
        <div
          class="uploader__progress-bar-fill"
          :style="{ width: `${uploadProgress}%` }"
          role="progressbar"
          :aria-valuenow="uploadProgress"
          aria-valuemin="0"
          aria-valuemax="100"
        ></div>
      </div>
      <span class="uploader__progress-text">{{ uploadProgress }}% — Extracting...</span>
    </div>

    <!-- Error state -->
    <div v-if="state === 'error'" class="uploader__error">
      <AlertCircle :size="24" class="uploader__error-icon" />
      <p class="uploader__error-title">Upload failed</p>
      <p class="uploader__error-message">{{ localError?.message }}</p>
      <div class="uploader__error-actions">
        <button class="uploader__retry-btn" type="button" @click="retryUpload">
          Retry
        </button>
        <button class="uploader__dismiss-btn" type="button" @click="dismissError">
          Dismiss
        </button>
      </div>
    </div>

    <!-- Success state — completed -->
    <div v-if="state === 'success'" class="uploader__success">
      <CheckCircle :size="24" class="uploader__success-icon" />
      <p class="uploader__success-text">Extracted {{ stepCount }} steps from {{ currentFileName }}</p>
      <button class="uploader__clear-btn" type="button" @click="resetUpload">
        Upload another
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AlertCircle, CheckCircle, FileText, Upload } from "@lucide/vue";
import { computed, ref } from "vue";
import {
	ALLOWED_FORMATS,
	FORMAT_LABELS,
	uploadDocument,
} from "../../api/extraction";
import type {
	ExtractionError,
	ExtractionPreview,
} from "../../types/extraction";

const props = defineProps<{
	ontologyId: string;
	maxFileSizeMb?: number;
	allowedFormats?: string[];
	mode?: "single" | "batch";
}>();

const emit = defineEmits<{
	"upload-complete": [result: ExtractionPreview];
	"upload-error": [error: ExtractionError];
	"upload-progress": [progress: number];
}>();

// ── State ──────────────────────────────────────────────────────────────────

type UploadState = "idle" | "uploading" | "success" | "error";

const state = ref<UploadState>("idle");
const isDragOver = ref(false);
const uploadProgress = ref(0);
const currentFileName = ref("");
const currentFileSize = ref("");
const localError = ref<ExtractionError | null>(null);
const pendingFile = ref<File | null>(null);
const lastResult = ref<ExtractionPreview | null>(null);
const stepCount = ref(0);
const fileInputRef = ref<HTMLInputElement | null>(null);

const maxSize = computed(() => props.maxFileSizeMb ?? 20);
const formats = computed(() => props.allowedFormats ?? ALLOWED_FORMATS);
const acceptedExtensions = computed(() => formats.value.join(","));
const formatList = computed(() => {
	return formats.value.map((f) => FORMAT_LABELS[f] || f).join(", ");
});
const disabled = computed(() => state.value === "uploading");

// ── File validation ─────────────────────────────────────────────────────────

function validateFile(file: File): string | null {
	const ext = `.${file.name.split(".").pop()?.toLowerCase()}`;

	if (!formats.value.includes(ext)) {
		return `Unsupported format "${ext}". Allowed: ${formatList.value}`;
	}

	const maxBytes = maxSize.value * 1024 * 1024;
	if (file.size > maxBytes) {
		return `File too large (${formatFileSize(file.size)}). Maximum: ${maxSize.value} MB`;
	}

	if (file.size === 0) {
		return "File is empty";
	}

	return null;
}

function formatFileSize(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// ── Drag & drop handlers ────────────────────────────────────────────────────

let dragCounter = 0;

function onDragEnter() {
	if (disabled.value) return;
	dragCounter++;
	isDragOver.value = true;
}

function onDragOver() {
	if (disabled.value) return;
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

	if (disabled.value) return;

	const files = event.dataTransfer?.files;
	if (!files || files.length === 0) return;

	handleFile(files[0]);
}

// ── Browse handlers ─────────────────────────────────────────────────────────

function onBrowseClick() {
	if (disabled.value) return;
	fileInputRef.value?.click();
}

function onFileSelected(event: Event) {
	const input = event.target as HTMLInputElement;
	const file = input.files?.[0];
	if (file) {
		handleFile(file);
	}
	input.value = "";
}

// ── Upload logic ────────────────────────────────────────────────────────────

async function handleFile(file: File) {
	const validationError = validateFile(file);
	if (validationError) {
		const err: ExtractionError = {
			code: "VALIDATION_FAILED",
			message: validationError,
		};
		console.warn("[DocumentUploader] validation failed", {
			fileName: file.name,
			fileSize: file.size,
			error: validationError,
		});
		localError.value = err;
		state.value = "error";
		emit("upload-error", err);
		return;
	}

	pendingFile.value = file;
	currentFileName.value = file.name;
	currentFileSize.value = formatFileSize(file.size);
	state.value = "uploading";
	uploadProgress.value = 0;

	try {
		console.debug("[DocumentUploader] starting upload", {
			fileName: file.name,
			fileSize: file.size,
		});

		const preview = await uploadDocument({
			ontologyId: props.ontologyId,
			file,
			onProgress: (pct) => {
				uploadProgress.value = pct;
				emit("upload-progress", pct);
			},
		});

		console.info("[DocumentUploader] upload complete", {
			fileName: file.name,
			steps: preview.steps.length,
		});

		lastResult.value = preview;
		stepCount.value = preview.totalSteps;
		state.value = "success";
		emit("upload-complete", preview);
	} catch (error) {
		const err = error as ExtractionError;
		localError.value = err;
		state.value = "error";
		emit("upload-error", err);
	}
}

function retryUpload() {
	if (pendingFile.value) {
		handleFile(pendingFile.value);
	}
}

function dismissError() {
	localError.value = null;
	state.value = "idle";
	pendingFile.value = null;
	uploadProgress.value = 0;
}

function resetUpload() {
	lastResult.value = null;
	stepCount.value = 0;
	currentFileName.value = "";
	currentFileSize.value = "";
	uploadProgress.value = 0;
	state.value = "idle";
}
</script>

<style scoped>
.uploader {
  border: 2px dashed var(--border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-6);
  transition: all var(--transition-fast);
  background: var(--surface-primary);
  position: relative;
}

.uploader--dragover {
  border-color: var(--primary);
  background: rgba(16, 185, 129, 0.05);
}

.uploader--disabled {
  opacity: 0.6;
  pointer-events: none;
}

.uploader--error {
  border-color: var(--status-error);
  background: rgba(239, 68, 68, 0.05);
}

.uploader__input {
  display: none;
}

.uploader__dropzone {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-2);
  cursor: pointer;
  padding: var(--spacing-4) 0;
}

.uploader__icon {
  color: var(--text-muted);
  margin-bottom: var(--spacing-2);
}

.uploader--dragover .uploader__icon {
  color: var(--primary);
}

.uploader__title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  margin: 0;
}

.uploader__subtitle {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  margin: 0;
}

.uploader__browse-link {
  background: none;
  border: none;
  color: var(--primary);
  cursor: pointer;
  font-size: inherit;
  text-decoration: underline;
  padding: 0;
}

.uploader__browse-link:hover {
  color: var(--primary-hover);
}

.uploader__formats {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  margin: var(--spacing-2) 0 0;
}

.uploader__progress {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  padding: var(--spacing-2) 0;
}

.uploader__file-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.uploader__file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: var(--font-weight-medium);
}

.uploader__file-size {
  color: var(--text-muted);
  font-size: var(--font-size-xs);
}

.uploader__progress-bar-track {
  width: 100%;
  height: 8px;
  background: var(--surface-tertiary);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.uploader__progress-bar-fill {
  height: 100%;
  background: var(--primary);
  border-radius: var(--radius-full);
  transition: width 0.3s ease;
}

.uploader__progress-text {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  text-align: center;
}

.uploader__error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) 0;
}

.uploader__error-icon {
  color: var(--status-error);
}

.uploader__error-title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  margin: 0;
}

.uploader__error-message {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  margin: 0;
  text-align: center;
}

.uploader__error-actions {
  display: flex;
  gap: var(--spacing-2);
  margin-top: var(--spacing-2);
}

.uploader__retry-btn {
  padding: var(--spacing-1) var(--spacing-3);
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--surface-primary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: var(--font-size-sm);
}

.uploader__retry-btn:hover {
  background: var(--surface-secondary);
}

.uploader__dismiss-btn {
  padding: var(--spacing-1) var(--spacing-3);
  border-radius: var(--radius-md);
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: var(--font-size-sm);
}

.uploader__dismiss-btn:hover {
  color: var(--text-primary);
}

.uploader__success {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) 0;
}

.uploader__success-icon {
  color: var(--status-success);
}

.uploader__success-text {
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  margin: 0;
}

.uploader__clear-btn {
  padding: var(--spacing-1) var(--spacing-3);
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--primary);
  cursor: pointer;
  font-size: var(--font-size-sm);
}

.uploader__clear-btn:hover {
  background: rgba(16, 185, 129, 0.1);
}
</style>
