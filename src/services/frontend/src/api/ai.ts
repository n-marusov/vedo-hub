// REST API client for AI-assisted ontology operations
//
// Covers NL→OWL generation, iterative refinement, and AI suggestions.
// The API Gateway proxies these to the ai-orchestration-service.

import axios from "axios";
import type { AiSuggestion, SequenceStep } from "../types/extraction";

const AI_BASE = "/api/v1";

const api = axios.create({
	baseURL: AI_BASE,
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

export interface AiGenerationRequest {
	ontologyId: string;
	text: string;
	model?: string;
}

export interface AiGenerationResult {
	id: string;
	ontologyId: string;
	steps: SequenceStep[];
	tokenUsage?: {
		prompt: number;
		completion: number;
		total: number;
	};
}

export interface AiRefinementRequest {
	ontologyId: string;
	sequenceId: string;
	previousSteps: SequenceStep[];
	feedback: string;
	changes?: Array<{ stepId: string; operation: string }>;
}

export interface AiRefinementResult {
	id: string;
	ontologyId: string;
	steps: SequenceStep[];
	round: number;
	maxRounds: number;
}

export interface AiSuggestionRequest {
	ontologyId: string;
	classId?: string;
	context?: {
		classTree?: string[];
		existingProperties?: string[];
	};
}

export interface AiSuggestionResult {
	suggestions: AiSuggestion[];
}

/// Generate ontology structure from natural language description
export async function generateFromText(
	request: AiGenerationRequest,
): Promise<AiGenerationResult> {
	console.debug("[ai] generate-from-text start", {
		ontologyId: request.ontologyId,
		textLength: request.text.length,
	});

	try {
		const response = await api.post<AiGenerationResult>(
			`/ontologies/${request.ontologyId}/generate-from-text`,
			{
				text: request.text,
				model: request.model,
			},
			{
				timeout: 120000, // 2 min for LLM generation
			},
		);

		console.info("[ai] generate-from-text complete", {
			steps: response.data.steps.length,
			id: response.data.id,
		});

		return response.data;
	} catch (error) {
		let message: string;
		if (axios.isAxiosError(error) && error.response?.data) {
			// Extract human-readable error from the API response body
			const body = error.response.data as Record<string, unknown>;
			const errDetail = body.error as Record<string, unknown> | undefined;
			message = (errDetail?.message as string) || error.message;
		} else {
			message = error instanceof Error ? error.message : String(error);
		}

		console.error("[ai] generate-from-text failed", {
			error: message,
		});
		throw new Error(message);
	}
}

/// Refine a previously generated sequence using user feedback
export async function refineSequence(
	request: AiRefinementRequest,
): Promise<AiRefinementResult> {
	console.debug("[ai] refine start", {
		ontologyId: request.ontologyId,
		sequenceId: request.sequenceId,
		feedbackLength: request.feedback.length,
	});

	try {
		const response = await api.post<AiRefinementResult>(
			`/ontologies/${request.ontologyId}/ai/refine`,
			{
				sequence_id: request.sequenceId,
				previous_steps: request.previousSteps,
				feedback: request.feedback,
				changes: request.changes,
			},
			{
				timeout: 120000,
			},
		);

		console.info("[ai] refine complete", {
			steps: response.data.steps.length,
			round: response.data.round,
		});

		return response.data;
	} catch (error) {
		console.error("[ai] refine failed", {
			error: error instanceof Error ? error.message : String(error),
		});
		throw error;
	}
}

/// Suggest classes for a given ontology context
export async function suggestClasses(
	request: AiSuggestionRequest,
): Promise<AiSuggestionResult> {
	console.debug("[ai] suggest-classes start", {
		ontologyId: request.ontologyId,
		classId: request.classId,
	});

	try {
		const response = await api.post<AiSuggestionResult>(
			`/ontologies/${request.ontologyId}/ai/suggest-classes`,
			{
				class_id: request.classId,
				context: request.context,
			},
			{ timeout: 60000 },
		);
		return response.data;
	} catch (error) {
		console.error("[ai] suggest-classes failed", {
			error: error instanceof Error ? error.message : String(error),
		});
		throw error;
	}
}

/// Suggest properties for a given class
export async function suggestProperties(
	request: AiSuggestionRequest,
): Promise<AiSuggestionResult> {
	console.debug("[ai] suggest-properties start", {
		ontologyId: request.ontologyId,
		classId: request.classId,
	});

	try {
		const response = await api.post<AiSuggestionResult>(
			`/ontologies/${request.ontologyId}/ai/suggest-properties`,
			{
				class_id: request.classId,
				context: request.context,
			},
			{ timeout: 60000 },
		);
		return response.data;
	} catch (error) {
		console.error("[ai] suggest-properties failed", {
			error: error instanceof Error ? error.message : String(error),
		});
		throw error;
	}
}

/// Suggest relationships between classes
export async function suggestRelationships(
	request: AiSuggestionRequest,
): Promise<AiSuggestionResult> {
	console.debug("[ai] suggest-relationships start", {
		ontologyId: request.ontologyId,
		classId: request.classId,
	});

	try {
		const response = await api.post<AiSuggestionResult>(
			`/ontologies/${request.ontologyId}/ai/suggest-relationships`,
			{
				class_id: request.classId,
				context: request.context,
			},
			{ timeout: 60000 },
		);
		return response.data;
	} catch (error) {
		console.error("[ai] suggest-relationships failed", {
			error: error instanceof Error ? error.message : String(error),
		});
		throw error;
	}
}
