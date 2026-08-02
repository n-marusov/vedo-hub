<!--
  @hlv:artifact code-publish-browse-ui implements spec-landing-f11
  @ctx: Public landing page (F11.1) — hero, email CTA, role-based audience
  sections, and a clickable demo ontology showcase. No auth required.
-->
<template>
  <div class="landing" role="main" aria-label="VEDO Hub landing">
    <!-- Hero -->
    <header class="landing-hero">
      <div class="landing-hero__inner">
        <p class="landing-hero__eyebrow">VEDO Hub</p>
        <h1 class="landing-hero__title">GitHub for ontologies</h1>
        <p class="landing-hero__subtitle">
          Create, govern, version, and publish knowledge graphs with your team.
          Branch, merge, review, and release ontologies like code.
        </p>
        <form class="landing-cta" @submit.prevent="onEmailSubmit">
          <input
            v-model="email"
            class="landing-cta__input"
            type="email"
            placeholder="you@company.com"
            aria-label="Work email"
          />
          <button class="landing-cta__btn" type="submit">
            Request early access
          </button>
        </form>
        <p v-if="emailStatus" class="landing-cta__status" role="status">
          {{ emailStatus }}
        </p>
      </div>
    </header>

    <!-- Audience sections -->
    <section class="landing-audience" aria-label="Who it is for">
      <div class="landing-audience__grid">
        <div v-for="aud in audiences" :key="aud.title" class="audience-card">
          <h3 class="audience-card__title">{{ aud.title }}</h3>
          <p class="audience-card__body">{{ aud.body }}</p>
        </div>
      </div>
    </section>

    <!-- Demo showcase -->
    <section class="landing-demos" aria-label="Demo ontologies">
      <h2 class="landing-demos__title">Try a demo ontology</h2>
      <p class="landing-demos__hint">
        Explore a ready-made ontology — no account needed.
      </p>
      <div class="landing-demos__grid">
        <button
          v-for="demo in demos"
          :key="demo.id"
          class="demo-card"
          type="button"
          @click="openDemo(demo.id)"
        >
          <span class="demo-card__name">{{ demo.name }}</span>
          <span class="demo-card__desc">{{ demo.description }}</span>
          <span class="demo-card__meta">{{ demo.classCount }} classes</span>
        </button>
      </div>
    </section>

    <!-- Footer -->
    <footer class="landing-footer">
      <span>VEDO Hub — ontology platform</span>
      <span class="landing-footer__sep">·</span>
      <span>v{{ appVersion }}</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { type DemoOntology, VEDO_DEMOS } from "../data/demos";

const emit = defineEmits<{
	openDemo: [demo: DemoOntology];
}>();

const appVersion =
	(window as unknown as { __VEDO_CONFIG__?: { APP_VERSION?: string } })
		.__VEDO_CONFIG__?.APP_VERSION ?? "dev";

const demos = VEDO_DEMOS;

const audiences = [
	{
		title: "Ontology engineers",
		body: "Visual editing with Git-like versioning: branches, commits, diffs, and merge reviews.",
	},
	{
		title: "IT architects",
		body: "Publish governed knowledge graphs with visibility control and release pipelines.",
	},
	{
		title: "Data analysts",
		body: "Query ontologies with SPARQL and browse class hierarchies without setup.",
	},
];

const email = ref("");
const emailStatus = ref("");

function onEmailSubmit() {
	const value = email.value.trim();
	if (!value) return;
	// MVP: capture interest locally. A backend subscription endpoint is post-MVP.
	emailStatus.value = `Thanks — we'll notify ${value} when early access opens.`;
	email.value = "";
}

function openDemo(id: string) {
	const demo = VEDO_DEMOS.find((d) => d.id === id);
	if (demo) emit("openDemo", demo);
}
</script>

<style scoped>
.landing {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #0b0f14;
  color: #e6edf3;
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
}

.landing-hero {
  padding: 96px 24px 72px;
  text-align: center;
  background: radial-gradient(ellipse at top, #12202b 0%, #0b0f14 60%);
}

.landing-hero__inner {
  max-width: 720px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.landing-hero__eyebrow {
  font-size: 13px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: #4ade80;
  margin: 0;
}

.landing-hero__title {
  font-size: 44px;
  font-weight: 700;
  line-height: 1.1;
  margin: 0;
}

.landing-hero__subtitle {
  font-size: 16px;
  color: #9aa7b4;
  line-height: 1.6;
  margin: 0;
}

.landing-cta {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-top: 8px;
}

.landing-cta__input {
  padding: 10px 14px;
  border: 1px solid #2a3642;
  border-radius: 6px;
  background: #0b0f14;
  color: #e6edf3;
  font-family: inherit;
  font-size: 14px;
  width: 260px;
}

.landing-cta__btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  background: #10b981;
  color: #06281c;
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.landing-cta__status {
  font-size: 13px;
  color: #4ade80;
  margin: 0;
}

.landing-audience {
  padding: 48px 24px;
  max-width: 960px;
  margin: 0 auto;
  width: 100%;
}

.landing-audience__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}

.audience-card {
  border: 1px solid #2a3642;
  border-radius: 8px;
  padding: 20px;
  background: #0f151c;
}

.audience-card__title {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 8px;
  color: #4ade80;
}

.audience-card__body {
  font-size: 13px;
  color: #9aa7b4;
  line-height: 1.5;
  margin: 0;
}

.landing-demos {
  padding: 48px 24px 64px;
  max-width: 960px;
  margin: 0 auto;
  width: 100%;
}

.landing-demos__title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 4px;
}

.landing-demos__hint {
  font-size: 14px;
  color: #9aa7b4;
  margin: 0 0 20px;
}

.landing-demos__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px;
}

.demo-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 18px;
  border: 1px solid #2a3642;
  border-radius: 8px;
  background: #0f151c;
  color: #e6edf3;
  font-family: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.15s ease;
}

.demo-card:hover {
  border-color: #10b981;
}

.demo-card__name {
  font-size: 16px;
  font-weight: 600;
}

.demo-card__desc {
  font-size: 12px;
  color: #9aa7b4;
}

.demo-card__meta {
  font-size: 11px;
  color: #4ade80;
}

.landing-footer {
  margin-top: auto;
  padding: 20px 24px;
  border-top: 1px solid #1c2733;
  display: flex;
  gap: 8px;
  justify-content: center;
  font-size: 12px;
  color: #6b7683;
}

.landing-footer__sep {
  color: #2a3642;
}
</style>
