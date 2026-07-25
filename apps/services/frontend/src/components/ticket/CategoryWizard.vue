<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Category wizard — conditional field display based on selected category -->
<!-- @hlv:sec [INPUT_VALIDATION] — category selection drives conditional validation rules -->

<template>
  <div class="category-wizard" role="group" aria-labelledby="category-wizard-title">
    <h3 id="category-wizard-title" class="category-wizard__title">Category</h3>

    <div class="category-wizard__selector">
      <label for="category-select" class="sr-only">Select ticket category</label>
      <select
        id="category-select"
        :value="modelValue.category"
        @change="onCategoryChange"
        class="category-wizard__select"
        :aria-invalid="showCategoryError"
        aria-describedby="category-error"
      >
        <option value="" disabled>Select a category</option>
        <option v-for="cat in categories" :key="cat" :value="cat">{{ catLabels[cat] }}</option>
      </select>
      <span v-if="showCategoryError" id="category-error" class="category-wizard__error" role="alert">
        Category is required
      </span>
    </div>

    <div class="category-wizard__severity">
      <label for="severity-select" class="category-wizard__label">Severity</label>
      <select
        id="severity-select"
        :value="modelValue.user_severity"
        @change="onSeverityChange"
        class="category-wizard__select"
        :aria-invalid="showSeverityError"
        aria-describedby="severity-error"
      >
        <option value="" disabled>Select severity</option>
        <option v-for="sev in severities" :key="sev" :value="sev">{{ sevLabels[sev] }}</option>
      </select>
      <span v-if="showSeverityError" id="severity-error" class="category-wizard__error" role="alert">
        Severity is required
      </span>
    </div>

    <!-- @ctx: conditional fields — bug requires steps_to_reproduce -->
    <!-- @hlv TICKET-UI-MISSING-STEPS -->
    <div v-if="modelValue.category === 'bug'" class="category-wizard__conditional">
      <label for="steps-reproduce" class="category-wizard__label">
        Steps to reproduce <span class="required" aria-hidden="true">*</span>
      </label>
      <textarea
        id="steps-reproduce"
        :value="modelValue.steps_to_reproduce"
        @input="onStepsInput"
        class="category-wizard__textarea"
        rows="4"
        placeholder="1. Open...&#10;2. Click...&#10;3. Observe..."
        :aria-invalid="showStepsError"
        aria-describedby="steps-error"
      />
      <span v-if="showStepsError" id="steps-error" class="category-wizard__error" role="alert">
        Steps to reproduce are required for bug reports
      </span>
    </div>

    <!-- @ctx: conditional fields — feature requires expected_behavior -->
    <!-- @hlv TICKET-UI-MISSING-EXPECTED -->
    <div v-if="modelValue.category === 'feature'" class="category-wizard__conditional">
      <label for="expected-behavior" class="category-wizard__label">
        Expected behavior <span class="required" aria-hidden="true">*</span>
      </label>
      <textarea
        id="expected-behavior"
        :value="modelValue.expected_behavior"
        @input="onExpectedInput"
        class="category-wizard__textarea"
        rows="4"
        placeholder="Describe what you expected to happen..."
        :aria-invalid="showExpectedError"
        aria-describedby="expected-error"
      />
      <span v-if="showExpectedError" id="expected-error" class="category-wizard__error" role="alert">
        Expected behavior is required for feature requests
      </span>
    </div>

    <!-- @ctx: conditional fields — performance shows duration fields -->
    <div v-if="modelValue.category === 'performance'" class="category-wizard__conditional">
      <label for="expected-duration" class="category-wizard__label">Expected duration</label>
      <input
        id="expected-duration"
        type="text"
        :value="modelValue.expected_duration || ''"
        @input="onExpectedDurationInput"
        class="category-wizard__input"
        placeholder="e.g. 2s, 500ms"
      />
      <label for="actual-duration" class="category-wizard__label">Actual duration</label>
      <input
        id="actual-duration"
        type="text"
        :value="modelValue.actual_duration || ''"
        @input="onActualDurationInput"
        class="category-wizard__input"
        placeholder="e.g. 15s, 3000ms"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: {
    category: string
    user_severity: string
    steps_to_reproduce: string
    expected_behavior: string
    expected_duration?: string
    actual_duration?: string
  }
  errors: {
    category?: boolean
    severity?: boolean
    steps?: boolean
    expected?: boolean
  }
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string>]
}>()

const categories = ['bug', 'performance', 'question', 'feature', 'documentation', 'access'] as const
const severities = ['critical', 'high', 'medium', 'low'] as const

const catLabels: Record<string, string> = {
  bug: 'Bug',
  performance: 'Performance',
  question: 'Question',
  feature: 'Feature Request',
  documentation: 'Documentation',
  access: 'Access Issue'
}

const sevLabels: Record<string, string> = {
  critical: 'Critical',
  high: 'High',
  medium: 'Medium',
  low: 'Low'
}

const showCategoryError = computed(() => !!props.errors.category)
const showSeverityError = computed(() => !!props.errors.severity)
const showStepsError = computed(() => !!props.errors.steps)
const showExpectedError = computed(() => !!props.errors.expected)

function onCategoryChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', { category: target.value })
}

function onSeverityChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', { user_severity: target.value })
}

function onStepsInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', { steps_to_reproduce: target.value })
}

function onExpectedInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', { expected_behavior: target.value })
}

function onExpectedDurationInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', { expected_duration: target.value })
}

function onActualDurationInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', { actual_duration: target.value })
}
</script>

<style scoped>
.category-wizard { display: flex; flex-direction: column; gap: var(--spacing-4); }
.category-wizard__title { font-size: var(--font-size-md); font-weight: var(--font-weight-semibold); margin: 0; }
.category-wizard__selector, .category-wizard__severity, .category-wizard__conditional { display: flex; flex-direction: column; gap: var(--spacing-1); }
.category-wizard__label { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); }
.category-wizard__select, .category-wizard__input, .category-wizard__textarea {
  padding: var(--spacing-2) var(--spacing-3); border: 1px solid var(--border-default);
  border-radius: var(--radius-md); font-size: var(--font-size-sm); font-family: inherit;
  background: var(--surface-primary); color: var(--font-primary);
}
.category-wizard__select[aria-invalid="true"], .category-wizard__textarea[aria-invalid="true"] {
  border-color: var(--color-error);
}
.category-wizard__error { font-size: var(--font-size-xs); color: var(--color-error); }
.category-wizard__conditional { padding: var(--spacing-3); background: var(--surface-secondary); border-radius: var(--radius-md); }
.required { color: var(--color-error); }
</style>
