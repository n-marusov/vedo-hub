<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism ConflictResolver — modal for resolving name conflicts between batch files -->
<template>
  <Teleport to="body">
    <div v-if="visible" class="conflict-overlay" @click.self="$emit('close')" role="dialog" aria-modal="true" aria-label="Resolve conflicts">
      <div class="conflict-modal">
        <div class="conflict-modal__header">
          <h3 class="conflict-modal__title">Resolve Conflicts</h3>
          <p class="conflict-modal__desc">
            {{ conflicts.length }} {{ conflicts.length === 1 ? 'conflict needs' : 'conflicts need' }} your attention.
            The same entity has different labels across files.
          </p>
        </div>

        <div class="conflict-modal__list">
          <div
            v-for="(conflict, index) in conflicts"
            :key="`${conflict.stepA.entityId}-${index}`"
            class="conflict-item"
          >
            <div class="conflict-item__header">
              <AlertTriangle :size="16" class="conflict-item__icon" />
              <span class="conflict-item__entity">{{ conflict.stepA.entityId }}</span>
            </div>

            <div class="conflict-item__values">
              <div class="conflict-item__value">
                <span class="conflict-item__file">{{ conflict.stepA.sourceFile || 'File A' }}</span>
                <span class="conflict-item__label">{{ conflict.valueA }}</span>
                <button
                  class="conflict-item__choose"
                  :class="{ 'conflict-item__choose--selected': conflict.resolution === 'keep_a' }"
                  type="button"
                  @click="selectResolution(index, 'keep_a')"
                >
                  {{ conflict.resolution === 'keep_a' ? '✓ Selected' : 'Keep A' }}
                </button>
              </div>
              <div class="conflict-item__value">
                <span class="conflict-item__file">{{ conflict.stepB.sourceFile || 'File B' }}</span>
                <span class="conflict-item__label">{{ conflict.valueB }}</span>
                <button
                  class="conflict-item__choose"
                  :class="{ 'conflict-item__choose--selected': conflict.resolution === 'keep_b' }"
                  type="button"
                  @click="selectResolution(index, 'keep_b')"
                >
                  {{ conflict.resolution === 'keep_b' ? '✓ Selected' : 'Keep B' }}
                </button>
              </div>
            </div>

            <div class="conflict-item__custom">
              <label class="conflict-item__custom-label">
                <input
                  type="radio"
                  :checked="conflict.resolution === 'custom'"
                  @change="selectResolution(index, 'custom', conflict.customValue)"
                />
                Custom label
              </label>
              <input
                v-if="conflict.resolution === 'custom'"
                :value="conflict.customValue"
                class="conflict-item__custom-input"
                placeholder="Enter custom label..."
                @input="onCustomInput(index, ($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>
        </div>

        <div class="conflict-modal__footer">
          <button class="conflict-btn conflict-btn--secondary" type="button" @click="$emit('close')">
            Cancel
          </button>
          <button
            class="conflict-btn conflict-btn--primary"
            type="button"
            :disabled="!allResolved"
            @click="$emit('apply')"
          >
            Apply Resolutions
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { AlertTriangle } from "@lucide/vue";
import { computed } from "vue";
import type { ConflictResolution } from "../../types/extraction";

const props = defineProps<{
	visible: boolean;
	conflicts: ConflictResolution[];
}>();

const emit = defineEmits<{
	resolve: [
		index: number,
		resolution: ConflictResolution["resolution"],
		customValue?: string,
	];
	close: [];
	apply: [];
}>();

const allResolved = computed(() => props.conflicts.length > 0);

function selectResolution(
	index: number,
	resolution: ConflictResolution["resolution"],
	customValue?: string,
) {
	emit("resolve", index, resolution, customValue);
}

function onCustomInput(index: number, value: string) {
	emit("resolve", index, "custom", value);
}
</script>

<style scoped>
.conflict-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal, 1000);
}

.conflict-modal {
  width: 560px;
  max-width: 90vw;
  max-height: 85vh;
  background: var(--surface-primary, #101010);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl, 12px);
  display: flex;
  flex-direction: column;
}

.conflict-modal__header {
  padding: var(--spacing-4, 16px) var(--spacing-6, 24px);
  border-bottom: 1px solid var(--border);
}

.conflict-modal__title {
  font-size: var(--font-size-lg, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #fafafa);
  margin: 0 0 var(--spacing-1);
}

.conflict-modal__desc {
  font-size: var(--font-size-sm, 13px);
  color: var(--text-muted, #64748b);
  margin: 0;
}

.conflict-modal__list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4, 16px) var(--spacing-6, 24px);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4, 16px);
}

.conflict-item {
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: var(--spacing-3, 12px);
  background: rgba(245, 158, 11, 0.04);
}

.conflict-item__header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2, 8px);
  margin-bottom: var(--spacing-2, 8px);
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  color: var(--warning, #f59e0b);
}

.conflict-item__icon {
  flex-shrink: 0;
}

.conflict-item__entity {
  font-family: 'IBM Plex Mono', monospace;
  font-size: var(--font-size-xs, 12px);
}

.conflict-item__values {
  display: flex;
  gap: var(--spacing-2, 8px);
}

.conflict-item__value {
  flex: 1;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  padding: var(--spacing-2, 8px);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1, 4px);
}

.conflict-item__file {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-muted, #64748b);
}

.conflict-item__label {
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  color: var(--text-primary, #fafafa);
}

.conflict-item__choose {
  margin-top: var(--spacing-1, 4px);
  padding: var(--spacing-1, 4px) var(--spacing-2, 8px);
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-secondary, #6b7280);
  font-size: var(--font-size-xs, 12px);
  cursor: pointer;
  align-self: flex-start;
}

.conflict-item__choose:hover {
  border-color: var(--primary, #10b981);
  color: var(--primary, #10b981);
}

.conflict-item__choose--selected {
  border-color: var(--primary, #10b981);
  background: rgba(16, 185, 129, 0.1);
  color: var(--primary, #10b981);
}

.conflict-item__custom {
  margin-top: var(--spacing-2, 8px);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1, 4px);
}

.conflict-item__custom-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-1, 4px);
  font-size: var(--font-size-sm, 13px);
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
}

.conflict-item__custom-input {
  padding: var(--spacing-1, 4px) var(--spacing-2, 8px);
  font-size: var(--font-size-sm, 13px);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-primary, #101010);
  color: var(--text-primary, #fafafa);
}

.conflict-item__custom-input:focus {
  outline: none;
  border-color: var(--primary, #10b981);
}

.conflict-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-2, 8px);
  padding: var(--spacing-4, 16px) var(--spacing-6, 24px);
  border-top: 1px solid var(--border);
}

.conflict-btn {
  padding: var(--spacing-2, 8px) var(--spacing-4, 16px);
  border-radius: var(--radius-md, 8px);
  font-size: var(--font-size-sm, 13px);
  font-weight: var(--font-weight-medium, 500);
  cursor: pointer;
  transition: all var(--transition-fast, 0.15s);
  border: none;
}

.conflict-btn--primary {
  background: var(--primary, #10b981);
  color: var(--text-inverse, #0a0a0a);
}

.conflict-btn--primary:hover:not(:disabled) {
  background: var(--primary-hover, #34d399);
}

.conflict-btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.conflict-btn--secondary {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text-secondary, #6b7280);
}

.conflict-btn--secondary:hover {
  background: var(--surface-secondary, #141414);
}
</style>
