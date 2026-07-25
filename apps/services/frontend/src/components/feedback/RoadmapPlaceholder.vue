<!-- @hlv:artifact code-frontend implements spec-feedback-001 -->
<!-- @ctx: Roadmap placeholder — "coming soon" message for public roadmap and voting -->

<template>
  <div class="roadmap-placeholder" aria-labelledby="roadmap-title">
    <div class="roadmap-placeholder__icon">🗺️</div>
    <h2 id="roadmap-title" class="roadmap-placeholder__title">Public Roadmap</h2>
    <p class="roadmap-placeholder__text">
      We're building a public roadmap where you can see what's coming next and vote on features.
    </p>
    <p class="roadmap-placeholder__coming-soon">Coming soon!</p>

    <!-- Placeholder voting section -->
    <div class="roadmap-placeholder__voting">
      <h3 class="roadmap-placeholder__voting-title">Feature Voting</h3>
      <p class="roadmap-placeholder__voting-text">
        Soon you'll be able to vote on planned features and suggest new ones.
      </p>
      <button class="btn btn--ghost" disabled aria-disabled="true">
        Vote on features (coming soon)
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
// @ctx: placeholder — no functionality, display only
// @hlv atomicity — roadmap and voting are placeholders
</script>

<style scoped>
.roadmap-placeholder {
  display: flex; flex-direction: column; align-items: center; gap: var(--spacing-4);
  padding: var(--spacing-8); text-align: center;
}
.roadmap-placeholder__icon { font-size: 48px; }
.roadmap-placeholder__title { font-size: var(--font-size-lg); font-weight: var(--font-weight-semibold); margin: 0; }
.roadmap-placeholder__text { font-size: var(--font-size-sm); color: var(--text-muted); max-width: 400px; margin: 0; }
.roadmap-placeholder__coming-soon {
  font-size: var(--font-size-md); font-weight: var(--font-weight-semibold);
  color: var(--color-primary); margin: 0;
}
.roadmap-placeholder__voting {
  margin-top: var(--spacing-4); padding: var(--spacing-4); background: var(--surface-secondary);
  border-radius: var(--radius-md); display: flex; flex-direction: column; align-items: center; gap: var(--spacing-2);
}
.roadmap-placeholder__voting-title { font-size: var(--font-size-md); font-weight: var(--font-weight-semibold); margin: 0; }
.roadmap-placeholder__voting-text { font-size: var(--font-size-sm); color: var(--text-muted); margin: 0; }
.btn { padding: var(--spacing-2) var(--spacing-4); border-radius: var(--radius-md); font-size: var(--font-size-sm); cursor: pointer; border: none; }
.btn--ghost { background: transparent; color: var(--font-secondary); border: 1px solid var(--border-default); }
.btn--ghost:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
