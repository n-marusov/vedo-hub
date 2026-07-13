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
          </div>

          <div class="workspace-row">
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
import { computed, ref, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { useRoute } from 'vue-router'
import {
  ChevronRight,
  Columns3,
  Folder,
  GripVertical,
  Pencil,
  Plus,
  Search,
  Share2,
  Table2,
  Trash2,
  User,
  Workflow
} from 'lucide-vue-next'
import {
  ONTOLOGY_QUERY,
  CLASS_TREE_QUERY,
  LIST_INDIVIDUALS_QUERY
} from '../apollo/queries'

const route = useRoute()
const ontologyId = ref((route.params.id as string) || 'default')
const selectedClassId = ref<string | null>(null)
const selectedIndividualId = ref<string | null>(null)

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

const {
  result: classTreeResult,
  loading: treeLoading
} = useQuery(CLASS_TREE_QUERY, () => ({ ontologyId: ontologyId.value }), {
  fetchPolicy: 'cache-and-network'
})

const classTree = computed(() => classTreeResult.value?.classTree || [])

// ── Individuals by class ─────────────────────────────────────────────────────────────

const {
  result: individualsResult,
  loading: individualsLoading,
  refetch: refetchIndividuals
} = useQuery(LIST_INDIVIDUALS_QUERY, () => ({
  ontologyId: ontologyId.value,
  classId: selectedClassId.value || '',
  page: 0,
  perPage: 50
}), {
  fetchPolicy: 'cache-and-network',
  enabled: computed(() => !!selectedClassId.value)
})

const individuals = computed(() => individualsResult.value?.individuals?.items || [])

// ── Selection handling ────────────────────────────────────────────────────────────────

function selectClass(classId: string) {
  selectedClassId.value = classId
  selectedIndividualId.value = null
}

function selectIndividual(individualId: string) {
  selectedIndividualId.value = individualId
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
