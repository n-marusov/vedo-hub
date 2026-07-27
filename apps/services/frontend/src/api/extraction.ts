// REST API client for document extraction endpoints
//
// Document upload and extraction use REST (not GraphQL) because they require
// multipart file upload with progress tracking. The API Gateway proxies
// these requests to the document-extractor service.

import axios, { type AxiosProgressEvent } from "axios";
import type {
	ApplyResult,
	ExtractionError,
	ExtractionPreview,
	SequenceStep,
} from "../types/extraction";

const EXTRACTION_BASE = "/api/v1/documents";

const api = axios.create({
	baseURL: EXTRACTION_BASE,
	headers: { "X-Requested-With": "XMLHttpRequest" },
});

// Inject JWT token from localStorage
api.interceptors.request.use((config) => {
	const token = localStorage.getItem("vedo-jwt-token");
	if (token) {
		config.headers.Authorization = `Bearer ${token}`;
	}
	return config;
});

export interface UploadOptions {
	ontologyId: string;
	file: File;
	onProgress?: (progress: number) => void;
	signal?: AbortSignal;
}

export interface BatchUploadOptions {
	ontologyId: string;
	files: File[];
	onFileProgress?: (fileName: string, progress: number) => void;
	signal?: AbortSignal;
}

export interface ApplySequenceOptions {
	ontologyId: string;
	steps: SequenceStep[];
	commitMessage: string;
	signal?: AbortSignal;
}

function mapProgress(event: AxiosProgressEvent): number {
	if (!event.total) return 0;
	return Math.round((event.loaded / event.total) * 100);
}

/// Upload a single document for extraction
export async function uploadDocument(
	options: UploadOptions,
): Promise<ExtractionPreview> {
	const formData = new FormData();
	formData.append("file", options.file);
	formData.append("ontology_id", options.ontologyId);

	console.debug("[extraction] upload start", {
		fileName: options.file.name,
		fileSize: options.file.size,
		fileType: options.file.type,
		ontologyId: options.ontologyId,
	});

	try {
		const response = await api.post<ExtractionPreview>("/extract", formData, {
			headers: { "Content-Type": "multipart/form-data" },
			onUploadProgress: (event) => {
				const pct = mapProgress(event);
				options.onProgress?.(pct);
			},
			signal: options.signal,
			timeout: 120000, // 2 min for extraction
		});

		console.info("[extraction] upload complete", {
			fileName: options.file.name,
			steps: response.data.steps.length,
			previewId: response.data.id,
		});

		return response.data;
	} catch (error) {
		const extractionError: ExtractionError = {
			code: "UPLOAD_FAILED",
			message: error instanceof Error ? error.message : "Unknown upload error",
		};

		console.warn("[extraction] upload failed", {
			fileName: options.file.name,
			error: extractionError.message,
		});

		throw extractionError;
	}
}

/// Upload multiple files for batch extraction
export async function batchUploadDocuments(
	options: BatchUploadOptions,
): Promise<ExtractionPreview[]> {
	const previews: ExtractionPreview[] = [];

	console.info("[extraction] batch upload start", {
		fileCount: options.files.length,
		ontologyId: options.ontologyId,
	});

	for (const file of options.files) {
		try {
			const preview = await uploadDocument({
				ontologyId: options.ontologyId,
				file,
				onProgress: (pct) => options.onFileProgress?.(file.name, pct),
				signal: options.signal,
			});
			previews.push(preview);
		} catch (error) {
			console.warn("[extraction] batch file failed", {
				fileName: file.name,
				error: error instanceof Error ? error.message : String(error),
			});
			// Re-throw to let the caller handle per-file failures
			throw error;
		}
	}

	console.info("[extraction] batch upload complete", {
		fileCount: previews.length,
		totalSteps: previews.reduce((sum, p) => sum + p.steps.length, 0),
	});

	return previews;
}

/// Apply extracted sequence to the ontology
export async function applySequence(
	options: ApplySequenceOptions,
): Promise<ApplyResult> {
	console.debug("[extraction] apply start", {
		ontologyId: options.ontologyId,
		stepCount: options.steps.filter((s) => s.included).length,
	});

	try {
		const response = await api.post<ApplyResult>(
			"/apply",
			{
				ontology_id: options.ontologyId,
				steps: options.steps.filter((s) => s.included),
				commit_message: options.commitMessage,
			},
			{
				signal: options.signal,
				timeout: 300000, // 5 min for apply
			},
		);

		const result = response.data;

		if (result.success) {
			console.info("[extraction] apply complete", {
				appliedCount: result.appliedCount,
				commitId: result.commitId,
			});
		} else {
			console.warn("[extraction] apply partial failure", {
				appliedCount: result.appliedCount,
				errorCount: result.errors?.length ?? 0,
			});
		}

		return result;
	} catch (error) {
		const message =
			error instanceof Error ? error.message : "Unknown apply error";

		console.error("[extraction] apply failed", {
			ontologyId: options.ontologyId,
			error: message,
		});

		// Return a structured ApplyResult with success:false instead of throwing
		// a plain ExtractionError object. useApplySequence checks result.errors
		// for the error message — throwing means the catch block in the composable
		// only sees error.message, which for non-Error objects defaults to 'Unknown error'.
		return {
			success: false,
			appliedCount: 0,
			errors: [{ code: "APPLY_FAILED", message }],
		};
	}
}

/// Get extraction preview by ID (for re-fetching after navigation)
export async function getExtractionPreview(
	ontologyId: string,
	previewId: string,
): Promise<ExtractionPreview> {
	console.debug("[extraction] get preview", { ontologyId, previewId });

	const response = await api.get<ExtractionPreview>(`/preview/${previewId}`, {
		params: { ontology_id: ontologyId },
	});

	return response.data;
}

/// Check supported file formats from the server
export async function getSupportedFormats(): Promise<string[]> {
	const response = await api.get<{ formats: string[] }>("/formats");
	return response.data.formats;
}

export const ALLOWED_FORMATS = [
	".md",
	".txt",
	".pdf",
	".docx",
	".json",
	".xml",
	".csv",
	".xlsx",
];

export const ALLOWED_MIME_TYPES: Record<string, string[]> = {
	".md": ["text/markdown", "text/plain"],
	".txt": ["text/plain"],
	".pdf": ["application/pdf"],
	".docx": [
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	],
	".json": ["application/json"],
	".xml": ["text/xml", "application/xml"],
	".csv": ["text/csv"],
	".xlsx": [
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	],
};

export const FORMAT_LABELS: Record<string, string> = {
	".md": "Markdown",
	".txt": "Text",
	".pdf": "PDF",
	".docx": "DOCX",
	".json": "JSON",
	".xml": "XML",
	".csv": "CSV",
	".xlsx": "XLSX",
};
