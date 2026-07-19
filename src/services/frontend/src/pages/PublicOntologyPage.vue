<!-- @ctx: Public ontology view — dynamically loaded from public-browse-api -->
<template>
  <div class="public-page" role="main" aria-label="Public Ontology View">
    <header class="public-header">
      <div class="public-brand">
        <img src="/vedo-core-logo-1.jpg" alt="VEDO Core" class="public-brand-logo" />
        <span class="public-brand-text">VEDO Core</span>
      </div>
      <span class="public-center">{{ metadata?.name ?? 'Public Ontology View' }}</span>
      <span class="public-readonly">Read-only</span>
    </header>

    <div v-if="loading" class="public-loading">Loading ontology...</div>
    <div v-else-if="error" class="public-error">{{ error }}</div>

    <div v-else class="public-main">
      <aside class="public-class card-side">
        <div class="public-tools">
          <Search :size="14" class="muted" />
          <div class="panel-input">Filter classes...</div>
        </div>
        <div
          v-for="node in classTree"
          :key="node.id"
          :class="['public-tree-row', { 'public-tree-row--active': selectedClassId === node.id }]"
          @click="selectClass(node.id)"
        >
          <ChevronDown :size="12" />
          <Folder :size="14" />
          {{ node.label }}
        </div>
        <div v-if="classTree.length === 0" class="public-tree-row muted">No classes found</div>
      </aside>

      <section class="public-graph card-side">
        <div class="graph-head">
          <span class="col-ind">Individual</span>
          <span class="col-prop">Property</span>
          <span class="col-val">Value</span>
        </div>
        <div class="public-empty" v-if="metadata">Published ontology with {{ metadata.classCount }} classes, {{ metadata.propertyCount }} properties, {{ metadata.individualCount }} individuals</div>
      </section>

      <aside class="public-props card-side">
        <h2 class="prop-title">Properties</h2>
        <div v-for="prop in properties" :key="prop.id" class="prop-row">
          <span>{{ prop.label }}</span>
          <span class="muted">{{ prop.propertyType }}</span>
        </div>
        <div v-if="properties.length === 0" class="muted prop-row">No properties</div>
      </aside>
    </div>

    <footer class="public-banner">
      <Globe :size="14" class="muted" />
      <span>You are viewing a published snapshot. Edits are disabled.</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { usePublicOntology } from "@/composables/usePublicOntology";
import { ChevronDown, Folder, Globe, Search } from "lucide-vue-next";
import { useRoute } from "vue-router";

const route = useRoute();
const slug = (route.params.id as string) || "default";
const {
	metadata,
	classTree,
	properties,
	loading,
	error,
	selectedClassId,
	selectClass,
} = usePublicOntology(slug);
</script>

<style scoped>
.public-page {
  min-height: 100vh;
  background: var(--background);
  color: var(--foreground);
  display: flex;
  flex-direction: column;
}

.public-loading,
.public-error {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
}

.public-error {
  color: var(--danger);
}

.public-empty {
  padding: 16px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
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
  cursor: pointer;
}

.public-tree-row:hover {
  background: var(--muted);
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
