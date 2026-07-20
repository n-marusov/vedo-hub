<!-- @hlv:artifact code-frontend implements spec-gui-login-001 -->
<!-- @ctx: Login page per design/frontend.pen — centered card, logo header, 5 OAuth provider buttons -->
<!-- @hlv:sec [AUTH_BOUNDARY] — login page renders without authentication -->
<template>
  <div class="login-page" role="main" aria-labelledby="login-title">
    <div class="login-card">
      <!-- Header row: logo + text column -->
      <div class="login-header">
        <img
          src="/vedo-core-logo-1.jpg"
          alt="VEDO Core"
          class="login-logo-icon"
        />
        <div class="login-text-col">
          <h1 class="login-title" id="login-title" role="heading">Sign in to VEDO</h1>
          <span class="login-subtitle">Build, connect, and share knowledge at scale</span>
        </div>
      </div>

      <!-- OAuth provider buttons -->
      <div class="oauth-buttons" aria-label="OAuth providers">
        <button
          v-for="provider in providers"
          :key="provider.id"
          class="oauth-btn"
          :disabled="provider.disabled"
          :aria-disabled="provider.disabled"
          :aria-label="provider.label"
          @click="handleLogin(provider.id)"
        >
          <span class="oauth-btn__label">{{ provider.label }}</span>
        </button>
      </div>

      <p v-if="error" class="login-error" role="alert" aria-live="assertive">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { initiateLogin } from "@/auth/keycloak";
import { ref } from "vue";

// @ctx: OAuth providers per GUI-LOGIN-001 Input provider enum — matches design/frontend.pen
// @hlv:sec [AUTH_BOUNDARY] — only Corporate SSO enabled; external providers disabled until configured
const providers = [
	{ id: "vk" as const, label: "VK ID", disabled: true },
	{ id: "yandex" as const, label: "Yandex ID", disabled: true },
	{ id: "mailru" as const, label: "Mail.ru", disabled: true },
	{ id: "google" as const, label: "Google", disabled: true },
	{ id: "corporate_sso" as const, label: "Corporate SSO", disabled: false },
];

const error = ref<string | null>(null);

// @hlv:sec [INPUT_VALIDATION] — provider validated against known enum before OAuth redirect
async function handleLogin(providerId: string): Promise<void> {
	error.value = null;

	// @hlv LOGIN_PROVIDER_UNSUPPORTED
	const knownProviders = ["vk", "yandex", "mailru", "google", "corporate_sso"];
	if (!knownProviders.includes(providerId)) {
		error.value = "Unsupported OAuth provider.";
		return;
	}

	const provider = providers.find((p) => p.id === providerId);
	if (provider?.disabled) {
		error.value = `${provider.label} is not configured for this instance.`;
		return;
	}

	// @hlv:sec [AUTH_BOUNDARY] — Corporate SSO uses local Keycloak OIDC flow with PKCE
	if (providerId === "corporate_sso") {
		try {
			await initiateLogin();
		} catch (e: unknown) {
			error.value = e instanceof Error ? e.message : "Failed to initiate login";
		}
		return;
	}

	// @hlv LOGIN_SSO_CONFIG_MISSING
	error.value = `${provider?.label} is not configured for this instance.`;
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--background);
}

.login-card {
  width: 720px;
  max-width: 90vw;
  padding: 32px;
  background: var(--card);
  border-radius: 16px;
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Header row */
.login-header {
  display: flex;
  gap: 16px;
  align-items: center;
  width: 100%;
}

.login-logo-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  object-fit: contain;
  flex-shrink: 0;
}

.login-text-col {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.login-title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 18px;
  font-weight: 600;
  color: var(--foreground);
}

.login-subtitle {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 400;
  color: var(--muted-foreground);
}

/* OAuth buttons */
.oauth-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.oauth-btn {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  width: 100%;
  height: 44px;
  padding: 0 16px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 400;
  color: var(--foreground);
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 8px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.oauth-btn:hover:not(:disabled) {
  background: var(--muted);
  border-color: var(--primary);
}

.oauth-btn:focus-visible {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
}

.oauth-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.oauth-btn:not(:disabled) {
  cursor: pointer;
}

.oauth-btn__label {
  white-space: nowrap;
}

/* Error */
.login-error {
  padding: 12px;
  background: var(--error-bg);
  color: var(--destructive);
  border-radius: 8px;
  font-size: 14px;
  font-family: 'IBM Plex Mono', monospace;
}
</style>
