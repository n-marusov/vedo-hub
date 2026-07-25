<!-- @hlv:artifact code-frontend implements spec-gui-login-001 -->
<!-- @ctx: UI-Kit OAuthButton component — renders OAuth provider button with icon and label -->
<template>
  <button
    :class="['oauth-btn', `oauth-btn--${provider}`]"
    :disabled="disabled"
    :aria-disabled="disabled"
    :aria-label="`Sign in with ${label}`"
    @click="$emit('click', $event)"
  >
    <span v-if="icon" class="oauth-btn__icon" aria-hidden="true">{{ icon }}</span>
    <span class="oauth-btn__label">{{ label }}</span>
  </button>
</template>

<script setup lang="ts">
defineProps<{
  provider: 'vk' | 'yandex' | 'mailru' | 'google' | 'corporate_sso'
  label: string
  icon?: string
  disabled?: boolean
}>()

defineEmits<{
  click: [event: MouseEvent]
}>()
</script>

<style scoped>
.oauth-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  width: 100%;
  padding: var(--spacing-3) var(--spacing-4);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-default);
  background-color: var(--surface-primary);
  color: var(--text-primary);
  transition: all var(--transition-fast);
  cursor: pointer;
}

.oauth-btn:hover:not(:disabled) {
  background-color: var(--surface-secondary);
  border-color: var(--primary);
}

.oauth-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.oauth-btn__icon {
  font-size: var(--font-size-lg);
}

.oauth-btn--google { border-left: 3px solid #4285f4; }
.oauth-btn--vk { border-left: 3px solid #0077ff; }
.oauth-btn--yandex { border-left: 3px solid #fc3f1d; }
.oauth-btn--mailru { border-left: 3px solid #005ff9; }
.oauth-btn--corporate_sso { border-left: 3px solid var(--secondary); }
</style>
