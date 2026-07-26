<!-- @m4 — MFA Challenge Dialog -->
<!-- @hlv:artifact mfa-challenge-dialog implements GUI-OW-001 -->
<template>
  <Dialog :open="open" title="Two-Factor Authentication" size="sm" :modal="true" @close="$emit('close')">
    <div class="mfa-form">
      <p class="mfa-description">Enter the 6-digit code from your authenticator app.</p>

      <div class="form-group">
        <label class="form-label">Verification Code</label>
        <input
          ref="codeInput"
          v-model="code"
          type="text"
          class="form-input form-input--code"
          placeholder="000000"
          maxlength="6"
          inputmode="numeric"
          autocomplete="one-time-code"
          :disabled="verifying"
          @input="onCodeInput"
        />
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
      </div>

      <button class="resend-link" :disabled="verifying || resendCooldown > 0" @click="resendCode">
        {{ resendCooldown > 0 ? `Resend code in ${resendCooldown}s` : 'Resend code' }}
      </button>
    </div>

    <template #footer>
      <GhostButton :disabled="verifying" @click="$emit('close')">Cancel</GhostButton>
      <PrimaryButton :loading="verifying" :disabled="code.length !== 6" @click="verify">
        {{ verifying ? 'Verifying...' : 'Verify' }}
      </PrimaryButton>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import Dialog from '@/components/ui-kit/Dialog.vue'
import GhostButton from '@/components/ui-kit/GhostButton.vue'
import PrimaryButton from '@/components/ui-kit/PrimaryButton.vue'
import { useErrorPresentation } from '@/composables/useErrorPresentation'
import { onMounted, ref } from 'vue'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: []; verified: [] }>()

const { addError } = useErrorPresentation()

const code = ref('')
const verifying = ref(false)
const errorMessage = ref<string | null>(null)
const resendCooldown = ref(0)
const codeInput = ref<HTMLInputElement | null>(null)

let cooldownTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  if (props.open) {
    codeInput.value?.focus()
  }
})

function onCodeInput(): void {
  errorMessage.value = null
  // Strip non-numeric characters
  code.value = code.value.replace(/\D/g, '').slice(0, 6)
}

async function verify(): Promise<void> {
  console.debug(
    JSON.stringify({
      level: 'debug',
      msg: 'MfaChallenge.verify_started',
      ts: new Date().toISOString()
    })
  )

  if (code.value.length !== 6) {
    errorMessage.value = 'Please enter a 6-digit code'
    return
  }

  verifying.value = true
  errorMessage.value = null

  try {
    await new Promise((resolve) => setTimeout(resolve, 800))
    console.debug(
      JSON.stringify({
        level: 'debug',
        msg: 'MfaChallenge.verify_success',
        ts: new Date().toISOString()
      })
    )
    emit('verified')
    reset()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    errorMessage.value = 'Invalid code. Please try again.'
    addError('MFA-VERIFY-FAILED', msg)
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'MfaChallenge.verify_failed',
        error: msg,
        ts: new Date().toISOString()
      })
    )
  } finally {
    verifying.value = false
  }
}

function resendCode(): void {
  console.debug(
    JSON.stringify({
      level: 'debug',
      msg: 'MfaChallenge.resend',
      ts: new Date().toISOString()
    })
  )

  resendCooldown.value = 30
  if (cooldownTimer) clearInterval(cooldownTimer)
  cooldownTimer = setInterval(() => {
    resendCooldown.value--
    if (resendCooldown.value <= 0) {
      if (cooldownTimer) clearInterval(cooldownTimer)
      cooldownTimer = null
    }
  }, 1000)
}

function reset(): void {
  code.value = ''
  errorMessage.value = null
  verifying.value = false
  if (cooldownTimer) {
    clearInterval(cooldownTimer)
    cooldownTimer = null
  }
  resendCooldown.value = 0
}
</script>

<style scoped>
.mfa-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.mfa-description {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  color: var(--muted-foreground);
  margin: 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-label {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 600;
  color: var(--foreground);
}

.form-input {
  height: 36px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--foreground);
  padding: 0 10px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  outline: none;
}

.form-input:focus {
  border-color: var(--primary);
}

.form-input--code {
  font-size: 24px;
  letter-spacing: 8px;
  text-align: center;
  height: 48px;
}

.form-error {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--danger);
  margin: 0;
}

.resend-link {
  background: none;
  border: none;
  color: var(--primary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  cursor: pointer;
  text-align: left;
  padding: 0;
}

.resend-link:disabled {
  color: var(--muted-foreground);
  cursor: default;
}
</style>
