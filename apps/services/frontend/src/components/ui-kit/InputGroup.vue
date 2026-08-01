<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: UI-Kit InputGroup component — text input with an immutable prefix
     (PrimeVue InputGroup style). Used for group/project URL fields where the
     domain is a fixed prefix and only the slug is editable. -->
<template>
  <div :class="['input-group', { 'input-group--error': error }]">
    <label v-if="label" :for="inputId" class="input-group__label">{{ label }}</label>
    <div class="input-group__wrap">
      <span v-if="prefix" class="input-group__prefix" aria-hidden="true">{{ prefix }}</span>
      <input
        :id="inputId"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :autocomplete="autocomplete"
        :aria-label="ariaLabel"
        :aria-invalid="!!error"
        :aria-describedby="error ? errorId : undefined"
        class="input-group__input"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <span v-if="error" :id="errorId" class="input-group__error" role="alert">
      {{ error }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
	modelValue?: string;
	label?: string;
	/** Immutable prefix rendered inside the field (e.g. "vedo-core.local/"). */
	prefix?: string;
	type?: string;
	placeholder?: string;
	disabled?: boolean;
	autocomplete?: string;
	/** Explicit id for the input element (used for label for= association). */
	id?: string;
	ariaLabel?: string;
	error?: string;
}>();

defineEmits<{
	"update:modelValue": [value: string];
}>();

const inputId = computed(
	() =>
		props.id ??
		(props.label
			? `input-${props.label.toLowerCase().replace(/\s+/g, "-")}`
			: "input-group"),
);
const errorId = computed(() => `${inputId.value}-error`);
</script>

<style scoped>
.input-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.input-group__label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
}

.input-group__wrap {
  display: flex;
  align-items: center;
  height: 36px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background-color: var(--surface-primary);
  overflow: hidden;
  transition: border-color var(--transition-fast);
}

.input-group__wrap:focus-within {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 2px var(--primary-muted);
}

.input-group__prefix {
  padding: 0 var(--spacing-2);
  font-size: var(--font-size-sm);
  font-family: 'IBM Plex Mono', monospace;
  color: var(--text-muted);
  white-space: nowrap;
  user-select: none;
}

.input-group__input {
  flex: 1;
  min-width: 0;
  padding: var(--spacing-2) var(--spacing-3);
  border: none;
  outline: none;
  background: transparent;
  font-size: var(--font-size-sm);
  font-family: 'IBM Plex Mono', monospace;
  color: var(--text-primary);
}

.input-group__input::placeholder {
  color: var(--text-muted);
}

.input-group__input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.input-group--error .input-group__wrap {
  border-color: var(--border-error);
}

.input-group__error {
  font-size: var(--font-size-xs);
  color: var(--status-error);
}
</style>
