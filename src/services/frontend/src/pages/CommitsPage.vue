<template>
  <div class="page">
    <h1 class="page-title">Commits</h1>
    <p class="page-empty">No commits yet.</p>
  </div>
</template>

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
