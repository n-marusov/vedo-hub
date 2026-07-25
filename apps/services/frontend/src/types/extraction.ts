// Shared types for document extraction, sequence preview, and OWL generation
//
// These types match the API contracts from document-extractor service
// and ontology-service (POST /api/v1/ontology/apply-sequence).

export interface ExtractionError {
	code: string;
	message: string;
	details?: Record<string, unknown>;
}

export interface ExtractionPreview {
	id: string;
	ontologyId: string;
	sourceFile: string;
	steps: SequenceStep[];
	totalSteps: number;
	createdAt: string;
}

export interface ExtractionProgress {
	loaded: number;
	total: number;
	percentage: number;
}

export type OperationType =
	| "CREATE_CLASS"
	| "CREATE_PROPERTY"
	| "CREATE_INDIVIDUAL"
	| "UPDATE_LABEL"
	| "UPDATE_COMMENT"
	| "DELETE";

export interface SequenceStep {
	id: string;
	operation: OperationType;
	entityId: string;
	label: string;
	comment?: string;
	parentId?: string; // parent class for subclass relation
	parentLabel?: string;
	domain?: string; // for properties
	range?: string; // for properties
	xsdType?: string; // for datatype properties
	sourceFile?: string; // batch mode — which file this step originated from
	isDuplicate?: boolean;
	duplicateOf?: string;
	included: boolean;
}

export interface ApplyResult {
	success: boolean;
	appliedCount: number;
	commitId?: string;
	commitUrl?: string;
	errors?: Array<{ stepId: string; message: string }>;
	timestamp: string;
}

export interface ApplyProgress {
	totalSteps: number;
	completedSteps: number;
	currentStep: string;
	status: "idle" | "confirming" | "applying" | "success" | "error";
	errorMessage?: string;
}

export interface ConflictResolution {
	stepA: SequenceStep;
	stepB: SequenceStep;
	field: string;
	valueA: string;
	valueB: string;
	resolution: "keep_a" | "keep_b" | "custom";
	customValue?: string;
}

export interface BatchUploadResult {
	completedFiles: number;
	failedFiles: number;
	totalSteps: number;
	previews: ExtractionPreview[];
	failed: Array<{ fileName: string; error: string }>;
}

/** AI suggestion for classes, properties, or relationships. */
export interface AiSuggestion {
	id: string;
	label: string;
	description?: string;
	rationale: string;
	confidence: number;
	parentId?: string;
	parentLabel?: string;
	type: "class" | "property" | "relationship";
}
