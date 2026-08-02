<!--
  @hlv:artifact code-publish-browse-ui implements spec-landing-f11
  @ctx: Demo ontology view — renders a demo's class structure as a graph/tree.
  No auth required.
-->
<template>
  <div class="demo-view" role="main" aria-label="Demo ontology">
    <div class="demo-view__topbar">
      <button class="demo-view__back" type="button" @click="$emit('back')">
        ← All demos
      </button>
      <span class="demo-view__title">{{ demo?.name }}</span>
    </div>

    <div v-if="demo" class="demo-view__body">
      <p class="demo-view__desc">{{ demo.description }}</p>
      <div class="demo-graph" aria-label="Class graph">
        <div class="demo-graph__root">
          <span class="demo-node demo-node--root">owl:Thing</span>
        </div>
        <div class="demo-graph__children">
          <div
            v-for="cls in demo.classTree"
            :key="cls.id"
            class="demo-graph__child"
          >
            <span class="demo-node">{{ cls.label }}</span>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="demo-view__missing">Demo not found.</div>
  </div>
</template>

<script setup lang="ts">
import type { DemoOntology } from "../data/demos";

defineProps<{
	demo: DemoOntology | null;
}>();

defineEmits<{
	back: [];
}>();
</script>

<style scoped>
.demo-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #0b0f14;
  color: #e6edf3;
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
}

.demo-view__topbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 24px;
  border-bottom: 1px solid #1c2733;
}

.demo-view__back {
  padding: 6px 12px;
  border: 1px solid #2a3642;
  border-radius: 6px;
  background: transparent;
  color: #9aa7b4;
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
}

.demo-view__back:hover {
  color: #e6edf3;
  border-color: #10b981;
}

.demo-view__title {
  font-size: 18px;
  font-weight: 600;
}

.demo-view__body {
  max-width: 720px;
  margin: 0 auto;
  padding: 40px 24px;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.demo-view__desc {
  font-size: 14px;
  color: #9aa7b4;
  margin: 0;
}

.demo-graph {
  border: 1px solid #2a3642;
  border-radius: 10px;
  padding: 32px 24px;
  background: #0f151c;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.demo-graph__root {
  text-align: center;
}

.demo-graph__children {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
}

.demo-node {
  display: inline-block;
  padding: 10px 16px;
  border: 1px solid #2a3642;
  border-radius: 8px;
  background: #0b0f14;
  font-size: 13px;
}

.demo-node--root {
  border-color: #10b981;
  color: #4ade80;
}

.demo-view__missing {
  padding: 48px;
  color: #9aa7b4;
}
</style>
