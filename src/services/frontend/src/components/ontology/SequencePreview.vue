<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism SequencePreview — interactive table for reviewing and editing extracted ontology steps -->
<template>
  <div class="preview" role="region" aria-label="Extraction sequence preview">
    <!-- Header with summary -->
    <div class="preview__header">
      <h3 class="preview__title">Extraction Preview</h3>
      <span class="preview__counter">
        Will be imported: <strong>{{ includedCount }}</strong> of <strong>{{ steps.length }}</strong> steps
      </span>
    </div>

    <!-- Empty state -->
    <div v-if="steps.length === 0" class="preview__empty">
      <p class="preview__empty-text">No extraction steps to preview. Upload a document to get started.</p>
    </div>

    <template v-else>
      <!-- Group by source file (batch mode) -->
      <template v-if="sourceFiles && sourceFiles.length > 1">
        <div
          v-for="group in groupedSteps"
          :key="group.sourceFile"
          class="preview__group"
        >
          <button
            class="preview__group-header"
            type="button"
            :aria-expanded="group.expanded"
            @click="toggleGroup(group.sourceFile)"
          >
            <ChevronRight :size="14" :class="['preview__group-chevron', { 'preview__group-chevron--open': group.expanded }]" />
            <FileText :size="14" />
            <span class="preview__group-title">{{ group.sourceFile }}</span>
            <span class="preview__group-count">{{ group.steps.length }} steps</span>
          </button>

          <div v-if="group.expanded" class="preview__group-body">
            <!-- Column headers -->
            <div class="preview__column-headers">
              <span class="preview__col-checkbox"></span>
              <span class="preview__col-operation">Operation</span>
              <span class="preview__col-entity">Entity ID</span>
              <span class="preview__col-label">Label</span>
              <span class="preview__col-context">Parent/Range</span>
              <span class="preview__col-duplicate"></span>
            </div>

            <!-- Rows -->
            <SequencePreviewRow
              v-for="step in group.steps"
              :key="step.id"
              :step="step"
              @toggle-include="onToggleInclude"
              @update-label="onUpdateLabel"
            />
          </div>
        </div>
      </template>

      <!-- Flat list (single file mode) -->
      <template v-else>
        <!-- Column headers -->
        <div class="preview__column-headers">
          <span class="preview__col-checkbox"></span>
          <span class="preview__col-operation">Operation</span>
          <span class="preview__col-entity">Entity ID</span>
          <span class="preview__col-label">Label</span>
          <span class="preview__col-context">Parent/Range</span>
          <span class="preview__col-duplicate"></span>
        </div>

        <SequencePreviewRow
          v-for="step in steps"
          :key="step.id"
          :step="step"
          @toggle-include="onToggleInclude"
          @update-label="onUpdateLabel"
        />
      </template>

      <!-- Footer actions -->
      <div class="preview__footer">
        <div class="preview__footer-actions">
          <button class="preview__select-all-btn" type="button" @click="selectAll">
            Select all
          </button>
          <button class="preview__deselect-all-btn" type="button" @click="deselectAll">
            Deselect all
          </button>
          <span class="preview__separator">|</span>
          <span class="preview__duplicates-info" v-if="duplicateCount > 0">
            {{ duplicateCount }} duplicate{{ duplicateCount > 1 ? 's' : '' }} detected
          </span>
        </div>

        <div class="preview__footer-apply">
          <button
            class="preview__apply-btn"
            type="button"
            :disabled="includedCount === 0"
            @click="$emit('apply', filteredIncluded)"
          >
            Apply Import ({{ includedCount }})
          </button>
          <button
            class="preview__cancel-btn"
            type="button"
            @click="$emit('cancel')"
          >
            Cancel
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ChevronRight, FileText } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import type { SequenceStep } from '../../types/extraction'
import SequencePreviewRow from './SequencePreviewRow.vue'

const props = defineProps<{
  steps: SequenceStep[]
  ontologyId: string
  sourceFiles?: string[]
}>()

defineEmits<{
  apply: [steps: SequenceStep[]]
  cancel: []
}>()

// ── Reactive state for local step modifications ─────────────────────────────

const localSteps = ref<SequenceStep[]>(props.steps.map((s) => ({ ...s })))

// ── Computed ────────────────────────────────────────────────────────────────

const includedCount = computed(() => localSteps.value.filter((s) => s.included).length)
const duplicateCount = computed(() => localSteps.value.filter((s) => s.isDuplicate).length)
const filteredIncluded = computed(() => localSteps.value.filter((s) => s.included))

// Group by source file for batch mode
interface StepGroup {
  sourceFile: string
  steps: SequenceStep[]
  expanded: boolean
}

const expandedGroups = ref<Set<string>>(new Set())

const groupedSteps = computed<StepGroup[]>(() => {
  const groups = new Map<string, SequenceStep[]>()
  for (const step of localSteps.value) {
    const key = step.sourceFile || 'unknown'
    if (!groups.has(key)) {
      groups.set(key, [])
    }
    groups.get(key)?.push(step)
  }

  return Array.from(groups.entries()).map(([sourceFile, steps]) => ({
    sourceFile,
    steps,
    expanded: expandedGroups.value.has(sourceFile)
  }))
})

// ── Handlers ────────────────────────────────────────────────────────────────

function onToggleInclude(stepId: string) {
  const step = localSteps.value.find((s) => s.id === stepId)
  if (step) {
    step.included = !step.included
  }
}

function onUpdateLabel(stepId: string, label: string) {
  const step = localSteps.value.find((s) => s.id === stepId)
  if (step) {
    step.label = label
  }
}

function toggleGroup(sourceFile: string) {
  if (expandedGroups.value.has(sourceFile)) {
    expandedGroups.value.delete(sourceFile)
  } else {
    expandedGroups.value.add(sourceFile)
  }
}

function selectAll() {
  console.debug('[SequencePreview] select all steps')
  for (const step of localSteps.value) {
    step.included = true
  }
}

function deselectAll() {
  console.debug('[SequencePreview] deselect all steps')
  for (const step of localSteps.value) {
    step.included = false
  }
}
</script>

<style scoped>
.preview {
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--surface-primary);
  overflow: hidden;
}

.preview__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border);
  background: var(--surface-secondary);
}

.preview__title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  margin: 0;
  color: var(--text-primary);
}

.preview__counter {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}

.preview__counter strong {
  color: var(--text-primary);
}

.preview__empty {
  padding: var(--spacing-8);
  text-align: center;
}

.preview__empty-text {
  color: var(--text-muted);
  font-size: var(--font-size-sm);
  margin: 0;
}

.preview__group {
  border-bottom: 1px solid var(--border);
}

.preview__group:last-child {
  border-bottom: none;
}

.preview__group-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  width: 100%;
  padding: var(--spacing-2) var(--spacing-4);
  background: var(--surface-secondary);
  border: none;
  cursor: pointer;
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  text-align: left;
}

.preview__group-header:hover {
  background: var(--surface-tertiary);
}

.preview__group-chevron {
  transition: transform 0.2s;
  color: var(--text-muted);
}

.preview__group-chevron--open {
  transform: rotate(90deg);
}

.preview__group-title {
  flex: 1;
}

.preview__group-count {
  color: var(--text-muted);
  font-size: var(--font-size-xs);
}

.preview__group-body {
  /* Nested inside group - rows render here */
}

.preview__column-headers {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--surface-primary);
  border-bottom: 2px solid var(--border);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.preview__col-checkbox {
  width: 36px;
  flex-shrink: 0;
}

.preview__col-operation {
  width: 80px;
  flex-shrink: 0;
}

.preview__col-entity {
  width: 180px;
  flex-shrink: 0;
}

.preview__col-label {
  flex: 1;
}

.preview__col-context {
  width: 120px;
  flex-shrink: 0;
}

.preview__col-duplicate {
  width: 90px;
  flex-shrink: 0;
}

.preview__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-top: 1px solid var(--border);
  background: var(--surface-secondary);
}

.preview__footer-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
}

.preview__select-all-btn,
.preview__deselect-all-btn {
  background: none;
  border: none;
  color: var(--primary);
  cursor: pointer;
  font-size: var(--font-size-sm);
  padding: 0;
}

.preview__select-all-btn:hover,
.preview__deselect-all-btn:hover {
  text-decoration: underline;
}

.preview__separator {
  color: var(--border);
}

.preview__duplicates-info {
  color: var(--warning);
  font-size: var(--font-size-xs);
}

.preview__footer-apply {
  display: flex;
  gap: var(--spacing-2);
}

.preview__apply-btn {
  padding: var(--spacing-2) var(--spacing-4);
  border-radius: var(--radius-md);
  border: 1px solid var(--primary);
  background: var(--primary);
  color: var(--text-inverse);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.preview__apply-btn:hover:not(:disabled) {
  background: var(--primary-hover);
}

.preview__apply-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.preview__cancel-btn {
  padding: var(--spacing-2) var(--spacing-4);
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  cursor: pointer;
}

.preview__cancel-btn:hover {
  background: var(--surface-secondary);
}
</style>
