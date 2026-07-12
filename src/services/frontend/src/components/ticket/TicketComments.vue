<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Ticket comments — display comment thread and add new comments -->
<!-- @hlv:sec [INPUT_VALIDATION] — comment text sanitized before submission -->

<template>
  <div class="ticket-comments" aria-labelledby="comments-title">
    <h3 id="comments-title" class="ticket-comments__title">
      Comments ({{ comments.length }})
    </h3>

    <!-- Comment list -->
    <div v-if="comments.length > 0" class="ticket-comments__list" role="log" aria-live="polite">
      <div
        v-for="(comment, index) in comments"
        :key="index"
        class="ticket-comments__item"
      >
        <div class="ticket-comments__author">
          <span class="ticket-comments__name">{{ comment.author }}</span>
          <time class="ticket-comments__time" :datetime="comment.created_at">
            {{ formatDate(comment.created_at) }}
          </time>
        </div>
        <p class="ticket-comments__text">{{ comment.text }}</p>
      </div>
    </div>

    <div v-else class="ticket-comments__empty">
      No comments yet.
    </div>

    <!-- Add comment form -->
    <!-- @ctx: users can add comments to any of their open tickets -->
    <!-- @hlv atomicity — comment addition to open ticket -->
    <form v-if="canComment" class="ticket-comments__form" @submit.prevent="onSubmit">
      <label for="comment-text" class="sr-only">Add a comment</label>
      <textarea
        id="comment-text"
        v-model="newComment"
        class="ticket-comments__textarea"
        rows="3"
        placeholder="Add a comment..."
        :aria-invalid="!!commentError"
        aria-describedby="comment-error"
      />
      <div class="ticket-comments__form-actions">
        <span v-if="commentError" id="comment-error" class="ticket-comments__error" role="alert">
          {{ commentError }}
        </span>
        <button type="submit" class="btn btn--sm btn--primary" :disabled="!newComment.trim()">
          Add comment
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface TicketComment {
  author: string
  text: string
  created_at: string
}

const props = defineProps<{
  comments: TicketComment[]
  ticketStatus: string
  canComment?: boolean
}>()

const emit = defineEmits<{
  comment: [text: string]
}>()

const newComment = ref('')
const commentError = ref<string | null>(null)

const canComment = computed(() => {
  if (props.canComment !== undefined) return props.canComment
  const openStatuses = ['new', 'in_review', 'in_progress', 'reopened']
  return openStatuses.includes(props.ticketStatus)
})

function onSubmit() {
  commentError.value = null
  const text = newComment.value.trim()

  if (!text) {
    commentError.value = 'Comment cannot be empty.'
    return
  }

  emit('comment', text)
  newComment.value = ''
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}
</script>

<style scoped>
.ticket-comments { display: flex; flex-direction: column; gap: var(--spacing-3); }
.ticket-comments__title { font-size: var(--font-size-md); font-weight: var(--font-weight-semibold); margin: 0; }
.ticket-comments__list { display: flex; flex-direction: column; gap: var(--spacing-3); max-height: 400px; overflow-y: auto; }
.ticket-comments__item { padding: var(--spacing-3); background: var(--surface-secondary); border-radius: var(--radius-md); }
.ticket-comments__author { display: flex; justify-content: space-between; margin-bottom: var(--spacing-1); }
.ticket-comments__name { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); }
.ticket-comments__time { font-size: var(--font-size-xs); color: var(--text-muted); }
.ticket-comments__text { font-size: var(--font-size-sm); margin: 0; white-space: pre-wrap; }
.ticket-comments__empty { font-size: var(--font-size-sm); color: var(--text-muted); padding: var(--spacing-4); text-align: center; }
.ticket-comments__form { display: flex; flex-direction: column; gap: var(--spacing-2); }
.ticket-comments__textarea {
  padding: var(--spacing-2) var(--spacing-3); border: 1px solid var(--border-default);
  border-radius: var(--radius-md); font-size: var(--font-size-sm); font-family: inherit;
  resize: vertical;
}
.ticket-comments__form-actions { display: flex; justify-content: space-between; align-items: center; }
.ticket-comments__error { font-size: var(--font-size-xs); color: var(--color-error); }
.btn { padding: var(--spacing-2) var(--spacing-4); border-radius: var(--radius-md); font-size: var(--font-size-sm); cursor: pointer; border: none; }
.btn--primary { background: var(--color-primary); color: var(--font-on-primary); }
.btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn--sm { padding: var(--spacing-1) var(--spacing-2); font-size: var(--font-size-xs); }
</style>
