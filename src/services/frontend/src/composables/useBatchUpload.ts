// Composable for managing batch document upload and deduplication

import { computed, ref } from "vue";
import { batchUploadDocuments } from "../api/extraction";
import type {
	BatchUploadResult,
	ConflictResolution,
	ExtractionPreview,
	SequenceStep,
} from "../types/extraction";

export interface BatchFileState {
	file: File;
	fileName: string;
	progress: number;
	status: "pending" | "uploading" | "success" | "error";
	preview?: ExtractionPreview;
	error?: string;
}

export function useBatchUpload() {
	// ── State ────────────────────────────────────────────────────────────────

	const files = ref<BatchFileState[]>([]);
	const mergedSteps = ref<SequenceStep[]>([]);
	const conflicts = ref<ConflictResolution[]>([]);
	const isProcessing = ref(false);
	const abortController = ref<AbortController | null>(null);

	// ── Computed ─────────────────────────────────────────────────────────────

	const pendingCount = computed(
		() => files.value.filter((f) => f.status === "pending").length,
	);
	const uploadingCount = computed(
		() => files.value.filter((f) => f.status === "uploading").length,
	);
	const successCount = computed(
		() => files.value.filter((f) => f.status === "success").length,
	);
	const failedCount = computed(
		() => files.value.filter((f) => f.status === "error").length,
	);
	const completedCount = computed(() => successCount.value + failedCount.value);
	const totalCount = computed(() => files.value.length);
	const hasConflicts = computed(() => conflicts.value.length > 0);
	const allDone = computed(
		() => completedCount.value === totalCount.value && totalCount.value > 0,
	);
	const hasFailedFiles = computed(() => failedCount.value > 0);

	const batchResult = computed<BatchUploadResult>(() => ({
		completedFiles: successCount.value,
		failedFiles: failedCount.value,
		totalSteps: mergedSteps.value.length,
		previews: files.value
			.filter((f) => f.status === "success" && f.preview)
			.map(
				(f) => f.preview as import("../../types/extraction").ExtractionPreview,
			),
		failed: files.value
			.filter((f) => f.status === "error")
			.map((f) => ({
				fileName: f.fileName,
				error: f.error || "Unknown error",
			})),
	}));

	// ── Actions ──────────────────────────────────────────────────────────────

	/// Add files to the upload queue
	function addFiles(newFiles: File[]) {
		for (const file of newFiles) {
			// Skip duplicates by name
			if (files.value.some((f) => f.fileName === file.name)) {
				console.debug("[useBatchUpload] skipping duplicate file", {
					fileName: file.name,
				});
				continue;
			}

			files.value.push({
				file,
				fileName: file.name,
				progress: 0,
				status: "pending",
			});
		}

		console.debug("[useBatchUpload] files added", {
			newCount: newFiles.length,
			totalCount: files.value.length,
		});
	}

	/// Remove a file from the queue
	function removeFile(fileName: string) {
		files.value = files.value.filter((f) => f.fileName !== fileName);
		// Re-merge if files changed
		mergeSteps();
	}

	/// Start uploading all pending files
	async function uploadAll(ontologyId: string) {
		if (isProcessing.value) return;

		isProcessing.value = true;
		abortController.value = new AbortController();

		const pendingFiles = files.value.filter((f) => f.status === "pending");

		console.info("[useBatchUpload] batch upload start", {
			fileCount: pendingFiles.length,
			ontologyId,
		});

		for (const fileState of pendingFiles) {
			fileState.status = "uploading";

			try {
				const preview = await batchUploadDocuments({
					ontologyId,
					files: [fileState.file],
					onFileProgress: (_, pct) => {
						fileState.progress = pct;
					},
					signal: abortController.value.signal,
				});

				fileState.status = "success";
				fileState.progress = 100;
				fileState.preview = preview[0];

				console.debug("[useBatchUpload] file uploaded", {
					fileName: fileState.fileName,
					steps: preview[0].steps.length,
				});
			} catch (error) {
				fileState.status = "error";
				fileState.error =
					error instanceof Error ? error.message : "Upload failed";

				console.warn("[useBatchUpload] file failed", {
					fileName: fileState.fileName,
					error: fileState.error,
				});
			}
		}

		// After all uploads, merge and detect conflicts
		mergeSteps();
		detectConflicts();

		isProcessing.value = false;
		abortController.value = null;

		console.info("[useBatchUpload] batch upload complete", {
			successCount: successCount.value,
			failedCount: failedCount.value,
			totalSteps: mergedSteps.value.length,
			conflicts: conflicts.value.length,
		});
	}

	/// Cancel upload
	function cancelUpload() {
		if (abortController.value) {
			abortController.value.abort();
			abortController.value = null;
		}
		isProcessing.value = false;
	}

	/// Merge all successful previews into a single steps list
	function mergeSteps() {
		const allSteps: SequenceStep[] = [];
		const seenIds = new Set<string>();

		for (const fileState of files.value) {
			if (fileState.status !== "success" || !fileState.preview) continue;

			for (const step of fileState.preview.steps) {
				const stepWithSource: SequenceStep = {
					...step,
					sourceFile: fileState.fileName,
				};

				// Mark duplicates
				if (seenIds.has(step.entityId)) {
					stepWithSource.isDuplicate = true;
					const existing = allSteps.find((s) => s.entityId === step.entityId);
					if (existing) {
						stepWithSource.duplicateOf = existing.id;
					}
				}

				allSteps.push(stepWithSource);
				seenIds.add(step.entityId);
			}
		}

		mergedSteps.value = allSteps;

		console.debug("[useBatchUpload] steps merged", {
			totalSteps: allSteps.length,
			uniqueEntities: seenIds.size,
		});
	}

	/// Detect label conflicts (same entity, different labels across files)
	function detectConflicts() {
		const entityLabels = new Map<string, Map<string, string>>(); // entityId → Map<sourceFile, label>
		const found: ConflictResolution[] = [];

		for (const fileState of files.value) {
			if (fileState.status !== "success" || !fileState.preview) continue;

			for (const step of fileState.preview.steps) {
				if (!entityLabels.has(step.entityId)) {
					entityLabels.set(step.entityId, new Map());
				}
				const sources = entityLabels.get(step.entityId) as Map<string, string>;
				sources.set(fileState.fileName, step.label);
			}
		}

		// Check for conflicting labels
		for (const [entityId, sources] of entityLabels) {
			if (sources.size < 2) continue;

			const labels = Array.from(sources.entries());
			for (let i = 0; i < labels.length; i++) {
				for (let j = i + 1; j < labels.length; j++) {
					if (labels[i][1] !== labels[j][1]) {
						const stepA = findStep(entityId, labels[i][0]);
						const stepB = findStep(entityId, labels[j][0]);

						if (stepA && stepB) {
							found.push({
								stepA,
								stepB,
								field: "label",
								valueA: labels[i][1],
								valueB: labels[j][1],
								resolution: "keep_a",
							});
						}
					}
				}
			}
		}

		conflicts.value = found;

		if (found.length > 0) {
			console.info("[useBatchUpload] conflicts detected", {
				count: found.length,
			});
		}
	}

	/// Apply a conflict resolution
	function resolveConflict(
		index: number,
		resolution: ConflictResolution["resolution"],
		customValue?: string,
	) {
		if (index < 0 || index >= conflicts.value.length) return;

		const conflict = conflicts.value[index];
		conflict.resolution = resolution;
		conflict.customValue = customValue;

		console.debug("[useBatchUpload] conflict resolved", {
			index,
			resolution,
			entityId: conflict.stepA.entityId,
		});

		// Apply the resolution to merged steps
		if (resolution === "keep_a") {
			// keep_a is the default — no change needed
		} else if (resolution === "keep_b") {
			// Use stepB's label instead
			const stepInMerge = mergedSteps.value.find(
				(s) => s.id === conflict.stepA.id,
			);
			if (stepInMerge) {
				stepInMerge.label = conflict.valueB;
			}
		} else if (resolution === "custom" && customValue) {
			const stepInMerge = mergedSteps.value.find(
				(s) => s.id === conflict.stepA.id,
			);
			if (stepInMerge) {
				stepInMerge.label = customValue;
			}
		}
	}

	/// Reset the batch state
	function reset() {
		files.value = [];
		mergedSteps.value = [];
		conflicts.value = [];
		isProcessing.value = false;
		abortController.value = null;
	}

	// ── Helpers ──────────────────────────────────────────────────────────────

	function findStep(
		entityId: string,
		sourceFile: string,
	): SequenceStep | undefined {
		for (const fileState of files.value) {
			if (fileState.fileName !== sourceFile || !fileState.preview) continue;
			return fileState.preview.steps.find((s) => s.entityId === entityId);
		}
		return undefined;
	}

	return {
		// State
		files,
		mergedSteps,
		conflicts,
		isProcessing,
		batchResult,

		// Computed
		pendingCount,
		uploadingCount,
		successCount,
		failedCount,
		completedCount,
		totalCount,
		hasConflicts,
		allDone,
		hasFailedFiles,

		// Actions
		addFiles,
		removeFile,
		uploadAll,
		cancelUpload,
		resolveConflict,
		reset,
	};
}
