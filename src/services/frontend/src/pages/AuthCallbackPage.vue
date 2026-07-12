<!-- @ctx: Auth callback page — handles OIDC authorization code response per GUI-LOGIN-001 -->
<!-- @hlv:sec [AUTH_BOUNDARY] — validates state, exchanges code, creates session -->
<template>
  <div class="callback-page">
    <div class="callback-card">
      <div v-if="error" class="callback-error" role="alert">
        <h2>Authentication Failed</h2>
        <p>{{ error }}</p>
        <button class="retry-button" @click="retryLogin">Return to Login</button>
      </div>
      <div v-else class="callback-loading">
        <div class="spinner" aria-hidden="true"></div>
        <p>Signing you in...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { handleCallback } from '@/auth/keycloak'
import { logAuthRedirect } from '@/utils/structured-logger'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    await handleCallback()
    const redirect = new URLSearchParams(window.location.search).get('redirect') || '/dashboard'
    router.replace(redirect)
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : 'Unknown authentication error'
    error.value = message
    logAuthRedirect({ reason: 'callback_error', target: '/auth/callback' })
  }
})

function retryLogin(): void {
  router.replace('/login')
}
</script>

<style scoped>
.callback-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--background);
  padding: var(--space-4);
}

.callback-card {
  width: 100%;
  max-width: 400px;
  background: var(--surface);
  border-radius: var(--radius-xl);
  padding: var(--space-8);
  box-shadow: var(--shadow-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.callback-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-default);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.callback-error {
  text-align: center;
  color: var(--error);
}

.callback-error h2 {
  margin: 0 0 var(--space-2);
  font-size: var(--font-size-lg);
}

.callback-error p {
  margin: 0 0 var(--space-4);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.retry-button {
  padding: var(--space-2) var(--space-4);
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-size-base);
}

.retry-button:hover {
  background: var(--primary-dark, var(--primary));
}
</style>
