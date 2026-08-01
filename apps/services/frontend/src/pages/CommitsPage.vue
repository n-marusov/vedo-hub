<template>
  <div class="page">
    <h1 class="page-title">{{ t('nav.commits') }}</h1>
    <p class="page-empty">{{ t('commits.no_commits') }}</p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/useI18n";

const { t } = useI18n();
</script>

<style scoped>
.page {
  padding: 24px 32px;
}
.page-title {
  margin: 0 0 16px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 24px;
  font-weight: 600;
}
.page-empty {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
}
</style>
