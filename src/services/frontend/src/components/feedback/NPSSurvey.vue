<!-- @hlv:artifact code-frontend implements spec-feedback-001 -->
<!-- @ctx: NPS Survey — standalone survey with frequency limiting -->
<!-- @hlv:sec [INPUT_VALIDATION] — NPS score validated 0-10 range -->

<template>
  <div v-if="showSurvey" class="nps-survey" role="dialog" aria-modal="true" aria-labelledby="nps-title">
    <div class="nps-survey__card">
      <button class="nps-survey__dismiss" @click="onDismiss" aria-label="Dismiss survey">✕</button>

      <h3 id="nps-title" class="nps-survey__title">How likely are you to recommend VEDO?</h3>
      <p class="nps-survey__subtitle">Your feedback helps us improve.</p>

      <!-- NPS score selector -->
      <div class="nps-survey__scores">
        <button
          v-for="score in 11"
          :key="score - 1"
          type="button"
          class="nps-survey__score-btn"
          :class="scoreClass(score - 1)"
          @click="onSelect(score - 1)"
        >
          <span class="nps-survey__score-num">{{ score - 1 }}</span>
        </button>
      </div>

      <div class="nps-survey__labels">
        <span>Not likely</span>
        <span>Very likely</span>
      </div>

      <!-- Optional comment -->
      <div v-if="selectedScore !== null" class="nps-survey__comment">
        <label for="nps-comment" class="nps-survey__label">Tell us more (optional)</label>
        <textarea
          id="nps-comment"
          v-model="comment"
          class="nps-survey__textarea"
          rows="2"
          placeholder="What's the main reason for your score?"
        />
        <button class="btn btn--sm btn--primary" @click="onSubmit">Submit</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

const STORAGE_KEY_LAST_SHOWN = 'vedo_nps_last_shown'
const STORAGE_KEY_DISMISSED = 'vedo_nps_dismissed'
const NPS_INTERVAL_MS = 90 * 24 * 60 * 60 * 1000 // 90 days
const DISMISS_COOLDOWN_MS = 30 * 24 * 60 * 60 * 1000 // 30 days

const showSurvey = ref(false)
const selectedScore = ref<number | null>(null)
const comment = ref('')

const emit = defineEmits<{
  submit: [score: number, comment: string]
  dismiss: []
}>()

// @ctx: check if survey should be shown based on frequency limits
// @hlv atomicity — NPS survey frequency limit (90 days)
onMounted(() => {
  const lastShown = localStorage.getItem(STORAGE_KEY_LAST_SHOWN)
  const dismissed = localStorage.getItem(STORAGE_KEY_DISMISSED)

  if (lastShown) {
    const elapsed = Date.now() - Number.parseInt(lastShown, 10)
    if (elapsed < NPS_INTERVAL_MS) return
  }

  if (dismissed) {
    const elapsed = Date.now() - Number.parseInt(dismissed, 10)
    if (elapsed < DISMISS_COOLDOWN_MS) return
  }

  showSurvey.value = true
})

function scoreClass(score: number): string {
  if (score <= 6) return 'nps-survey__score-btn--detractor'
  if (score <= 8) return 'nps-survey__score-btn--passive'
  return 'nps-survey__score-btn--promoter'
}

function onSelect(score: number) {
  selectedScore.value = score
}

function onSubmit() {
  if (selectedScore.value === null) return

  // @hlv FEEDBACK-INVALID-NPS
  if (selectedScore.value < 0 || selectedScore.value > 10) return

  emit('submit', selectedScore.value, comment.value.trim())
  localStorage.setItem(STORAGE_KEY_LAST_SHOWN, Date.now().toString())
  showSurvey.value = false
}

function onDismiss() {
  emit('dismiss')
  localStorage.setItem(STORAGE_KEY_DISMISSED, Date.now().toString())
  showSurvey.value = false
}
</script>

<style scoped>
.nps-survey { position: fixed; bottom: var(--spacing-6); right: var(--spacing-6); z-index: var(--z-floating); }
.nps-survey__card {
  width: 400px; padding: var(--spacing-6); background: var(--surface-primary);
  border-radius: var(--radius-xl); box-shadow: var(--shadow-xl); position: relative;
}
.nps-survey__dismiss {
  position: absolute; top: var(--spacing-3); right: var(--spacing-3);
  background: none; border: none; font-size: var(--font-size-lg); cursor: pointer; color: var(--text-muted);
}
.nps-survey__title { font-size: var(--font-size-md); font-weight: var(--font-weight-semibold); margin: 0 0 var(--spacing-1); }
.nps-survey__subtitle { font-size: var(--font-size-sm); color: var(--text-muted); margin: 0 0 var(--spacing-4); }
.nps-survey__scores { display: flex; gap: 6px; justify-content: center; margin-bottom: var(--spacing-2); }
.nps-survey__score-btn {
  width: 36px; height: 36px; border-radius: var(--radius-md); border: 2px solid var(--border-default);
  background: var(--surface-primary); cursor: pointer; transition: transform 0.15s;
}
.nps-survey__score-btn:hover { transform: scale(1.1); }
.nps-survey__score-btn--detractor { border-color: #fca5a5; }
.nps-survey__score-btn--detractor:hover { background: #fee2e2; }
.nps-survey__score-btn--passive { border-color: #fcd34d; }
.nps-survey__score-btn--passive:hover { background: #fef3c7; }
.nps-survey__score-btn--promoter { border-color: #6ee7b7; }
.nps-survey__score-btn--promoter:hover { background: #d1fae5; }
.nps-survey__score-num { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); }
.nps-survey__labels { display: flex; justify-content: space-between; font-size: var(--font-size-xs); color: var(--text-muted); margin-bottom: var(--spacing-3); }
.nps-survey__comment { display: flex; flex-direction: column; gap: var(--spacing-2); padding-top: var(--spacing-3); border-top: 1px solid var(--border-default); }
.nps-survey__label { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); }
.nps-survey__textarea {
  padding: var(--spacing-2) var(--spacing-3); border: 1px solid var(--border-default);
  border-radius: var(--radius-md); font-size: var(--font-size-sm); font-family: inherit; resize: vertical;
}
.btn { padding: var(--spacing-2) var(--spacing-4); border-radius: var(--radius-md); font-size: var(--font-size-sm); cursor: pointer; border: none; }
.btn--primary { background: var(--color-primary); color: var(--font-on-primary); }
.btn--sm { padding: var(--spacing-1) var(--spacing-2); font-size: var(--font-size-xs); align-self: flex-end; }
</style>
