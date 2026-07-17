<!-- @ctx: Ontology workspace page — wired to real backend via Apollo GraphQL queries -->
<template>
  <div class="workspace-page" role="main" aria-label="Ontology Workspace content">
    <div v-if="loading" class="workspace-loading">
      <span>Loading ontology...</span>
    </div>

    <div v-else-if="error" class="workspace-error">
      <span>Failed to load ontology: {{ error }}</span>
    </div>

    <template v-else>
      <div class="workspace-main">
        <aside class="group-sidebar card-side">
          <div class="group-header">Project group</div>
        </aside>

        <div class="splitter"><GripVertical :size="8" /></div>

        <section class="workspace-content">
            <div class="workspace-toolbar">
              <span class="toolbar-title">{{ ontologyData?.name || ontologyId }}</span>
              <span v-if="ontologyData?.branch" class="toolbar-branch-badge">{{ ontologyData.branch }}</span>
              <span v-if="ontologyData?.dirty" class="toolbar-dirty-badge">Dirty</span>
              <span class="toolbar-spacer"></span>
              <button class="toolbar-btn" type="button">Publish</button>
              <button class="toolbar-btn toolbar-btn--primary" type="button">Save</button>
              <button
                :class="['toolbar-btn', { 'toolbar-btn--active': showAiPanel }]"
                type="button"
                @click="showAiPanel = !showAiPanel"
              >
                <Zap :size="14" />
                AI Import
              </button>
            </div>

            <!-- AI Import Panel (toggleable) -->
            <div v-if="showAiPanel" class="workspace-ai-panel">
              <div class="ai-panel__header">
                <h3 class="ai-panel__title">AI-Assisted Ontology Import</h3>
                <p class="ai-panel__desc">Upload documents to extract ontology classes, properties, and individuals.</p>
              </div>

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

              <!-- Mode toggle -->
              <div class="ai-panel__mode-toggle">
                <button
                  :class="['ai-panel__mode-btn', { 'ai-panel__mode-btn--active': uploadMode === 'single' }]"
                  type="button"
                  @click="uploadMode = 'single'"
                >
                  Single file
                </button>
                <button
                  :class="['ai-panel__mode-btn', { 'ai-panel__mode-btn--active': uploadMode === 'batch' }]"
                  type="button"
                  @click="uploadMode = 'batch'"
                >
                  Batch upload
                </button>
              </div>

              <!-- Sequence Preview (after upload) -->
              <div v-if="extractionSteps.length > 0" class="ai-panel__section">
                <SequencePreview
                  :steps="extractionSteps"
                  :ontology-id="ontologyId"
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
                <div class="panel-input">Filter classes...</div>
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
                  Graph
                </button>
                <button
                  class="graph-panel__toggle"
                  :class="{ 'graph-panel__toggle--active': viewMode === 'table' }"
                  @click="viewMode = 'table'"
                >
                  Table
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

              <!-- Table view (existing) -->
              <template v-else>
                <div class="graph-head">
                  <span class="col-ind">Individual</span>
                  <span class="col-prop">Property</span>
                  <span class="col-val">Value</span>
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
                  <span class="muted">No individuals. Select a class to browse.</span>
                </div>
              </template>
            </section>

            <div class="splitter"><GripVertical :size="8" /></div>

            <aside class="property-panel card-side">
              <div class="panel-title">Individuals</div>
              <div class="panel-tools">
                <Search :size="14" class="muted" />
                <div class="panel-input">Filter individuals...</div>
              </div>
              <div class="indiv-head"><span class="i-name">Name</span><span class="i-type">Type</span></div>
              <div v-for="item in individuals" :key="item.id" class="indiv-row">
                <span class="i-name">{{ item.label }}</span>
                <span class="i-type i-type--active">{{ item.classLabel }}</span>
              </div>
            </aside>
          </div>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import ApplySequenceButton from '@/components/ontology/ApplySequenceButton.vue'
import BatchUploader from '@/components/ontology/BatchUploader.vue'
import DocumentUploader from '@/components/ontology/DocumentUploader.vue'
import SequencePreview from '@/components/ontology/SequencePreview.vue'
import GraphVisualization from '@/components/organisms/GraphVisualization.vue'
import { useQuery } from '@vue/apollo-composable'
import { ChevronRight, Folder, GripVertical, Search, Zap } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { CLASS_TREE_QUERY, LIST_INDIVIDUALS_QUERY, ONTOLOGY_QUERY } from '../apollo/queries'
import type { ExtractionPreview, SequenceStep } from '../types/extraction'

const route = useRoute()
const ontologyId = ref((route.params.id as string) || 'default')
const selectedClassId = ref<string | null>(null)
const selectedIndividualId = ref<string | null>(null)
const viewMode = ref<'graph' | 'table'>('table')

// ── AI Import Panel state ──────────────────────────────────────────────────

const showAiPanel = ref(false)
const uploadMode = ref<'single' | 'batch'>('single')
const extractionSteps = ref<SequenceStep[]>([])
const showApplyButton = ref(false)

function onUploadComplete(result: ExtractionPreview) {
  console.info('[OntologyWorkspace] upload complete', {
    steps: result.steps.length
  })
  extractionSteps.value = result.steps.map((s) => ({ ...s, included: true }))
  showApplyButton.value = true
}

function onUploadError(error: { code: string; message: string } | string) {
  const errMsg = typeof error === 'string' ? error : error.message
  console.error('[OntologyWorkspace] upload error', {
    message: errMsg
  })
}

function onApplySequence(steps: SequenceStep[]) {
  // Steps flow through to ApplySequenceButton
  console.debug('[OntologyWorkspace] ready to apply', {
    stepCount: steps.filter((s) => s.included).length
  })
}

function onBatchComplete(result: { steps: SequenceStep[] }) {
  console.info('[OntologyWorkspace] batch complete', {
    steps: result.steps.length
  })
  extractionSteps.value = result.steps.map((s) => ({ ...s, included: true }))
  showApplyButton.value = true
}

function onBatchReset() {
  extractionSteps.value = []
  showApplyButton.value = false
}

function onApplyCancel() {
  extractionSteps.value = []
  showApplyButton.value = false
}

function onApplySuccess(result: { commitId?: string; commitUrl?: string }) {
  console.info('[OntologyWorkspace] apply success', {
    commitId: result.commitId
  })
  // Reset after successful apply
  extractionSteps.value = []
  showApplyButton.value = false
  showAiPanel.value = false
}

function onApplyError(error: string) {
  console.error('[OntologyWorkspace] apply error', { error })
}

// ── Graph visualization data ───────────────────────────────────────────────────────────

const graphNodes = computed(() => {
  const nodes: Array<{
    id: string
    label: string
    type: 'class' | 'property' | 'individual'
    x: number
    y: number
  }> = []
  let idx = 0
  // Add classes as nodes
  for (const cls of classTree.value) {
    nodes.push({
      id: cls.id,
      label: cls.label,
      type: 'class',
      x: 50 + (idx % 5) * 200,
      y: 50 + Math.floor(idx / 5) * 80
    })
    idx++
  }
  // Add individuals as nodes
  for (const ind of individuals.value) {
    if (!ind.id) continue
    nodes.push({
      id: ind.id,
      label: ind.label || ind.id,
      type: 'individual',
      x: 50 + (idx % 5) * 200,
      y: 50 + Math.floor(idx / 5) * 80
    })
    idx++
  }
  return nodes
})

const graphEdges = computed(() => {
  const edges: Array<{
    source: string
    target: string
    type: 'subclass_of' | 'object_property' | 'datatype_property'
  }> = []
  for (const ind of individuals.value) {
    if (ind.classId) {
      edges.push({ source: ind.id, target: ind.classId, type: 'subclass_of' })
    }
  }
  return edges
})

function onGraphNodeClick(node: { id: string; label: string; type: string }) {
  if (node.type === 'individual') {
    selectedIndividualId.value = node.id
  } else if (node.type === 'class') {
    selectedClassId.value = node.id
  }
}

// ── Ontology metadata ────────────────────────────────────────────────────────────────

const {
  result: ontologyResult,
  loading,
  error
} = useQuery(ONTOLOGY_QUERY, () => ({ id: ontologyId.value }), {
  fetchPolicy: 'cache-and-network'
})

const ontologyData = computed(() => ontologyResult.value?.ontology)

// ── Class tree ───────────────────────────────────────────────────────────────────────

const { result: classTreeResult } = useQuery(
  CLASS_TREE_QUERY,
  () => ({ ontologyId: ontologyId.value }),
  {
    fetchPolicy: 'cache-and-network'
  }
)

const classTree = computed(() => classTreeResult.value?.classTree || [])

// ── Individuals by class ─────────────────────────────────────────────────────────────

const { result: individualsResult, refetch: refetchIndividuals } = useQuery(
  LIST_INDIVIDUALS_QUERY,
  () => ({
    ontologyId: ontologyId.value,
    classId: selectedClassId.value || '',
    page: 0,
    perPage: 50
  }),
  {
    fetchPolicy: 'cache-and-network',
    enabled: computed(() => !!selectedClassId.value)
  }
)

const individuals = computed(() => individualsResult.value?.individuals?.items || [])

// ── Selection handling ────────────────────────────────────────────────────────────────

function selectClass(classId: string) {
  selectedClassId.value = classId
  selectedIndividualId.value = null
}

watch(selectedClassId, () => {
  if (selectedClassId.value) {
    refetchIndividuals()
  }
})
</script>

<style scoped>
.workspace-page {
  min-height: calc(100vh - 56px);
}

.workspace-loading,
.workspace-error {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  font-family: 'IBM Plex Mono', monospace;
  color: var(--muted-foreground);
}

.workspace-main {
  display: flex;
  height: calc(100vh - 56px);
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

.ai-panel__section {
  /* Section wrapper */
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
