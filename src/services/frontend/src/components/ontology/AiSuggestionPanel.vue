<!-- @ctx: Organism AiSuggestionPanel — AI-powered class/property/relationship suggestions with accept/reject -->
<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<template>
  <div class="ai-suggestions" role="region" aria-label="AI suggestions">
    <!-- Header -->
    <div class="ai-suggestions__header">
      <h3 class="ai-suggestions__title">
        <Zap :size="14" class="ai-suggestions__icon" />
        AI Suggestions
      </h3>
      <span v-if="acceptedCount > 0" class="ai-suggestions__accepted-badge">
        {{ acceptedCount }} accepted
      </span>
    </div>

    <!-- Action buttons (shown when not loading and no suggestions active) -->
    <div v-if="!loading && suggestions.length === 0 && !errorMsg" class="ai-suggestions__actions">
      <button
        class="ai-suggestions__action-btn"
        type="button"
        data-testid="suggest-subclasses"
        :disabled="isRequesting"
        @click="fetchSuggestions('class')"
      >
        Suggest subclasses
      </button>
      <button
        class="ai-suggestions__action-btn"
        type="button"
        data-testid="suggest-properties"
        :disabled="isRequesting"
        @click="fetchSuggestions('property')"
      >
        Suggest properties
      </button>
      <button
        class="ai-suggestions__action-btn"
        type="button"
        data-testid="suggest-relationships"
        :disabled="isRequesting"
        @click="fetchSuggestions('relationship')"
      >
        Suggest relationships
      </button>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="ai-suggestions__loading">
      <span class="btn-spinner"></span>
      <span class="ai-suggestions__loading-text">Generating suggestions...</span>
    </div>

    <!-- Error state -->
    <div v-if="errorMsg" class="ai-suggestions__error">
      <span class="ai-suggestions__error-icon">!</span>
      <span class="ai-suggestions__error-text">{{ errorMsg }}</span>
      <button class="ai-suggestions__retry-btn" type="button" @click="retry">Retry</button>
    </div>

    <!-- Empty state -->
    <div v-if="!loading && !errorMsg && suggestions.length === 0 && !isRequesting" class="ai-suggestions__empty">
      <p class="ai-suggestions__empty-text">Select a class and request suggestions to get AI-powered recommendations.</p>
    </div>

    <!-- Suggestions list -->
    <div v-if="suggestions.length > 0 && !loading" class="ai-suggestions__list">
      <div
        v-for="suggestion in suggestions"
        :key="suggestion.id"
        class="ai-suggestion-item"
        :class="{ 'ai-suggestion-item--accepted': acceptedIds.has(suggestion.id) }"
        data-testid="ai-suggestion-item"
      >
        <div class="ai-suggestion-item__header">
          <span class="ai-suggestion-item__type-badge" :class="`ai-suggestion-item__type-badge--${suggestion.type}`">
            {{ typeLabel(suggestion.type) }}
          </span>
          <span class="ai-suggestion-item__confidence" :title="`Confidence: ${Math.round(suggestion.confidence * 100)}%`">
            {{ Math.round(suggestion.confidence * 100) }}%
          </span>
        </div>

        <div class="ai-suggestion-item__label">{{ suggestion.label }}</div>
        <div v-if="suggestion.description" class="ai-suggestion-item__desc">{{ suggestion.description }}</div>
        <div v-if="suggestion.parentLabel" class="ai-suggestion-item__parent">
          Parent: <code>{{ suggestion.parentLabel }}</code>
        </div>
        <div class="ai-suggestion-item__rationale">{{ suggestion.rationale }}</div>

        <div class="ai-suggestion-item__actions">
          <button
            v-if="!acceptedIds.has(suggestion.id)"
            class="ai-suggestion-item__accept-btn"
            type="button"
            :disabled="acceptingId === suggestion.id"
            @click="acceptSuggestion(suggestion)"
          >
            <span v-if="acceptingId === suggestion.id" class="btn-spinner"></span>
            {{ acceptingId === suggestion.id ? 'Applying...' : 'Accept' }}
          </button>
          <button
            v-else
            class="ai-suggestion-item__accepted-btn"
            type="button"
            disabled
          >
            ✓ Accepted
          </button>
          <button
            class="ai-suggestion-item__reject-btn"
            type="button"
            :disabled="acceptingId === suggestion.id"
            @click="rejectSuggestion(suggestion.id)"
          >
            Dismiss
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Zap } from "lucide-vue-next";
import { computed, ref } from "vue";
import {
	suggestClasses,
	suggestProperties,
	suggestRelationships,
} from "../../api/ai";
import type { AiSuggestionResult } from "../../api/ai";
import { useErrorPresentation } from "../../composables/useErrorPresentation";
import type { AiSuggestion } from "../../types/extraction";

const props = defineProps<{
	ontologyId: string;
	classId: string | null;
	disabled?: boolean;
}>();

const emit = defineEmits<{
	"suggestion-accepted": [suggestion: AiSuggestion];
	"suggestion-rejected": [suggestionId: string];
}>();

const { addError } = useErrorPresentation();

// ── State ──────────────────────────────────────────────────────────────────

const loading = ref(false);
const isRequesting = ref(false);
const errorMsg = ref<string | null>(null);
const suggestions = ref<AiSuggestion[]>([]);
const acceptedIds = ref(new Set<string>());
const acceptingId = ref<string | null>(null);
const currentType = ref<"class" | "property" | "relationship" | null>(null);

// ── Computed ───────────────────────────────────────────────────────────────

const acceptedCount = computed(() => acceptedIds.value.size);

function typeLabel(type: string): string {
	switch (type) {
		case "class":
			return "Class";
		case "property":
			return "Property";
		case "relationship":
			return "Relationship";
		default:
			return type;
	}
}

// ── Fetch suggestions ────────────────────────────────────────────────────

async function fetchSuggestions(
	type: "class" | "property" | "relationship",
): Promise<void> {
	if (!props.classId || isRequesting.value) return;

	isRequesting.value = true;
	loading.value = true;
	errorMsg.value = null;
	suggestions.value = [];
	currentType.value = type;

	console.info("[AiSuggestionPanel] fetching suggestions", {
		ontologyId: props.ontologyId,
		classId: props.classId,
		type,
	});

	try {
		let result: AiSuggestionResult | undefined;
		switch (type) {
			case "class":
				result = await suggestClasses({
					ontologyId: props.ontologyId,
					classId: props.classId,
				});
				break;
			case "property":
				result = await suggestProperties({
					ontologyId: props.ontologyId,
					classId: props.classId,
				});
				break;
			case "relationship":
				result = await suggestRelationships({
					ontologyId: props.ontologyId,
					classId: props.classId,
				});
				break;
		}

		suggestions.value = result?.suggestions ?? [];

		console.info("[AiSuggestionPanel] suggestions received", {
			count: result?.suggestions.length ?? 0,
			type,
		});
	} catch (err) {
		const msg = err instanceof Error ? err.message : String(err);
		errorMsg.value = msg;
		addError("AI-SUGGEST-FAILED", msg);
		console.error("[AiSuggestionPanel] failed to fetch suggestions", {
			error: msg,
		});
	} finally {
		loading.value = false;
		isRequesting.value = false;
	}
}

// ── Accept suggestion ─────────────────────────────────────────────────────

async function acceptSuggestion(suggestion: AiSuggestion): Promise<void> {
	acceptingId.value = suggestion.id;

	console.info("[AiSuggestionPanel] accepting suggestion", {
		id: suggestion.id,
		label: suggestion.label,
		type: suggestion.type,
	});

	try {
		// Emit for the parent to handle the actual CRUD operation
		emit("suggestion-accepted", suggestion);
		acceptedIds.value.add(suggestion.id);

		console.info("[AiSuggestionPanel] suggestion accepted", {
			id: suggestion.id,
		});
	} catch (err) {
		const msg = err instanceof Error ? err.message : String(err);
		addError("AI-ACCEPT-FAILED", msg);
		console.error("[AiSuggestionPanel] accept failed", {
			error: msg,
		});
	} finally {
		acceptingId.value = null;
	}
}

// ── Reject / dismiss suggestion ───────────────────────────────────────────

function rejectSuggestion(suggestionId: string): void {
	console.info("[AiSuggestionPanel] dismissing suggestion", {
		id: suggestionId,
	});
	suggestions.value = suggestions.value.filter((s) => s.id !== suggestionId);
	emit("suggestion-rejected", suggestionId);
}

// ── Retry after error ─────────────────────────────────────────────────────

function retry(): void {
	if (isRequesting.value) return;
	if (currentType.value) {
		fetchSuggestions(currentType.value);
	} else {
		errorMsg.value = null;
		suggestions.value = [];
	}
}
</script>

<style scoped>
.ai-suggestions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3, 12px);
  padding: var(--spacing-3, 12px);
  border-top: 1px solid var(--border);
}

.ai-suggestions__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.ai-suggestions__title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 6px);
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #fafafa);
  margin: 0;
}

.ai-suggestions__icon {
  color: var(--primary, #10b981);
}

.ai-suggestions__accepted-badge {
  font-size: var(--font-size-xs, 12px);
  color: var(--primary, #10b981);
  background: rgba(16, 185, 129, 0.1);
  padding: 2px 8px;
  border-radius: var(--radius-sm, 6px);
}

.ai-suggestions__actions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2, 6px);
}

.ai-suggestions__action-btn {
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-secondary, #141414);
  color: var(--text-primary, #fafafa);
  font-size: var(--font-size-sm, 13px);
  font-family: "IBM Plex Mono", monospace;
  cursor: pointer;
  text-align: left;
  transition: border-color var(--transition-fast, 0.15s);
}

.ai-suggestions__action-btn:hover:not(:disabled) {
  border-color: var(--primary, #10b981);
}

.ai-suggestions__action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ai-suggestions__loading {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 8px);
  padding: var(--spacing-3, 12px);
  color: var(--text-muted, #6b7280);
  font-size: var(--font-size-sm, 13px);
}

.ai-suggestions__loading-text {
  font-style: italic;
}

.ai-suggestions__error {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 8px);
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-md, 8px);
  background: rgba(239, 68, 68, 0.05);
  color: var(--status-error, #ef4444);
  font-size: var(--font-size-sm, 13px);
}

.ai-suggestions__error-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(239, 68, 68, 0.15);
  font-weight: 700;
  font-size: 12px;
  flex-shrink: 0;
}

.ai-suggestions__error-text {
  flex: 1;
}

.ai-suggestions__retry-btn {
  padding: var(--spacing-1, 4px) var(--spacing-2, 8px);
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-secondary, #6b7280);
  font-size: var(--font-size-xs, 12px);
  cursor: pointer;
}

.ai-suggestions__retry-btn:hover {
  background: var(--surface-secondary, #141414);
}

.ai-suggestions__empty {
  padding: var(--spacing-3, 12px);
}

.ai-suggestions__empty-text {
  margin: 0;
  font-size: var(--font-size-sm, 13px);
  color: var(--text-muted, #6b7280);
  font-style: italic;
}

/* ── Suggestion items ─────────────────────────────────── */

.ai-suggestions__list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2, 8px);
}

.ai-suggestion-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1, 4px);
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-secondary, #141414);
  transition: border-color var(--transition-fast, 0.15s);
}

.ai-suggestion-item--accepted {
  border-color: var(--primary, #10b981);
  background: rgba(16, 185, 129, 0.05);
}

.ai-suggestion-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.ai-suggestion-item__type-badge {
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-medium, 500);
  padding: 1px 6px;
  border-radius: var(--radius-sm, 6px);
}

.ai-suggestion-item__type-badge--class {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.ai-suggestion-item__type-badge--property {
  background: rgba(168, 85, 247, 0.15);
  color: #c084fc;
}

.ai-suggestion-item__type-badge--relationship {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
}

.ai-suggestion-item__confidence {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #6b7280);
  font-weight: var(--font-weight-medium, 500);
}

.ai-suggestion-item__label {
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #fafafa);
}

.ai-suggestion-item__desc {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #a1a1aa);
}

.ai-suggestion-item__parent {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #6b7280);
}

.ai-suggestion-item__parent code {
  color: var(--text-secondary, #a1a1aa);
}

.ai-suggestion-item__rationale {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #6b7280);
  font-style: italic;
  line-height: 1.4;
}

.ai-suggestion-item__actions {
  display: flex;
  gap: var(--spacing-2, 8px);
  margin-top: var(--spacing-1, 4px);
}

.ai-suggestion-item__accept-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1, 4px);
  padding: var(--spacing-1, 4px) var(--spacing-2, 8px);
  border: 1px solid var(--primary, #10b981);
  border-radius: var(--radius-sm, 6px);
  background: var(--primary, #10b981);
  color: var(--text-inverse, #0a0a0a);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-medium, 500);
  cursor: pointer;
}

.ai-suggestion-item__accept-btn:hover:not(:disabled) {
  background: var(--primary-hover, #34d399);
}

.ai-suggestion-item__accept-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ai-suggestion-item__accepted-btn {
  padding: var(--spacing-1, 4px) var(--spacing-2, 8px);
  border: 1px solid var(--primary, #10b981);
  border-radius: var(--radius-sm, 6px);
  background: transparent;
  color: var(--primary, #10b981);
  font-size: var(--font-size-xs, 12px);
  cursor: default;
}

.ai-suggestion-item__reject-btn {
  padding: var(--spacing-1, 4px) var(--spacing-2, 8px);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  background: transparent;
  color: var(--text-secondary, #6b7280);
  font-size: var(--font-size-xs, 12px);
  cursor: pointer;
}

.ai-suggestion-item__reject-btn:hover:not(:disabled) {
  border-color: rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

.ai-suggestion-item__reject-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
