<template>
  <Teleport to="body">
    <Transition
      enter-active-class="toast-enter-active"
      leave-active-class="toast-leave-active"
      enter-from-class="toast-enter-from"
      leave-to-class="toast-leave-to"
    >
      <div
        v-if="toastVisible"
        :role="toastType === 'success' ? 'status' : 'alert'"
        aria-live="polite"
        class="toast"
        :class="`toast--${toastType}`"
      >
        <div class="toast-body">
          <component
            :is="toastType === 'success' ? CircleCheck : AlertCircle"
            :size="14"
            class="toast-icon"
          />
          <span class="toast-message">{{ toastMessage }}</span>
        </div>
        <button
          class="toast-close"
          type="button"
          :aria-label="t('toast.close')"
          @click="dismissToast"
        >
          <X :size="12" />
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/useI18n";
import { useToast } from "@/composables/useToast";
import { AlertCircle, CircleCheck, X } from "lucide-vue-next";

const { t } = useI18n();
const { toastMessage, toastType, toastVisible, dismissToast } = useToast();
</script>

<style scoped>
.toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: var(--z-toast, 500);

  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;

  min-width: 300px;
  max-width: 480px;
  padding: 8px 12px;

  border-radius: 6px;
  border: 1px solid #2a2a2a;
  background: #141414;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

.toast--success .toast-icon {
  color: #10b981;
}

.toast--success .toast-message {
  color: #d1fae5;
}

.toast--error .toast-icon {
  color: #ef4444;
}

.toast--error .toast-message {
  color: #fecaca;
}

.toast-body {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.toast-icon {
  flex-shrink: 0;
}

.toast-message {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: normal;
  line-height: 1.4;
  word-break: break-word;
}

.toast-close {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  width: 24px;
  height: 24px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
  transition: color 0.15s;
}

.toast-close:hover {
  color: #d1d5db;
}

/* Transition classes */
.toast-enter-active {
  transition: all 0.25s ease-out;
}

.toast-leave-active {
  transition: all 0.2s ease-in;
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(12px) scale(0.96);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(8px) scale(0.98);
}
</style>
