<!-- @ctx: Public ontology view strictly mirrored from design/frontend.pen frame pubOnt -->
<template>
  <div class="public-page" role="main" aria-label="Public Ontology View">
    <header class="public-header">
      <div class="public-brand">
        <img src="/vedo-core-logo-1.jpg" alt="VEDO Core" class="public-brand-logo" />
        <span class="public-brand-text">VEDO Core</span>
      </div>
      <span class="public-center">Public Ontology View</span>
      <span class="public-readonly">Read-only</span>
    </header>

    <div class="public-main">
      <aside class="public-class card-side">
        <div class="public-tools">
          <Search :size="14" class="muted" />
          <div class="panel-input">Filter classes...</div>
        </div>
        <div class="public-tree-row"><ChevronDown :size="12" /> <Folder :size="14" /> owl:Thing</div>
        <div class="public-tree-row public-tree-row--active"><ChevronRight :size="12" /> <User :size="14" /> Person</div>
        <div class="public-tree-row"><ChevronDown :size="12" /> <Building :size="14" /> Organization</div>
      </aside>

      <section class="public-graph card-side">
        <div class="graph-head">
          <span class="col-ind">Individual</span>
          <span class="col-prop">Property</span>
          <span class="col-val">Value</span>
        </div>
        <div class="graph-row"><span class="col-ind">Alice_Johnson</span><span class="col-prop">rdf:type</span><span class="col-val">Person</span></div>
        <div class="graph-row"><span class="col-ind">Bob_Smith</span><span class="col-prop">rdf:type</span><span class="col-val">Person</span></div>
        <div class="graph-row"><span class="col-ind">Acme_Corp</span><span class="col-prop">hasEmployee</span><span class="col-val">Alice_Johnson</span></div>
      </section>

      <aside class="public-props card-side">
        <h2 class="prop-title">Properties</h2>
        <div class="prop-row"><span>Person</span><span class="muted">Class</span></div>
        <div class="prop-row"><span>Organization</span><span class="muted">Class</span></div>
        <div class="prop-row"><span>worksFor</span><span class="muted">ObjectProperty</span></div>
      </aside>
    </div>

    <footer class="public-banner">
      <Globe :size="14" class="muted" />
      <span>You are viewing a published snapshot. Edits are disabled.</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { Building, ChevronDown, ChevronRight, Folder, Globe, Search, User } from 'lucide-vue-next'
</script>

<style scoped>
.public-page {
  min-height: 100vh;
  background: var(--background);
  color: var(--foreground);
  display: flex;
  flex-direction: column;
}

.public-header {
  height: 56px;
  border-bottom: 1px solid var(--border);
  background: var(--card);
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.public-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.public-brand-logo {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  object-fit: contain;
}

.public-brand-text {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
}

.public-center {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.public-readonly {
  border-radius: 999px;
  border: 1px solid var(--border);
  padding: 3px 10px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.public-main {
  flex: 1;
  display: flex;
  min-height: 0;
}

.card-side { background: var(--card); }

.public-class {
  width: 320px;
  border-right: 1px solid var(--border);
  padding: 8px;
}

.public-tools {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
}

.panel-input {
  flex: 1;
  height: 32px;
  border: 1px solid #2a2a2a;
  background: #101010;
  color: #6b7280;
  border-radius: 6px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.public-tree-row {
  border-radius: 4px;
  padding: 6px 8px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.public-tree-row--active {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.45);
  color: var(--primary);
  font-weight: 600;
}

.public-graph {
  flex: 1;
  min-width: 0;
  border-right: 1px solid var(--border);
}

.graph-head,
.graph-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  font-family: 'IBM Plex Mono', monospace;
}

.graph-head {
  color: var(--muted-foreground);
  font-size: 11px;
  border-bottom: 1px solid var(--border);
}

.graph-row {
  font-size: 12px;
  border-bottom: 1px solid rgba(42, 42, 42, 0.5);
}

.col-ind { width: 220px; }
.col-prop { width: 240px; }
.col-val { flex: 1; }

.public-props {
  width: 372px;
  border-left: 1px solid var(--border);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.prop-title {
  margin: 0 0 6px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
}

.prop-row {
  border-radius: 6px;
  border: 1px solid var(--border);
  padding: 8px 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.public-banner {
  border-top: 1px solid var(--border);
  background: var(--muted);
  padding: 8px 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.muted { color: var(--muted-foreground); }

@media (max-width: 1280px) {
  .public-props { display: none; }
}

@media (max-width: 920px) {
  .public-class { display: none; }
  .public-header { padding: 0 12px; }
  .public-center { display: none; }
}
</style>
