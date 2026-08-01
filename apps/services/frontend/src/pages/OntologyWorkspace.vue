<!-- @ctx: Ontology workspace page — wired to real backend via Apollo GraphQL queries -->
<template>
  <div class="workspace-page" role="main" :aria-label="t('ontology_workspace.page_label')">
    <div v-if="loading" class="workspace-loading">
      <span>{{ t('ontology_workspace.loading') }}</span>
    </div>

    <div v-else-if="error" class="workspace-error">
      <span>{{ t('ontology_workspace.load_error', { error }) }}</span>
    </div>

    <template v-else>
   <div class="workspace-main">
    <aside class="group-sidebar card-side panel-left">
          <div class="group-header">{{ t('ontology_workspace.project_group') }}</div>
        </aside>

        <div class="splitter"><GripVertical :size="8" /></div>

        <section class="workspace-content panel-center">
            <div class="workspace-toolbar">
              <span class="toolbar-title">{{ ontologyData?.name || ontologyId }}</span>
              <span v-if="ontologyData?.branch" class="toolbar-branch-badge">{{ ontologyData.branch }}</span>
              <span v-if="ontologyData?.dirty" class="toolbar-dirty-badge">{{ t('ontology_workspace.dirty') }}</span>
              <span class="toolbar-spacer"></span>
              <button class="toolbar-btn" type="button">{{ t('ontology_workspace.publish') }}</button>
              <button
                class="toolbar-btn toolbar-btn--primary"
                type="button"
                :disabled="!draftState.hasUnsavedChanges.value || saving"
                @click="handleSave"
              >
                <span v-if="saving" class="btn-spinner"></span>
                {{ saving ? t('ontology_workspace.saving') : t('common.save') }}
              </button>
              <div class="toolbar-create-group">
                <button
                  class="toolbar-btn toolbar-create-btn"
                  type="button"
                  :title="t('ontology_workspace.create_entity')"
                  @click="openCreateClass"
                >
                  <Plus :size="14" />
                  {{ t('ontology_workspace.create') }}
                </button>
                <div class="toolbar-create-dropdown">
                  <button type="button" @click="openCreateClass">{{ t('ontology_workspace.create_class') }}</button>
                  <button type="button" @click="openCreateProperty">{{ t('ontology_workspace.create_property') }}</button>
                  <button type="button" @click="openCreateIndividual">{{ t('ontology_workspace.create_individual') }}</button>
                </div>
              </div>
              <button
                :class="['toolbar-btn', { 'toolbar-btn--active': showAiPanel }]"
                type="button"
                @click="showAiPanel = !showAiPanel"
              >
                <Zap :size="14" />
                {{ t('ontology_workspace.ai_import') }}
              </button>
            </div>

            <!-- AI Import Panel (toggleable) -->
            <div v-if="showAiPanel" class="workspace-ai-panel">
              <div class="ai-panel__header">
                <h3 class="ai-panel__title">{{ t('ontology_workspace.ai_import_title') }}</h3>
                <p class="ai-panel__desc">{{ t('ontology_workspace.ai_import_desc') }}</p>
              </div>

              <!-- Tab bar: Document Import / NL→OWL -->
              <div class="ai-panel__tabs">
                <button
                  :class="['ai-panel__tab', { 'ai-panel__tab--active': aiTab === 'document' }]"
                  type="button"
                  data-testid="ai-tab-document"
                  @click="switchAiTab('document')"
                >
                  <FileText :size="14" />
                  {{ t('ontology_workspace.document_import') }}
                </button>
                <button
                  :class="['ai-panel__tab', { 'ai-panel__tab--active': aiTab === 'nl-to-owl' }]"
                  type="button"
                  data-testid="ai-tab-nl-to-owl"
                  @click="switchAiTab('nl-to-owl')"
                >
                  <MessageSquare :size="14" />
                  NL→OWL
                </button>
              </div>

              <!-- ── Document Import Tab ── -->
              <template v-if="aiTab === 'document'">
                <!-- Upload section -->
                <div v-if="uploadMode === 'single'" class="ai-panel__section">
                  <DocumentUploader
                    :ontology-id="ontologyId"
                    :mode="'single'"
                    @upload-complete="onUploadComplete"
                    @upload-error="onUploadError"
                  />
                </div>
                <div v-else class="ai-panel__section">
                  <BatchUploader
                    :ontology-id="ontologyId"
                    @batch-complete="onBatchComplete"
                    @batch-error="onUploadError"
                    @reset="onBatchReset"
                  />
                </div>

                <!-- Upload mode toggle -->
                <div class="ai-panel__mode-toggle">
                  <button
                    :class="['ai-panel__mode-btn', { 'ai-panel__mode-btn--active': uploadMode === 'single' }]"
                    type="button"
                    @click="uploadMode = 'single'"
                  >
                    {{ t('ontology_workspace.single_file') }}
                  </button>
                  <button
                    :class="['ai-panel__mode-btn', { 'ai-panel__mode-btn--active': uploadMode === 'batch' }]"
                    type="button"
                    @click="uploadMode = 'batch'"
                  >
                    {{ t('ontology_workspace.batch_upload') }}
                  </button>
                </div>
              </template>

              <!-- ── NL→OWL Generation Tab ── -->
              <template v-if="aiTab === 'nl-to-owl'">
                <!-- Prompt input -->
                <div class="ai-panel__section">
                  <div class="nl-prompt">
                    <label class="nl-prompt__label" for="nl-to-owl-input">
                      {{ t('ontology_workspace.nl_label') }}
                    </label>
                    <textarea
                      id="nl-to-owl-input"
                      v-model="nlPrompt"
                      class="nl-prompt__textarea"
                      data-testid="nl-to-owl-input"
                      :placeholder="t('ontology_workspace.nl_placeholder')"
                      :disabled="isGenerating"
                      rows="4"
                    ></textarea>
                    <div class="nl-prompt__actions">
                      <span class="nl-prompt__hint">{{ t('ontology_workspace.nl_hint') }}</span>
                      <button
                        class="nl-prompt__generate-btn"
                        type="button"
                        data-testid="nl-to-owl-generate"
                        :disabled="!nlPrompt.trim() || isGenerating"
                        @click="handleGenerateFromText"
                      >
                        <span v-if="isGenerating" class="btn-spinner"></span>
                        {{ isGenerating ? t('ontology_workspace.generating') : t('ontology_workspace.generate') }}
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Generation error -->
                <div v-if="generationError" class="ai-panel__section">
                  <div class="nl-error">
                    <span class="nl-error__icon">!</span>
                    <span>{{ generationError }}</span>
                    <button class="nl-error__retry" type="button" @click="handleGenerateFromText">{{ t('common.retry') }}</button>
                  </div>
                </div>

                <!-- Refinement section (shown after generation) -->
                <div v-if="nlResult && nlResult.steps.length > 0 && !isGenerating" class="ai-panel__section">
                  <div class="nl-refinement">
                    <label class="nl-refinement__label" for="refinement-input">
                      {{ t('ontology_workspace.refine_label') }}
                    </label>
                    <div class="nl-refinement__row">
                      <input
                        id="refinement-input"
                        v-model="refinementFeedback"
                        class="nl-refinement__input"
                        data-testid="refinement-input"
                        type="text"
                        :placeholder="t('ontology_workspace.refine_placeholder')"
                        :disabled="isRefining"
                      />
                      <button
                        class="nl-refinement__btn"
                        type="button"
                        data-testid="refinement-submit"
                        :disabled="!refinementFeedback.trim() || isRefining"
                        @click="handleRefine"
                      >
                        <span v-if="isRefining" class="btn-spinner"></span>
                        {{ isRefining ? t('ontology_workspace.refining') : t('ontology_workspace.refine') }}
                      </button>
                    </div>
                    <span v-if="refinementRound > 0" class="nl-refinement__round">
                      {{ t('ontology_workspace.refinement_round', { round: String(refinementRound), max: String(maxRefinementRounds) }) }}
                    </span>
                  </div>
                </div>
              </template>

              <!-- Sequence Preview (shared between tabs) -->
              <div v-if="extractionSteps.length > 0" class="ai-panel__section">
                <SequencePreview
                  :steps="extractionSteps"
                  :ontology-id="ontologyId"
                  :source-files="extractionSourceFiles"
                  @apply="onApplySequence"
                  @cancel="onApplyCancel"
                />
              </div>

              <!-- Apply button (shown after preview) -->
              <div v-if="showApplyButton && extractionSteps.length > 0" class="ai-panel__apply">
                <ApplySequenceButton
                  :ontology-id="ontologyId"
                  :steps="extractionSteps"
                  :disabled="extractionSteps.filter(s => s.included).length === 0"
                  @apply-success="onApplySuccess"
                  @apply-error="onApplyError"
                />
              </div>
            </div>

            <!-- Regular workspace view (hidden when AI panel is open) -->
            <div v-if="!showAiPanel" class="workspace-row">
            <aside class="class-panel card-side">
              <div class="panel-tools">
                <Search :size="14" class="muted" />
                <div class="panel-input">{{ t('ontology_workspace.filter_classes') }}</div>
              </div>

              <div class="class-list">
                <button
                  v-for="cls in classTree"
                  :key="cls.id"
                  class="class-row"
                  :class="{ 'class-row--active': cls.id === selectedClassId }"
                  type="button"
                  @click="selectClass(cls.id)"
                >
                  <ChevronRight v-if="cls.children?.length" :size="12" />
                  <Folder :size="14" />
                  {{ cls.label }}
                </button>
              </div>
            </aside>

            <div class="splitter"><GripVertical :size="8" /></div>

            <section class="graph-panel card-side">
              <div class="graph-panel__toolbar">
                <button
                  class="graph-panel__toggle"
                  :class="{ 'graph-panel__toggle--active': viewMode === 'graph' }"
                  @click="viewMode = 'graph'"
                >
                  {{ t('ontology_workspace.graph_view') }}
                </button>
                <button
                  class="graph-panel__toggle"
                  :class="{ 'graph-panel__toggle--active': viewMode === 'table' }"
                  @click="viewMode = 'table'"
                >
                  {{ t('ontology_workspace.table_view') }}
                </button>
              </div>

              <!-- Graph Visualization view -->
              <GraphVisualization
                v-if="viewMode === 'graph'"
                :nodes="graphNodes"
                :edges="graphEdges"
                :show-table-fallback="false"
                @node-click="onGraphNodeClick"
              />

              <!-- Detail panel overlay (visible when a class is selected) -->
              <div v-if="selectedClass" class="detail-panel">
                <div class="detail-panel__header">
                  <span class="detail-panel__title">{{ selectedClass.label }}</span>
                </div>
                <div v-if="selectedClass.comment" class="detail-panel__comment">
                  {{ selectedClass.comment }}
                </div>
              </div>

              <!-- Table view (existing) -->
              <template v-else>
                <div class="graph-head">
                  <span class="col-ind">{{ t('ontology_workspace.col_individual') }}</span>
                  <span class="col-prop">{{ t('ontology_workspace.col_property') }}</span>
                  <span class="col-val">{{ t('ontology_workspace.col_value') }}</span>
                </div>
                <div
                  v-for="ind in individuals"
                  :key="ind.id"
                  class="graph-row"
                  :class="{ 'graph-row--active': ind.id === selectedIndividualId }"
                >
                  <span class="col-ind">{{ ind.label }}</span>
                  <span class="col-prop">rdf:type</span>
                  <span class="col-val">{{ ind.classLabel }}</span>
                </div>
                <div v-if="individuals.length === 0" class="graph-empty">
                  <span class="muted">{{ t('ontology_workspace.no_individuals') }}</span>
                </div>
              </template>
            </section>

            <div class="splitter"><GripVertical :size="8" /></div>

            <aside class="property-panel card-side panel-right">
              <div class="panel-title">{{ t('ontology_workspace.individuals') }}</div>
              <div class="panel-tools">
                <Search :size="14" class="muted" />
                <div class="panel-input">{{ t('ontology_workspace.filter_individuals') }}</div>
              </div>
              <div class="indiv-head"><span class="i-name">{{ t('ontology_workspace.col_name') }}</span><span class="i-type">{{ t('ontology_workspace.col_type') }}</span></div>
              <div v-for="item in individuals" :key="item.id" class="indiv-row">
                <span class="i-name">{{ item.label }}</span>
                <span class="i-type i-type--active">{{ item.classLabel }}</span>
              </div>
              <AiSuggestionPanel
                v-if="selectedClassId"
                :ontology-id="ontologyId"
                :class-id="selectedClassId"
                @suggestion-accepted="onAiSuggestionAccepted"
              />
            </aside>
          </div>
        </section>
      </div>
    </template>

      <!-- Create Entity Dialogs -->
      <CreateClassDialog
        :open="showCreateClass"
        :ontology-id="ontologyId"
        @close="showCreateClass = false"
        @created="onClassCreated"
      />
      <CreatePropertyDialog
        :open="showCreateProperty"
        :ontology-id="ontologyId"
        @close="showCreateProperty = false"
        @created="onPropertyCreated"
      />
      <CreateIndividualDialog
        :open="showCreateIndividual"
        :ontology-id="ontologyId"
        @close="showCreateIndividual = false"
        @created="onIndividualCreated"
      />
  </div>
</template>

<script setup lang="ts">
import AiSuggestionPanel from "@/components/ontology/AiSuggestionPanel.vue";
import ApplySequenceButton from "@/components/ontology/ApplySequenceButton.vue";
import BatchUploader from "@/components/ontology/BatchUploader.vue";
import CreateClassDialog from "@/components/ontology/CreateClassDialog.vue";
import CreateIndividualDialog from "@/components/ontology/CreateIndividualDialog.vue";
import CreatePropertyDialog from "@/components/ontology/CreatePropertyDialog.vue";
import DocumentUploader from "@/components/ontology/DocumentUploader.vue";
import SequencePreview from "@/components/ontology/SequencePreview.vue";
import GraphVisualization from "@/components/organisms/GraphVisualization.vue";
import { useI18n } from "@/composables/useI18n";
import {
	ChevronRight,
	FileText,
	Folder,
	GripVertical,
	MessageSquare,
	Plus,
	Search,
	Zap,
} from "@lucide/vue";
import { useQuery } from "@vue/apollo-composable";
import axios from "axios";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { generateFromText, refineSequence } from "../api/ai";
import type { AiGenerationResult } from "../api/ai";
import { CLASS_TREE_QUERY, LIST_INDIVIDUALS_QUERY } from "../apollo/queries";
import { useDraftState } from "../composables/useDraftState";
import type {
	AiSuggestion,
	ExtractionPreview,
	SequenceStep,
} from "../types/extraction";

const route = useRoute();
const { t } = useI18n();
const ontologyId = ref((route.params.id as string) || "default");
const selectedClassId = ref<string | null>(null);
const selectedIndividualId = ref<string | null>(null);
const viewMode = ref<"graph" | "table">("table");

// ── Save Draft state ────────────────────────────────────────────────────────

const draftState = useDraftState();
const saving = ref(false);

// ── AI Import Panel state ──────────────────────────────────────────────────

const showAiPanel = ref(false);
const aiTab = ref<"document" | "nl-to-owl">("document");
const uploadMode = ref<"single" | "batch">("single");

// ── NL→OWL Generation state ────────────────────────────────────────────────

const nlPrompt = ref("");
const isGenerating = ref(false);
const generationError = ref<string | null>(null);
const nlResult = ref<AiGenerationResult | null>(null);

// ── Iterative Refinement state ────────────────────────────────────────────

const refinementFeedback = ref("");
const isRefining = ref(false);
const refinementRound = ref(0);
const maxRefinementRounds = 5;

// ── Extraction / Preview state (shared) ───────────────────────────────────

const extractionSteps = ref<SequenceStep[]>([]);
const showApplyButton = ref(false);
const extractionSourceFiles = computed(
	() =>
		Array.from(
			new Set(
				extractionSteps.value.map((step) => step.sourceFile).filter(Boolean),
			),
		) as string[],
);

// ── Create Dialog state ───────────────────────────────────────────────────

const showCreateClass = ref(false);
const showCreateProperty = ref(false);
const showCreateIndividual = ref(false);

function openCreateClass() {
	showCreateClass.value = true;
}
function openCreateProperty() {
	showCreateProperty.value = true;
}
function openCreateIndividual() {
	showCreateIndividual.value = true;
}

function onClassCreated(_name: string) {
	showCreateClass.value = false;
	draftState.trackChange("create:class", null, _name);
}

function onPropertyCreated(_name: string) {
	showCreateProperty.value = false;
	draftState.trackChange("create:property", null, _name);
}

function onIndividualCreated(_name: string) {
	showCreateIndividual.value = false;
	draftState.trackChange("create:individual", null, _name);
}

function onAiSuggestionAccepted(suggestion: AiSuggestion) {
	console.info("[OntologyWorkspace] AI suggestion accepted", {
		id: suggestion.id,
		label: suggestion.label,
		type: suggestion.type,
	});
	draftState.trackChange("ai:suggest", null, suggestion.label);
}

function normalizeExtractionStep(
	step: SequenceStep,
	index: number,
): SequenceStep {
	return {
		...step,
		id:
			step.id ||
			`${step.operation}-${step.entityId || step.label || index}-${index}`,
		included: step.included ?? true,
	};
}

function onUploadComplete(result: ExtractionPreview) {
	console.info("[OntologyWorkspace] upload complete", {
		steps: result.steps.length,
	});
	extractionSteps.value = result.steps.map(normalizeExtractionStep);
	showApplyButton.value = true;
}

function onUploadError(error: { code: string; message: string } | string) {
	const errMsg = typeof error === "string" ? error : error.message;
	console.error("[OntologyWorkspace] upload error", {
		message: errMsg,
	});
}

function onApplySequence(steps: SequenceStep[]) {
	// Steps flow through to ApplySequenceButton
	console.debug("[OntologyWorkspace] ready to apply", {
		stepCount: steps.filter((s) => s.included).length,
	});
}

function onBatchComplete(result: { steps: SequenceStep[] }) {
	console.info("[OntologyWorkspace] batch complete", {
		steps: result.steps.length,
	});
	extractionSteps.value = result.steps.map(normalizeExtractionStep);
	showApplyButton.value = true;
}

function onBatchReset() {
	extractionSteps.value = [];
	showApplyButton.value = false;
}

function onApplyCancel() {
	extractionSteps.value = [];
	showApplyButton.value = false;
}

function onApplySuccess(result: { commitId?: string; commitUrl?: string }) {
	console.info("[OntologyWorkspace] apply success", {
		commitId: result.commitId,
	});
	// Keep the AI panel and success modal mounted so the user can inspect
	// the import result and commit link before explicitly closing it.
}

function onApplyError(error: string) {
	console.error("[OntologyWorkspace] apply error", { error });
}

// ── AI Tab switching ────────────────────────────────────────────────────────

function switchAiTab(tab: "document" | "nl-to-owl") {
	aiTab.value = tab;
	if (tab === "nl-to-owl") {
		uploadMode.value = "single"; // reset document mode when switching away
	}
}

// ── NL→OWL Generation ───────────────────────────────────────────────────────

async function handleGenerateFromText(): Promise<void> {
	const text = nlPrompt.value.trim();
	if (!text || isGenerating.value) return;

	isGenerating.value = true;
	generationError.value = null;
	console.info("[OntologyWorkspace] NL→OWL generation start", {
		ontologyId: ontologyId.value,
		textLength: text.length,
	});

	try {
		const result = await generateFromText({
			ontologyId: ontologyId.value,
			text,
		});

		nlResult.value = result;
		refinementRound.value = 0;
		refinementFeedback.value = "";

		// Convert AI result steps to extraction steps and show preview
		extractionSteps.value = result.steps.map((s, i) => ({
			...s,
			id: s.id || `ai-${i}`,
			included: s.included ?? true,
			sourceFile: undefined,
		}));
		showApplyButton.value = true;

		console.info("[OntologyWorkspace] NL→OWL generation complete", {
			steps: result.steps.length,
			id: result.id,
		});
	} catch (err) {
		const msg = err instanceof Error ? err.message : String(err);
		generationError.value = msg;
		console.error("[OntologyWorkspace] NL→OWL generation failed", {
			error: msg,
		});
	} finally {
		isGenerating.value = false;
	}
}

// ── Iterative Refinement ────────────────────────────────────────────────────

async function handleRefine(): Promise<void> {
	const feedback = refinementFeedback.value.trim();
	if (!feedback || isRefining.value || !nlResult.value) return;

	if (refinementRound.value >= maxRefinementRounds) {
		console.warn("[OntologyWorkspace] max refinement rounds reached");
		return;
	}

	isRefining.value = true;
	const currentRound = refinementRound.value + 1;
	console.info("[OntologyWorkspace] refinement start", {
		round: currentRound,
		feedbackLength: feedback.length,
	});

	try {
		const result = await refineSequence({
			ontologyId: ontologyId.value,
			sequenceId: nlResult.value.id,
			previousSteps: extractionSteps.value,
			feedback,
		});

		refinementRound.value = currentRound;
		refinementFeedback.value = "";

		// Update steps with refined results
		extractionSteps.value = result.steps.map((s, i) => ({
			...s,
			id: s.id || `refine-${currentRound}-${i}`,
			included: s.included ?? true,
			sourceFile: undefined,
		}));

		console.info("[OntologyWorkspace] refinement complete", {
			round: currentRound,
			steps: result.steps.length,
		});
	} catch (err) {
		const msg = err instanceof Error ? err.message : String(err);
		console.error("[OntologyWorkspace] refinement failed", {
			round: currentRound,
			error: msg,
		});
	} finally {
		isRefining.value = false;
	}
}

// ── Graph visualization data ───────────────────────────────────────────────────────────

// Recursively flatten the class tree so nested children appear as graph nodes.
// CLASS_TREE_QUERY returns a recursive { id, label, comment, children } shape;
// without flattening, only top-level classes would be visualized and subclass edges
// (and any individual→child-class edges) would silently drop.
function flattenClassTree(
	nodes: ReadonlyArray<{
		id: string;
		label: string;
		comment?: string | null;
		children?: unknown[];
	}>,
): Array<{
	id: string;
	label: string;
	comment?: string | null;
	parentId?: string;
}> {
	const out: Array<{
		id: string;
		label: string;
		comment?: string | null;
		parentId?: string;
	}> = [];
	const walk = (
		list: ReadonlyArray<{
			id: string;
			label: string;
			comment?: string | null;
			children?: unknown[];
		}>,
		parentId?: string,
	) => {
		for (const n of list) {
			out.push({ id: n.id, label: n.label, comment: n.comment, parentId });
			const kids = Array.isArray(n.children)
				? (n.children as Array<{
						id: string;
						label: string;
						comment?: string | null;
						children?: unknown[];
					}>)
				: [];
			if (kids.length) walk(kids, n.id);
		}
	};
	walk(nodes);
	return out;
}

// Flattened class list (all classes including nested children) for graph rendering
const flatClasses = computed(() => flattenClassTree(classTree.value));

// Currently selected class details (label + comment) for the detail panel
const selectedClass = computed(
	() => flatClasses.value.find((c) => c.id === selectedClassId.value) ?? null,
);

const graphNodes = computed(() => {
	const nodes: Array<{
		id: string;
		label: string;
		type: "class" | "property" | "individual";
		x: number;
		y: number;
	}> = [];
	let idx = 0;
	// Add classes as nodes (flattened — includes nested children)
	for (const cls of flatClasses.value) {
		nodes.push({
			id: cls.id,
			label: cls.label,
			type: "class",
			x: 50 + (idx % 5) * 200,
			y: 50 + Math.floor(idx / 5) * 80,
		});
		idx++;
	}
	// Add individuals as nodes
	for (const ind of individuals.value) {
		if (!ind.id) continue;
		nodes.push({
			id: ind.id,
			label: ind.label || ind.id,
			type: "individual",
			x: 50 + (idx % 5) * 200,
			y: 50 + Math.floor(idx / 5) * 80,
		});
		idx++;
	}
	return nodes;
});

const graphEdges = computed(() => {
	const edges: Array<{
		source: string;
		target: string;
		type: "subclass_of" | "object_property" | "datatype_property";
	}> = [];
	// Subclass edges: child → parent (keeps hierarchy visible even with no individuals)
	for (const cls of flatClasses.value) {
		if (cls.parentId) {
			edges.push({ source: cls.id, target: cls.parentId, type: "subclass_of" });
		}
	}
	// Individual → class instance edges
	for (const ind of individuals.value) {
		if (ind.classId) {
			edges.push({ source: ind.id, target: ind.classId, type: "subclass_of" });
		}
	}
	return edges;
});

function onGraphNodeClick(node: { id: string; label: string; type: string }) {
	if (node.type === "individual") {
		selectedIndividualId.value = node.id;
	} else if (node.type === "class") {
		selectedClassId.value = node.id;
	}
}

// ── Ontology metadata (REST — migrated from ONTOLOGY_QUERY) ────────────────────

interface OntologyMeta {
	id: string;
	name?: string;
	branch?: string;
	dirty?: boolean;
}

const loading = ref(false);
const error = ref<string | null>(null);
const ontologyData = ref<OntologyMeta | null>(null);

async function fetchOntologyMeta() {
	loading.value = true;
	error.value = null;
	try {
		const token = localStorage.getItem("vedo-jwt-token");
		const { data } = await axios.get(`/api/v1/ontologies/${ontologyId.value}`, {
			headers: token ? { Authorization: `Bearer ${token}` } : {},
		});
		ontologyData.value = data;
	} catch (e: unknown) {
		const err = e as {
			response?: { data?: { error?: { message?: string } } };
			message?: string;
		};
		error.value =
			err.response?.data?.error?.message ??
			err.message ??
			"Failed to load ontology";
	} finally {
		loading.value = false;
	}
}

onMounted(() => {
	fetchOntologyMeta();
});

// Set ontology context when metadata loads for draft state
watch(ontologyData, (data) => {
	if (data?.id) {
		draftState.setOntologyContext(data.id);
		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "workspace.draft.context_set",
				ontologyId: data.id,
				ts: new Date().toISOString(),
			}),
		);
	}
});

async function handleSave(): Promise<void> {
	if (saving.value || !draftState.hasUnsavedChanges.value) return;
	saving.value = true;
	try {
		const success = await draftState.saveDraft();
		if (success) {
			console.debug(
				JSON.stringify({
					level: "debug",
					msg: "workspace.draft.save_success",
					ts: new Date().toISOString(),
				}),
			);
		} else {
			console.warn(
				JSON.stringify({
					level: "warn",
					msg: "workspace.draft.save_no_changes",
					ts: new Date().toISOString(),
				}),
			);
		}
	} catch (err) {
		const msg = err instanceof Error ? err.message : String(err);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "workspace.draft.save_failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		saving.value = false;
	}
}

// ── Class tree ───────────────────────────────────────────────────────────────────────

const { result: classTreeResult } = useQuery(
	CLASS_TREE_QUERY,
	() => ({ ontologyId: ontologyId.value }),
	{
		fetchPolicy: "cache-and-network",
	},
);

const classTree = computed(() => classTreeResult.value?.classTree || []);

// ── Individuals by class ─────────────────────────────────────────────────────────────

const { result: individualsResult, refetch: refetchIndividuals } = useQuery(
	LIST_INDIVIDUALS_QUERY,
	() => ({
		ontologyId: ontologyId.value,
		classId: selectedClassId.value || "",
		page: 0,
		perPage: 50,
	}),
	{
		fetchPolicy: "cache-and-network",
		enabled: computed(() => !!selectedClassId.value),
	},
);

const individuals = computed(
	() => individualsResult.value?.individuals?.items || [],
);

// ── Selection handling ────────────────────────────────────────────────────────────────

function selectClass(classId: string) {
	selectedClassId.value = classId;
	selectedIndividualId.value = null;
}

watch(selectedClassId, () => {
	if (selectedClassId.value) {
		refetchIndividuals();
	}
});
</script>

<style scoped>
.workspace-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.workspace-loading,
.workspace-error {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  font-family: 'IBM Plex Mono', monospace;
  color: var(--muted-foreground);
}

.workspace-main {
  flex: 1;
  min-height: 0;
  display: flex;
}

.card-side {
  background: var(--card);
}

.group-sidebar {
  width: 256px;
  border-right: 1px solid var(--border);
  padding: 14px;
}

.group-header {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 10px;
  color: var(--foreground);
}

.splitter {
  width: 8px;
  background: #0f0f0f;
  border-left: 1px solid #2a2a2a;
  border-right: 1px solid #2a2a2a;
  color: #4b5563;
  display: flex;
  align-items: center;
  justify-content: center;
}

.workspace-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.workspace-toolbar {
  height: 56px;
  border-bottom: 1px solid var(--border);
  background: var(--card);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}

.toolbar-title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
}

.toolbar-branch-badge {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  background: var(--surface-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  color: var(--primary);
}

.toolbar-dirty-badge {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  background: rgba(245, 158, 11, 0.15);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  color: #f59e0b;
}

.toolbar-spacer { flex: 1; }

.toolbar-btn--active {
  background: var(--primary);
  color: var(--primary-foreground);
}

.toolbar-btn {
  height: 32px;
  border-radius: 6px;
  border: 1px solid var(--border);
  padding: 0 10px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.toolbar-btn--primary {
  background: var(--primary);
  border-color: var(--primary);
  color: var(--primary-foreground);
}

.workspace-row {
  flex: 1;
  display: flex;
  min-height: 0;
}

.class-panel {
  width: 320px;
  border-right: 1px solid var(--border);
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.panel-tools {
  display: flex;
  align-items: center;
  gap: 6px;
}

.panel-input {
  flex: 1;
  height: 32px;
  border: 1px solid #2a2a2a;
  background: #101010;
  color: #6b7280;
  border-radius: 6px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.class-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.class-row {
  border-radius: 4px;
  padding: 6px 8px;
  display: flex;
  align-items: center;
  gap: 6px;
  color: #fafafa;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  width: 100%;
}

.class-row:hover {
  background-color: var(--surface-secondary);
}

.class-row--active {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.45);
  color: var(--primary);
  font-weight: 600;
}

.graph-panel {
  flex: 1;
  min-width: 0;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  position: relative;
}

.detail-panel {
  position: absolute;
  top: 50px;
  right: 12px;
  z-index: 10;
  min-width: 200px;
  max-width: 300px;
  padding: 12px 16px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  box-shadow: var(--shadow-md, 0 4px 12px rgba(0, 0, 0, 0.15));
}

.detail-panel__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.detail-panel__title {
  font-weight: 600;
  font-size: 14px;
}

.detail-panel__comment {
  font-size: 13px;
  color: var(--muted-foreground, #666);
  line-height: 1.4;
}

.graph-head,
.graph-row {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  font-family: 'IBM Plex Mono', monospace;
}

.graph-head {
  color: var(--muted-foreground);
  font-size: 11px;
  border-bottom: 1px solid var(--border);
}

.graph-row {
  font-size: 12px;
  border-bottom: 1px solid rgba(42, 42, 42, 0.5);
  cursor: pointer;
}

.graph-row:hover {
  background: var(--surface-secondary);
}

.graph-row--active {
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.4);
}

.graph-empty {
  padding: 16px;
  text-align: center;
  font-size: 12px;
}

.graph-panel__toolbar {
  display: flex;
  gap: 4px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-default);
}

.graph-panel__toggle {
  font-size: var(--font-size-xs);
  padding: 4px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-primary);
  cursor: pointer;
  color: var(--text-secondary);
}

.graph-panel__toggle--active {
  background: var(--primary);
  color: var(--primary-foreground);
  border-color: var(--primary);
}

.col-ind { width: 220px; }
.col-prop { width: 240px; }
.col-val { flex: 1; }

.property-panel {
  width: 372px;
  border-left: 1px solid var(--border);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 600;
}

.indiv-head,
.indiv-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  font-family: 'IBM Plex Mono', monospace;
}

.indiv-head {
  color: var(--muted-foreground);
  font-size: 11px;
}

.indiv-row {
  border-radius: 4px;
  font-size: 12px;
}

.indiv-row:hover {
  background: var(--surface-secondary);
}

.i-name { width: 170px; }
.i-type { width: 100px; }

.i-type--active { color: var(--primary); font-weight: 600; }

.muted { color: var(--muted-foreground); }

/* ── AI Import Panel ────────────────────────────────────────────────────── */

.workspace-ai-panel {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-6, 24px);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4, 16px);
}

.ai-panel__header {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1, 4px);
}

.ai-panel__title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: var(--font-size-lg, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #fafafa);
  margin: 0;
}

.ai-panel__desc {
  font-size: var(--font-size-sm, 13px);
  color: var(--text-muted, #64748b);
  margin: 0;
}

.ai-panel__mode-toggle {
  display: flex;
  gap: var(--spacing-2, 8px);
}

.ai-panel__mode-btn {
  padding: var(--spacing-1, 4px) var(--spacing-3, 12px);
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-secondary, #6b7280);
  font-size: var(--font-size-sm, 13px);
  cursor: pointer;
  transition: all var(--transition-fast, 0.15s);
}

.ai-panel__mode-btn:hover {
  border-color: var(--primary, #10b981);
  color: var(--text-primary, #fafafa);
}

.ai-panel__mode-btn--active {
  background: var(--primary, #10b981);
  border-color: var(--primary, #10b981);
  color: var(--text-inverse, #0a0a0a);
}

.ai-panel__apply {
  display: flex;
  justify-content: flex-start;
}

/* ── Toolbar Create Dropdown ────────────────────────── */

.toolbar-create-group {
  position: relative;
  display: inline-flex;
}

.toolbar-create-btn {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-create-group:hover .toolbar-create-dropdown {
  display: flex;
}

.toolbar-create-dropdown {
  display: none;
  position: absolute;
  top: 100%;
  left: 0;
  flex-direction: column;
  background: var(--surface-secondary);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 4px;
  z-index: 100;
  min-width: 140px;
  margin-top: 2px;
}

.toolbar-create-dropdown button {
  padding: 6px 12px;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
}

.toolbar-create-dropdown button:hover {
  background: var(--surface-tertiary);
}

/* ── Tab bar ─────────────────────────────────────────── */

.ai-panel__tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border);
  margin-bottom: var(--spacing-3, 12px);
}

.ai-panel__tab {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 6px);
  padding: var(--spacing-2, 8px) var(--spacing-4, 16px);
  border: none;
  background: transparent;
  color: var(--text-muted, #6b7280);
  font-size: var(--font-size-sm, 13px);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: all var(--transition-fast, 0.15s);
}

.ai-panel__tab:hover {
  color: var(--text-primary, #fafafa);
}

.ai-panel__tab--active {
  color: var(--primary, #10b981);
  border-bottom-color: var(--primary, #10b981);
}

/* ── NL Prompt (NL→OWL tab) ─────────────────────────── */

.nl-prompt {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2, 8px);
}

.nl-prompt__label {
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  color: var(--text-primary, #fafafa);
}

.nl-prompt__textarea {
  width: 100%;
  min-height: 80px;
  padding: var(--spacing-2, 8px) var(--spacing-3, 12px);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-primary, #101010);
  color: var(--text-primary, #fafafa);
  font-family: 'IBM Plex Mono', monospace;
  font-size: var(--font-size-sm, 13px);
  resize: vertical;
  line-height: 1.5;
}

.nl-prompt__textarea:focus {
  outline: none;
  border-color: var(--primary, #10b981);
}

.nl-prompt__textarea:disabled {
  opacity: 0.5;
}

.nl-prompt__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.nl-prompt__hint {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #6b7280);
}

.nl-prompt__generate-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 6px);
  padding: var(--spacing-2, 8px) var(--spacing-4, 16px);
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--primary, #10b981);
  background: var(--primary, #10b981);
  color: var(--text-inverse, #0a0a0a);
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  cursor: pointer;
  transition: background var(--transition-fast, 0.15s);
}

.nl-prompt__generate-btn:hover:not(:disabled) {
  background: var(--primary-hover, #34d399);
}

.nl-prompt__generate-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ── Generation Error ───────────────────────────────── */

.nl-error {
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

.nl-error__icon {
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

.nl-error__retry {
  margin-left: auto;
  padding: var(--spacing-1, 4px) var(--spacing-2, 8px);
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-secondary, #6b7280);
  font-size: var(--font-size-xs, 12px);
  cursor: pointer;
}

.nl-error__retry:hover {
  background: var(--surface-secondary, #141414);
}

/* ── Refinement section ─────────────────────────────── */

.nl-refinement {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2, 6px);
  padding: var(--spacing-3, 12px);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-secondary, #141414);
}

.nl-refinement__label {
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  color: var(--text-primary, #fafafa);
}

.nl-refinement__row {
  display: flex;
  gap: var(--spacing-2, 8px);
}

.nl-refinement__input {
  flex: 1;
  height: 36px;
  padding: 0 var(--spacing-3, 12px);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-primary, #101010);
  color: var(--text-primary, #fafafa);
  font-family: 'IBM Plex Mono', monospace;
  font-size: var(--font-size-sm, 13px);
}

.nl-refinement__input:focus {
  outline: none;
  border-color: var(--primary, #10b981);
}

.nl-refinement__input:disabled {
  opacity: 0.5;
}

.nl-refinement__btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 6px);
  padding: 0 var(--spacing-3, 12px);
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--primary, #10b981);
  background: var(--primary, #10b981);
  color: var(--text-inverse, #0a0a0a);
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  cursor: pointer;
  white-space: nowrap;
}

.nl-refinement__btn:hover:not(:disabled) {
  background: var(--primary-hover, #34d399);
}

.nl-refinement__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.nl-refinement__round {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #6b7280);
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
  to { transform: rotate(360deg); }
}

@media (max-width: 1280px) {
  .group-sidebar,
  .property-panel { display: none; }
  .splitter { display: none; }
  .class-panel { width: 280px; }
}

@media (max-width: 900px) {
  .class-panel { display: none; }
  .workspace-toolbar { height: auto; padding: 10px 12px; flex-wrap: wrap; }
}
</style>
