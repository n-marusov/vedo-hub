<!-- @ctx: 404 Not Found page per GUI-OW-001 error GUI_ROUTE_NOT_FOUND -->
<template>
  <div class="not-found-page" role="main" aria-labelledby="nf-title">
    <h1 id="nf-title" class="nf-code">404</h1>
    <p class="nf-message">{{ t('not_found.message') }}</p>
    <PrimaryButton @click="goHome" :aria-label="t('not_found.go_dashboard')">{{ t('not_found.go_dashboard') }}</PrimaryButton>
  </div>
</template>

<script setup lang="ts">
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useI18n } from "@/composables/useI18n";
import { useRouter } from "vue-router";

const { t } = useI18n();
const router = useRouter();

function goHome(): void {
	router.push({ name: "dashboard" });
}
</script>

<style scoped>
.not-found-page { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 60vh; text-align: center; }
.nf-code { font-size: 72px; font-weight: 700; color: var(--color-brand); margin: 0; }
.nf-message { font-size: 18px; color: var(--color-font-secondary); margin: 16px 0 32px; }
</style>
