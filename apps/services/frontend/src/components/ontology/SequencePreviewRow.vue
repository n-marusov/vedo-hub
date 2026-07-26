<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism SequencePreviewRow — single step row with include/exclude toggle, inline edit, source info -->
<template>
  <div
    :class="['preview-row', {
      'preview-row--excluded': !step.included,
      'preview-row--duplicate': step.isDuplicate
    }]"
    role="row"
    :aria-label="`Step: ${step.label}`"
  >
    <!-- Include/Exclude checkbox -->
    <div class="preview-row__checkbox" role="gridcell">
      <label class="preview-row__toggle" :for="checkboxId">
        <input
          :id="checkboxId"
          type="checkbox"
          :checked="step.included"
          class="preview-row__toggle-input"
          @change="toggleInclude"
        />
        <span class="preview-row__toggle-box" aria-hidden="true"></span>
      </label>
    </div>

    <!-- Operation icon -->
    <div class="preview-row__operation" role="gridcell" :title="step.operation">
      <span :class="['preview-row__op-badge', `preview-row__op-badge--${operationColor}`]">
        {{ operationLabel }}
      </span>
    </div>

    <!-- Entity ID -->
    <div class="preview-row__entity-id" role="gridcell" :title="step.entityId">
      <code class="preview-row__id-text">{{ step.entityId }}</code>
    </div>

    <!-- Label (editable inline) -->
    <div class="preview-row__label" role="gridcell">
      <div v-if="editing" class="preview-row__edit">
        <input
          ref="editInputRef"
          :value="editValue"
          class="preview-row__edit-input"
          @input="editValue = ($event.target as HTMLInputElement).value"
          @keyup.enter="saveEdit"
          @keyup.escape="cancelEdit"
          @blur="saveEdit"
        />
      </div>
      <button
        v-else
        class="preview-row__label-text"
        type="button"
        :title="'Edit label'"
        @click="startEdit"
      >
        {{ step.label || step.entityId }}
        <Pencil :size="12" class="preview-row__edit-icon" />
      </button>
    </div>

    <!-- Parent / Domain / Range -->
    <div class="preview-row__context" role="gridcell">
      <span v-if="step.parentLabel" class="preview-row__context-tag" :title="`Parent: ${step.parentLabel}`">
        {{ step.parentLabel }}
      </span>
      <span v-else-if="step.domain" class="preview-row__context-tag" :title="`Domain: ${step.domain}`">
        {{ step.domain }}
      </span>
      <span v-else class="preview-row__context-empty">—</span>
    </div>

    <!-- Duplicate warning badge -->
    <div v-if="step.isDuplicate" class="preview-row__duplicate" role="gridcell">
      <span class="badge badge--warning" :title="`⚠ ${step.label} already exists${step.duplicateOf ? ' (duplicate of ' + step.duplicateOf + ')' : ''}`">
        ⚠ Duplicate
      </span>
    </div>

    <!-- Source file (batch mode) -->
    <div v-if="showSource" class="preview-row__source" role="gridcell">
      <span class="preview-row__source-badge">{{ step.sourceFile }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Pencil } from 'lucide-vue-next'
import { computed, nextTick, ref } from 'vue'
import type { OperationType, SequenceStep } from '../../types/extraction'

const props = defineProps<{
  step: SequenceStep
  showSource?: boolean
}>()

const emit = defineEmits<{
  'toggle-include': [stepId: string]
  'update-label': [stepId: string, label: string]
}>()

// ── Inline editing ─────────────────────────────────────────────────────────

const editing = ref(false)
const editValue = ref('')
const editInputRef = ref<HTMLInputElement | null>(null)
const checkboxId = computed(() => `step-toggle-${props.step.id}`)

function startEdit() {
  if (!props.step.included) return
  editValue.value = props.step.label
  editing.value = true
  nextTick(() => {
    editInputRef.value?.focus()
    editInputRef.value?.select()
  })
}

function saveEdit() {
  if (!editing.value) return
  editing.value = false
  const trimmed = editValue.value.trim()
  if (trimmed && trimmed !== props.step.label) {
    console.debug('[SequencePreviewRow] label updated', {
      stepId: props.step.id,
      oldLabel: props.step.label,
      newLabel: trimmed
    })
    emit('update-label', props.step.id, trimmed)
  }
}

function cancelEdit() {
  editing.value = false
  editValue.value = ''
}

function toggleInclude() {
  console.debug('[SequencePreviewRow] toggle include', {
    stepId: props.step.id,
    included: !props.step.included
  })
  emit('toggle-include', props.step.id)
}

// ── Display helpers ─────────────────────────────────────────────────────────

const operationLabels: Record<OperationType, string> = {
  CREATE_CLASS: 'Class',
  CREATE_PROPERTY: 'Property',
  CREATE_INDIVIDUAL: 'Individual',
  UPDATE_LABEL: 'Update',
  UPDATE_COMMENT: 'Comment',
  DELETE: 'Delete'
}

const operationColors: Record<OperationType, string> = {
  CREATE_CLASS: 'class',
  CREATE_PROPERTY: 'property',
  CREATE_INDIVIDUAL: 'individual',
  UPDATE_LABEL: 'update',
  UPDATE_COMMENT: 'update',
  DELETE: 'delete'
}

const operationLabel = computed(() => operationLabels[props.step.operation] || props.step.operation)
const operationColor = computed(() => operationColors[props.step.operation] || 'default')
</script>

<style scoped>
.preview-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  border-bottom: 1px solid var(--border);
  font-size: var(--font-size-sm);
  transition: background var(--transition-fast);
}

.preview-row:hover {
  background: var(--surface-secondary);
}

.preview-row--excluded {
  opacity: 0.5;
  text-decoration: line-through;
}

.preview-row--duplicate {
  background: rgba(245, 158, 11, 0.04);
}

.preview-row__checkbox {
  width: 36px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-row__toggle {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.preview-row__toggle-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.preview-row__toggle-box {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.preview-row__toggle-input:checked + .preview-row__toggle-box {
  background-color: var(--primary);
  border-color: var(--primary);
}

.preview-row__toggle-input:checked + .preview-row__toggle-box::after {
  content: '✓';
  color: white;
  font-size: 12px;
}

.preview-row__operation {
  width: 80px;
  flex-shrink: 0;
}

.preview-row__op-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.preview-row__op-badge--class {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.preview-row__op-badge--property {
  background: rgba(139, 92, 246, 0.15);
  color: #8b5cf6;
}

.preview-row__op-badge--individual {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.preview-row__op-badge--update {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.preview-row__op-badge--delete {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.preview-row__entity-id {
  width: 180px;
  flex-shrink: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-row__id-text {
  font-family: 'IBM Plex Mono', monospace;
  font-size: var(--font-size-xs);
  color: var(--text-muted);
}

.preview-row__label {
  flex: 1;
  min-width: 0;
}

.preview-row__label-text {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  background: none;
  border: none;
  color: var(--text-primary);
  cursor: pointer;
  font-size: var(--font-size-sm);
  padding: 2px 4px;
  border-radius: var(--radius-sm);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-row__label-text:hover {
  background: var(--surface-tertiary);
}

.preview-row__edit-icon {
  opacity: 0;
  color: var(--text-muted);
  flex-shrink: 0;
}

.preview-row__label-text:hover .preview-row__edit-icon {
  opacity: 1;
}

.preview-row__edit {
  width: 100%;
}

.preview-row__edit-input {
  width: 100%;
  padding: var(--spacing-1) var(--spacing-2);
  font-size: var(--font-size-sm);
  border: 1px solid var(--primary);
  border-radius: var(--radius-sm);
  background: var(--surface-primary);
  color: var(--text-primary);
  outline: none;
}

.preview-row__context {
  width: 120px;
  flex-shrink: 0;
}

.preview-row__context-tag {
  display: inline-block;
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  background: var(--surface-tertiary);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-row__context-empty {
  color: var(--text-muted);
  font-size: var(--font-size-xs);
}

.preview-row__duplicate {
  width: 90px;
  flex-shrink: 0;
}

.badge--warning {
  display: inline-block;
  padding: 1px 6px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  white-space: nowrap;
}

.preview-row__source {
  width: 100px;
  flex-shrink: 0;
}

.preview-row__source-badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  background: var(--surface-tertiary);
  color: var(--text-muted);
  font-size: var(--font-size-xs);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
