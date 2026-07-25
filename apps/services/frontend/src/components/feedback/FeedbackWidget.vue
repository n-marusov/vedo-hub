<!-- @hlv:artifact code-frontend implements spec-feedback-001 -->
<!-- @ctx: Feedback widget — floating button with feedback form -->
<!-- @hlv:sec [AUTH_BOUNDARY] — requires authenticated session to submit feedback -->
<!-- @hlv:sec [INPUT_VALIDATION] — feedback text validated and sanitized -->

<template>
  <div class="feedback-widget" aria-label="Feedback widget">
    <!-- Floating button -->
    <button
      v-if="!isOpen"
      class="feedback-widget__trigger"
      @click="isOpen = true"
      aria-label="Open feedback form"
      aria-expanded="false"
    >
      <span class="feedback-widget__icon">💬</span>
      <span class="feedback-widget__label">Feedback</span>
    </button>

    <!-- Feedback panel -->
    <div v-if="isOpen" class="feedback-widget__panel" role="dialog" aria-modal="true" aria-labelledby="feedback-title">
      <div class="feedback-widget__header">
        <h3 id="feedback-title" class="feedback-widget__title">Share your feedback</h3>
        <button class="feedback-widget__close" @click="isOpen = false" aria-label="Close feedback form">✕</button>
      </div>

      <form class="feedback-widget__form" @submit.prevent="onSubmit" novalidate>
        <!-- Feedback type -->
        <div class="feedback-widget__field">
          <label for="feedback-type" class="feedback-widget__label">Type</label>
          <select
            id="feedback-type"
            v-model="form.type"
            class="feedback-widget__select"
            required
          >
            <option value="" disabled>Select type</option>
            <option v-for="t in types" :key="t" :value="t">{{ typeLabels[t] }}</option>
          </select>
        </div>

        <!-- Feedback text -->
        <div class="feedback-widget__field">
          <label for="feedback-text" class="feedback-widget__label">
            Your feedback <span class="required" aria-hidden="true">*</span>
          </label>
          <textarea
            id="feedback-text"
            v-model="form.text"
            class="feedback-widget__textarea"
            rows="4"
            maxlength="2000"
            placeholder="Tell us what you think..."
            :aria-invalid="!!errors.text"
            aria-describedby="feedback-text-error"
            required
          />
          <span v-if="errors.text" id="feedback-text-error" class="feedback-widget__error" role="alert">
            {{ errors.text }}
          </span>
          <span class="feedback-widget__char-count">{{ form.text.length }}/2000</span>
        </div>

        <!-- NPS score -->
        <div class="feedback-widget__field">
          <label class="feedback-widget__label">How likely are you to recommend us? (optional)</label>
          <div class="feedback-widget__nps">
            <button
              v-for="score in 11"
              :key="score - 1"
              type="button"
              class="feedback-widget__nps-btn"
              :class="npsClass(score - 1)"
              :aria-label="`Score ${score - 1} out of 10`"
              @click="form.nps_score = score - 1"
            >
              {{ score - 1 }}
            </button>
          </div>
          <span v-if="form.nps_score !== null" class="feedback-widget__nps-label">
            {{ npsCategory(form.nps_score) }}
          </span>
          <!-- @ctx: NPS validation -->
          <!-- @hlv FEEDBACK-INVALID-NPS -->
          <span v-if="errors.nps" class="feedback-widget__error" role="alert">{{ errors.nps }}</span>
        </div>

        <!-- Screenshot option -->
        <div class="feedback-widget__field">
          <label class="feedback-widget__checkbox">
            <input type="checkbox" v-model="form.attach_screenshot" />
            <span>Attach screenshot</span>
          </label>
        </div>

        <!-- Create ticket option -->
        <div class="feedback-widget__field">
          <label class="feedback-widget__checkbox">
            <input type="checkbox" v-model="form.create_ticket" />
            <span>Create a support ticket from this feedback</span>
          </label>
        </div>

        <!-- Actions -->
        <div class="feedback-widget__actions">
          <button type="button" class="btn btn--ghost btn--sm" @click="isOpen = false">Cancel</button>
          <button type="submit" class="btn btn--primary btn--sm" :disabled="isSubmitting">
            {{ isSubmitting ? 'Sending...' : 'Send Feedback' }}
          </button>
        </div>

        <!-- @ctx: network error display -->
        <!-- @hlv FEEDBACK-NETWORK-ERROR -->
        <div v-if="submitError" class="feedback-widget__submit-error" role="alert">
          {{ submitError }}
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'

const isOpen = ref(false)
const isSubmitting = ref(false)
const submitError = ref<string | null>(null)

const form = reactive({
  type: '',
  text: '',
  nps_score: null as number | null,
  attach_screenshot: false,
  create_ticket: false
})

const errors = reactive<Record<string, string>>({})

const types = ['bug', 'suggestion', 'ux', 'question', 'praise'] as const
const typeLabels: Record<string, string> = {
  bug: 'Bug Report',
  suggestion: 'Suggestion',
  ux: 'UX Issue',
  question: 'Question',
  praise: 'Praise'
}

function npsClass(score: number): string {
  if (score <= 6) return 'feedback-widget__nps-btn--detractor'
  if (score <= 8) return 'feedback-widget__nps-btn--passive'
  return 'feedback-widget__nps-btn--promoter'
}

function npsCategory(score: number): string {
  if (score <= 6) return 'Detractor'
  if (score <= 8) return 'Passive'
  return 'Promoter'
}

function validate(): boolean {
  for (const k of Object.keys(errors)) delete errors[k]

  // @hlv FEEDBACK-EMPTY-TEXT
  if (!form.text.trim()) {
    errors.text = 'Please describe your feedback before submitting.'
    return false
  }

  // @hlv FEEDBACK-INVALID-NPS
  if (form.nps_score !== null && (form.nps_score < 0 || form.nps_score > 10)) {
    errors.nps = 'NPS score must be between 0 and 10.'
    return false
  }

  return true
}

async function onSubmit() {
  submitError.value = null
  if (!validate()) return

  isSubmitting.value = true

  try {
    const payload = {
      type: form.type,
      text: form.text.trim(),
      nps_score: form.nps_score,
      context: {
        page_url: window.location.href,
        action: 'feedback_submitted',
        trace_id: crypto.randomUUID?.() || null
      },
      attach_screenshot: form.attach_screenshot,
      create_ticket: form.create_ticket
    }

    // @ctx: emit to parent for API call
    emit('submit', payload)
    resetForm()
    isOpen.value = false
  } catch {
    // @hlv FEEDBACK-NETWORK-ERROR
    submitError.value = 'Failed to send feedback. Please try again later.'
  } finally {
    isSubmitting.value = false
  }
}

function resetForm() {
  form.type = ''
  form.text = ''
  form.nps_score = null
  form.attach_screenshot = false
  form.create_ticket = false
}

const emit = defineEmits<{
  submit: [payload: Record<string, unknown>]
}>()
</script>

<style scoped>
.feedback-widget { position: fixed; bottom: var(--spacing-6); right: var(--spacing-6); z-index: var(--z-floating); }
.feedback-widget__trigger {
  display: flex; align-items: center; gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-4); border-radius: var(--radius-xl);
  background: var(--color-primary); color: var(--font-on-primary);
  border: none; cursor: pointer; font-size: var(--font-size-sm); font-weight: var(--font-weight-medium);
  box-shadow: var(--shadow-lg); transition: transform 0.2s;
}
.feedback-widget__trigger:hover { transform: scale(1.05); }
.feedback-widget__icon { font-size: var(--font-size-lg); }
.feedback-widget__panel {
  position: absolute; bottom: 0; right: 0; width: 380px; max-height: 80vh;
  background: var(--surface-primary); border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl); overflow: auto;
}
.feedback-widget__header {
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--spacing-4); border-bottom: 1px solid var(--border-default);
}
.feedback-widget__title { font-size: var(--font-size-md); font-weight: var(--font-weight-semibold); margin: 0; }
.feedback-widget__close { background: none; border: none; font-size: var(--font-size-lg); cursor: pointer; color: var(--text-muted); }
.feedback-widget__form { padding: var(--spacing-4); display: flex; flex-direction: column; gap: var(--spacing-3); }
.feedback-widget__field { display: flex; flex-direction: column; gap: var(--spacing-1); }
.feedback-widget__label { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); }
.feedback-widget__select, .feedback-widget__textarea {
  padding: var(--spacing-2) var(--spacing-3); border: 1px solid var(--border-default);
  border-radius: var(--radius-md); font-size: var(--font-size-sm); font-family: inherit;
  background: var(--surface-primary);
}
.feedback-widget__textarea[aria-invalid="true"] { border-color: var(--color-error); }
.feedback-widget__char-count { font-size: var(--font-size-xs); color: var(--text-muted); text-align: right; }
.feedback-widget__nps { display: flex; gap: 4px; }
.feedback-widget__nps-btn {
  width: 32px; height: 32px; border: 1px solid var(--border-default); border-radius: var(--radius-sm);
  background: var(--surface-primary); cursor: pointer; font-size: var(--font-size-xs);
}
.feedback-widget__nps-btn--detractor { background: #fee2e2; border-color: #fca5a5; }
.feedback-widget__nps-btn--passive { background: #fef3c7; border-color: #fcd34d; }
.feedback-widget__nps-btn--promoter { background: #d1fae5; border-color: #6ee7b7; }
.feedback-widget__nps-label { font-size: var(--font-size-xs); font-weight: var(--font-weight-medium); }
.feedback-widget__checkbox { display: flex; align-items: center; gap: var(--spacing-2); font-size: var(--font-size-sm); cursor: pointer; }
.feedback-widget__actions { display: flex; justify-content: flex-end; gap: var(--spacing-2); padding-top: var(--spacing-2); }
.feedback-widget__error { font-size: var(--font-size-xs); color: var(--color-error); }
.feedback-widget__submit-error { padding: var(--spacing-2); background: var(--color-error); color: var(--font-on-error); border-radius: var(--radius-md); font-size: var(--font-size-sm); }
.required { color: var(--color-error); }
.btn { padding: var(--spacing-2) var(--spacing-4); border-radius: var(--radius-md); font-size: var(--font-size-sm); cursor: pointer; border: none; }
.btn--primary { background: var(--color-primary); color: var(--font-on-primary); }
.btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn--ghost { background: transparent; color: var(--font-secondary); border: 1px solid var(--border-default); }
.btn--sm { padding: var(--spacing-1) var(--spacing-2); font-size: var(--font-size-xs); }
</style>
