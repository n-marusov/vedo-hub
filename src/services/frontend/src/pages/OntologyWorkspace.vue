<!-- @ctx: Ontology workspace page strictly mirrored from design/frontend.pen frame ontoWs -->
<template>
  <div class="workspace-page" role="main" aria-label="Ontology Workspace content">
    <div class="workspace-main">
      <aside class="group-sidebar card-side">
        <div class="group-header">Project group</div>
        <ul class="group-list">
          <li>Alice Smith</li>
          <li>Bob Johnson</li>
          <li>Nikolay Marusov</li>
        </ul>
      </aside>

      <div class="splitter"><GripVertical :size="8" /></div>

      <section class="workspace-content">
        <div class="workspace-toolbar">
          <span class="toolbar-title">ProductOntology</span>
          <span class="toolbar-spacer"></span>
          <button class="toolbar-btn" type="button">Publish</button>
          <button class="toolbar-btn toolbar-btn--primary" type="button">Save</button>
        </div>

        <div class="workspace-row">
          <aside class="class-panel card-side">
            <div class="panel-tools">
              <Search :size="14" class="muted" />
              <div class="panel-input">Filter classes...</div>
              <button class="panel-icon-btn" type="button"><Workflow :size="14" /></button>
              <button class="panel-icon-btn" type="button"><Table2 :size="14" /></button>
              <button class="panel-icon-btn" type="button"><Share2 :size="14" /></button>
              <button class="panel-icon-btn" type="button"><Columns3 :size="14" /></button>
            </div>

            <div class="class-list">
              <button class="class-row" type="button"><ChevronDown :size="12" /> <Folder :size="14" /> owl:Thing</button>
              <button class="class-row class-row--active" type="button"><ChevronRight :size="12" /> <User :size="14" /> Person</button>
              <button class="class-row" type="button"><ChevronDown :size="12" /> <Building :size="14" /> Organization</button>
              <button class="class-row" type="button"><ChevronRight :size="12" /> <Package :size="14" /> Product</button>
            </div>
          </aside>

          <div class="splitter"><GripVertical :size="8" /></div>

          <section class="graph-panel card-side">
            <div class="graph-head">
              <span class="col-ind">Individual</span>
              <span class="col-prop">Property</span>
              <span class="col-val">Value</span>
            </div>
            <div class="graph-row graph-row--active"><span class="col-ind">Alice_Johnson</span><span class="col-prop">rdf:type</span><span class="col-val">Person</span></div>
            <div class="graph-row graph-row--active"><span class="col-ind">Bob_Smith</span><span class="col-prop">rdf:type</span><span class="col-val">Person</span></div>
            <div class="graph-row"><span class="col-ind">Acme_Corp</span><span class="col-prop">rdf:type</span><span class="col-val">Organization</span></div>
            <div class="graph-row"><span class="col-ind">Acme_Corp</span><span class="col-prop">hasEmployee</span><span class="col-val">Alice_Johnson</span></div>
            <div class="graph-row"><span class="col-ind">Bob_Smith</span><span class="col-prop">worksFor</span><span class="col-val">Acme_Corp</span></div>
            <div class="graph-row"><span class="col-ind">Order_1042</span><span class="col-prop">createdBy</span><span class="col-val">Bob_Smith</span></div>
          </section>

          <div class="splitter"><GripVertical :size="8" /></div>

          <aside class="property-panel card-side">
            <div class="panel-title">Individuals</div>
            <div class="panel-tools">
              <Search :size="14" class="muted" />
              <div class="panel-input">Filter individuals...</div>
              <button class="panel-icon-btn" type="button"><Columns3 :size="14" /></button>
              <button class="panel-icon-btn" type="button"><Plus :size="14" /></button>
            </div>
            <div class="indiv-head"><span class="i-name">Name</span><span class="i-type">Type</span><span class="i-actions">Actions</span></div>
            <div v-for="item in individuals" :key="item.name" class="indiv-row">
              <span class="i-name">{{ item.name }}</span>
              <span class="i-type i-type--active">Person</span>
              <span class="i-actions"><Pencil :size="13" class="muted" /><Trash2 :size="13" class="danger" /></span>
            </div>
          </aside>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Building,
  ChevronDown,
  ChevronRight,
  Columns3,
  Folder,
  GripVertical,
  Package,
  Pencil,
  Plus,
  Search,
  Share2,
  Table2,
  Trash2,
  User,
  Workflow
} from 'lucide-vue-next'

const individuals = [{ name: 'Alice_Johnson' }, { name: 'Bob_Smith' }, { name: 'John_Doe' }]
</script>

<style scoped>
.workspace-page {
  min-height: calc(100vh - 56px);
}

.workspace-main {
  display: flex;
  height: calc(100vh - 56px);
}

.card-side {
  background: var(--card);
}

.group-sidebar {
  width: 256px;
  border-right: 1px solid var(--border);
  padding: 14px;
}

.group-header {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 10px;
}

.group-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.splitter {
  width: 8px;
  background: #0f0f0f;
  border-left: 1px solid #2a2a2a;
  border-right: 1px solid #2a2a2a;
  color: #4b5563;
  display: flex;
  align-items: center;
  justify-content: center;
}

.workspace-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.workspace-toolbar {
  height: 56px;
  border-bottom: 1px solid var(--border);
  background: var(--card);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}

.toolbar-title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
}

.toolbar-spacer { flex: 1; }

.toolbar-btn {
  height: 32px;
  border-radius: 6px;
  border: 1px solid var(--border);
  padding: 0 10px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.toolbar-btn--primary {
  background: var(--primary);
  border-color: var(--primary);
  color: var(--primary-foreground);
}

.workspace-row {
  flex: 1;
  display: flex;
  min-height: 0;
}

.class-panel {
  width: 320px;
  border-right: 1px solid var(--border);
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.panel-tools {
  display: flex;
  align-items: center;
  gap: 6px;
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

.panel-icon-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #2a2a2a;
  border-radius: 6px;
  color: #6b7280;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.class-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.class-row {
  border-radius: 4px;
  padding: 6px 8px;
  display: flex;
  align-items: center;
  gap: 6px;
  color: #fafafa;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.class-row--active {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.45);
  color: var(--primary);
  font-weight: 600;
}

.graph-panel {
  flex: 1;
  min-width: 0;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}

.graph-head,
.graph-row {
  display: flex;
  align-items: center;
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

.graph-row--active {
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.4);
}

.col-ind { width: 220px; }
.col-prop { width: 240px; }
.col-val { flex: 1; }

.property-panel {
  width: 372px;
  border-left: 1px solid var(--border);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 600;
}

.indiv-head,
.indiv-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  font-family: 'IBM Plex Mono', monospace;
}

.indiv-head {
  color: var(--muted-foreground);
  font-size: 11px;
}

.indiv-row {
  border-radius: 4px;
  font-size: 12px;
}

.i-name { width: 170px; }
.i-type { width: 100px; }
.i-actions { flex: 1; display: flex; justify-content: flex-end; gap: 10px; }

.i-type--active { color: var(--primary); font-weight: 600; }

.muted { color: var(--muted-foreground); }
.danger { color: var(--destructive); }

@media (max-width: 1280px) {
  .group-sidebar,
  .property-panel { display: none; }
  .splitter { display: none; }
  .class-panel { width: 280px; }
}

@media (max-width: 900px) {
  .class-panel { display: none; }
  .workspace-toolbar { height: auto; padding: 10px 12px; flex-wrap: wrap; }
}
</style>
