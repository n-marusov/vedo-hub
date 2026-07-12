<!-- @hlv:artifact code-frontend implements spec-gui-dash-001 -->
<!-- @ctx: UI-Kit Avatar component — shows user avatar image or initials fallback -->
<template>
  <div
    :class="['avatar', `avatar--${size}`]"
    :style="avatarStyle"
    role="img"
    :aria-label="`Avatar for ${displayName}`"
  >
    <img
      v-if="src && !hasError"
      :src="src"
      :alt="`Avatar for ${displayName}`"
      @error="hasError = true"
    />
    <span v-else class="avatar__initials" :aria-hidden="!!src">
      {{ initials }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{
  src?: string
  displayName: string
  size?: 'sm' | 'md' | 'lg'
}>()

const hasError = ref(false)

const initials = computed(() => {
  return props.displayName
    .split(' ')
    .map((n) => n[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
})

const avatarStyle = computed(() => {
  if (!props.src || hasError.value) {
    return { backgroundColor: stringToColor(props.displayName) }
  }
  return {}
})

function stringToColor(str: string): string {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
  }
  const h = Math.abs(hash) % 360
  return `hsl(${h}, 40%, 60%)`
}
</script>

<style scoped>
.avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-full);
  overflow: hidden;
  flex-shrink: 0;
}

.avatar--sm { width: 32px; height: 32px; }
.avatar--md { width: 40px; height: 40px; }
.avatar--lg { width: 64px; height: 64px; }

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar__initials {
  color: white;
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
}

.avatar--lg .avatar__initials {
  font-size: var(--font-size-xl);
}
</style>
