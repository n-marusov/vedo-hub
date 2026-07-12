<!-- CommentsThread.vue -->
<template>
  <div class="comments-thread" role="region" :aria-label="'Comments'">
    <ul class="comments-thread__list">
      <li v-for="comment in comments" :key="comment.id" class="comments-thread__item">
        <Avatar :display-name="comment.author" size="sm" />
        <div class="comments-thread__content">
          <div class="comments-thread__meta">
            <strong>{{ comment.author }}</strong>
            <time :datetime="comment.created_at">{{ formatDate(comment.created_at) }}</time>
          </div>
          <p class="comments-thread__text">{{ comment.text }}</p>
        </div>
      </li>
    </ul>
    <div class="comments-thread__input">
      <TextInput v-model="newComment" label="Add comment" placeholder="Write a comment..." />
      <PrimaryButton :disabled="!newComment" @click="$emit('add', newComment)">Send</PrimaryButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Avatar from '../ui-kit/Avatar.vue'
import PrimaryButton from '../ui-kit/PrimaryButton.vue'
import TextInput from '../ui-kit/TextInput.vue'

defineProps<{ comments: Array<{ id: string; author: string; text: string; created_at: string }> }>()
defineEmits<{ add: [text: string] }>()

const newComment = ref('')

function formatDate(date: string): string {
  return new Date(date).toLocaleString()
}
</script>

<style scoped>
.comments-thread { padding: var(--spacing-4); }
.comments-thread__list { display: flex; flex-direction: column; gap: var(--spacing-4); margin-bottom: var(--spacing-4); }
.comments-thread__item { display: flex; gap: var(--spacing-3); }
.comments-thread__content { flex: 1; }
.comments-thread__meta { display: flex; gap: var(--spacing-2); font-size: var(--font-size-xs); color: var(--text-secondary); margin-bottom: var(--spacing-1); }
.comments-thread__text { font-size: var(--font-size-sm); }
.comments-thread__input { display: flex; gap: var(--spacing-2); }
</style>
