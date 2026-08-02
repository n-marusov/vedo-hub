<!--
  @hlv:artifact code-publish-browse-ui implements spec-landing-f11
  @ctx: Public landing page (F11.1) per ADR-DES.UI.public-landing-architecture:
  header nav, hero (positioning + dual CTA), 5 demo cards, 5 role cards,
  trust/metrics section, footer. No auth required.
-->
<template>
  <div class="landing" role="main" aria-label="VEDO Hub landing">
    <!-- Header / nav -->
    <header class="landing-header">
      <span class="landing-header__logo">VEDO Hub</span>
      <span class="landing-header__spacer" aria-hidden="true"></span>
      <nav class="landing-header__nav" aria-label="Site navigation">
        <a class="landing-header__link" href="#product">Продукт</a>
        <a class="landing-header__link" href="#pricing">Тарифы</a>
        <a class="landing-header__link" href="#docs">Документация</a>
        <button class="landing-header__login" type="button">Войти</button>
      </nav>
    </header>

    <!-- Hero -->
    <section class="landing-hero">
      <div class="landing-hero__inner">
        <h1 class="landing-hero__title">
          VEDO Hub — GitHub + Hugging Face для онтологий
        </h1>
        <p class="landing-hero__sub">
          Создайте онтологию из документа за 1 минуту — без OWL
        </p>
        <p class="landing-hero__sub">
          Работайте командой: ветки, слияния, ревью — как в Git
        </p>
        <p class="landing-hero__sub">
          Публикуйте с DOI, форкайте, цитируйте
        </p>
        <div class="landing-hero__cta">
          <button class="btn btn--primary" type="button">
            Создать аккаунт бесплатно
          </button>
          <button class="btn btn--outline" type="button" @click="scrollToDemos">
            Смотреть демо →
          </button>
        </div>
      </div>
    </section>

    <!-- Demo showcase -->
    <section id="demos" class="landing-demos">
      <h2 class="landing-section__title">Попробуйте прямо сейчас — без регистрации</h2>
      <p class="landing-section__hint">
        Выберите демо-онтологию и нажмите, чтобы открыть
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
          <span class="demo-card__meta">{{ demo.classCount }} классов</span>
          <span class="demo-card__open">Открыть →</span>
        </button>
      </div>
    </section>

    <!-- Roles -->
    <section id="product" class="landing-roles">
      <h2 class="landing-section__title">Для кого VEDO Hub</h2>
      <div class="landing-roles__grid">
        <div v-for="role in roles" :key="role.title" class="role-card">
          <h3 class="role-card__title">{{ role.title }}</h3>
          <p class="role-card__body">{{ role.body }}</p>
          <button class="btn btn--outline btn--sm" type="button">
            {{ role.cta }}
          </button>
        </div>
      </div>
    </section>

    <!-- Trust / metrics -->
    <section id="pricing" class="landing-trust">
      <div class="landing-metrics">
        <div v-for="m in metrics" :key="m.before" class="metric">
          <span class="metric__before">{{ m.before }}</span>
          <span class="metric__arrow">→</span>
          <span class="metric__after">{{ m.after }}</span>
          <span class="metric__label">{{ m.label }}</span>
        </div>
      </div>
      <div class="landing-open-core">
        <span class="landing-open-core__text">
          Публичные онтологии — бесплатно и навсегда · Приватные проекты,
          Enterprise-поддержка, on-premise — по подписке
        </span>
        <button class="btn btn--primary btn--sm" type="button">Тарифы →</button>
      </div>
    </section>

    <!-- Footer -->
    <footer id="docs" class="landing-footer">
      <div class="landing-footer__cols">
        <div v-for="col in footerCols" :key="col.title" class="footer-col">
          <h4 class="footer-col__title">{{ col.title }}</h4>
          <a v-for="link in col.links" :key="link" class="footer-col__link" href="#">
            {{ link }}
          </a>
        </div>
      </div>
      <div class="landing-footer__bottom">
        <span>© 2026 VEDO Hub</span>
        <span class="landing-footer__sep">·</span>
        <span>v{{ appVersion }}</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { type DemoOntology, VEDO_DEMOS } from "../data/demos";

const emit = defineEmits<{
	openDemo: [demo: DemoOntology];
}>();

const appVersion =
	(window as unknown as { __VEDO_CONFIG__?: { APP_VERSION?: string } })
		.__VEDO_CONFIG__?.APP_VERSION ?? "dev";

const demos = VEDO_DEMOS;

const roles = [
	{
		title: "Бизнес-аналитик / не-IT",
		body: "Создайте онтологию из документа или описания словами. Без знания OWL.",
		cta: "Смотреть демо",
	},
	{
		title: "Учёный / Исследователь",
		body: "Публикуйте с DOI, цитируйте в научных работах. Форкайте чужие онтологии.",
		cta: "Смотреть каталог",
	},
	{
		title: "Инженер знаний / Онтолог",
		body: "1M аксиом, класс с 1000 потомков — <1 сек.",
		cta: "Создать аккаунт",
	},
	{
		title: "Разработчик / Архитектор",
		body: "REST API, SPARQL, MCP-сервер. Полный CRUD программно.",
		cta: "API-документация",
	},
	{
		title: "DevOps / Администратор",
		body: "Docker Compose, Helm, air-gapped, vedo-cli.",
		cta: "Документация",
	},
];

const metrics = [
	{
		before: "45 сек",
		after: "<1 сек",
		label: "Открытие класса с 1000 потомков",
	},
	{ before: "2 часа", after: "10 минут", label: "Разрешение конфликта версий" },
	{
		before: "40-80 часов",
		after: "30 минут",
		label: "Маппинг БД на онтологию",
	},
	{
		before: "12 часов",
		after: "2 минуты",
		label: "Настройка доступа для 12 чел.",
	},
];

const footerCols = [
	{ title: "Продукт", links: ["Возможности", "Тарифы"] },
	{ title: "Ресурсы", links: ["Демо-онтологии", "Блог"] },
	{ title: "Документация", links: ["API-справочник", "Руководства"] },
	{ title: "Компания", links: ["О нас", "Контакты"] },
	{ title: "Сообщество", links: ["Поддержка", "Telegram"] },
];

function scrollToDemos() {
	document.getElementById("demos")?.scrollIntoView({ behavior: "smooth" });
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

/* ── Header ─────────────────────────────────────────────── */
.landing-header {
  display: flex;
  align-items: center;
  gap: 32px;
  height: 56px;
  padding: 0 360px;
  background: #0f151c;
  border-bottom: 1px solid #1c2733;
}

.landing-header__logo {
  font-size: 16px;
  font-weight: 700;
}

.landing-header__spacer {
  flex: 1;
}

.landing-header__nav {
  display: flex;
  align-items: center;
  gap: 32px;
}

.landing-header__link {
  font-size: 13px;
  color: #9aa7b4;
  text-decoration: none;
}

.landing-header__link:hover {
  color: #e6edf3;
}

.landing-header__login {
  padding: 6px 16px;
  border: 1px solid #2a3642;
  border-radius: 6px;
  background: transparent;
  color: #e6edf3;
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
}

/* ── Hero ───────────────────────────────────────────────── */
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
  gap: 12px;
}

.landing-hero__title {
  font-size: 32px;
  font-weight: 700;
  color: #4ade80;
  margin: 0 0 8px;
}

.landing-hero__sub {
  font-size: 18px;
  color: #9aa7b4;
  margin: 0;
}

.landing-hero__cta {
  display: flex;
  gap: 16px;
  justify-content: center;
  margin-top: 16px;
}

/* ── Buttons ─────────────────────────────────────────────── */
.btn {
  padding: 14px 32px;
  border: none;
  border-radius: 8px;
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.btn--primary {
  background: #10b981;
  color: #06281c;
}

.btn--outline {
  background: transparent;
  color: #e6edf3;
  border: 1px solid #2a3642;
  padding: 14px 28px;
}

.btn--sm {
  padding: 8px 16px;
  font-size: 12px;
}

/* ── Sections shared ─────────────────────────────────────── */
.landing-demos,
.landing-roles,
.landing-trust {
  padding: 48px 360px;
  width: 100%;
}

.landing-section__title {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 4px;
}

.landing-section__hint {
  font-size: 14px;
  color: #9aa7b4;
  margin: 0 0 20px;
}

/* ── Demo cards ──────────────────────────────────────────── */
.landing-demos__grid {
  display: flex;
  gap: 24px;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
}

.demo-card {
  flex: 0 0 280px;
  scroll-snap-align: start;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 20px;
  border: 1px solid #2a3642;
  border-radius: 12px;
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

.demo-card__open {
  font-size: 12px;
  font-weight: 600;
  color: #4ade80;
}

/* ── Role cards ──────────────────────────────────────────── */
.landing-roles__grid {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  justify-content: center;
}

.role-card {
  flex: 0 0 280px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 20px;
  border: 1px solid #2a3642;
  border-radius: 12px;
  background: #0f151c;
}

.role-card__title {
  font-size: 15px;
  font-weight: 600;
  margin: 0;
  color: #4ade80;
}

.role-card__body {
  font-size: 13px;
  color: #9aa7b4;
  line-height: 1.5;
  margin: 0;
}

/* ── Trust / metrics ─────────────────────────────────────── */
.landing-trust {
  background: #0f151c;
}

.landing-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 32px;
  justify-content: center;
}

.metric {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  min-width: 180px;
}

.metric__before {
  font-size: 18px;
  font-weight: 700;
}

.metric__arrow {
  color: #4ade80;
}

.metric__after {
  font-size: 18px;
  font-weight: 700;
  color: #4ade80;
}

.metric__label {
  font-size: 12px;
  color: #9aa7b4;
  text-align: center;
}

.landing-open-core {
  margin-top: 32px;
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: center;
  justify-content: center;
  padding: 16px;
  border: 1px solid #2a3642;
  border-radius: 8px;
}

.landing-open-core__text {
  font-size: 13px;
  color: #9aa7b4;
}

/* ── Footer ──────────────────────────────────────────────── */
.landing-footer {
  margin-top: auto;
  padding: 32px 360px;
  border-top: 1px solid #1c2733;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.landing-footer__cols {
  display: flex;
  gap: 48px;
  flex-wrap: wrap;
}

.footer-col {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.footer-col__title {
  font-size: 13px;
  font-weight: 600;
  margin: 0 0 4px;
}

.footer-col__link {
  font-size: 12px;
  color: #9aa7b4;
  text-decoration: none;
}

.footer-col__link:hover {
  color: #e6edf3;
}

.landing-footer__bottom {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #6b7683;
}

.landing-footer__sep {
  color: #2a3642;
}

/* ── Mobile (<768px) ─────────────────────────────────────── */
@media (max-width: 768px) {
  .landing-header,
  .landing-demos,
  .landing-roles,
  .landing-trust,
  .landing-footer {
    padding-left: 24px;
    padding-right: 24px;
  }

  .landing-hero__cta {
    flex-direction: column;
  }

  .landing-demos__grid,
  .landing-roles__grid {
    flex-direction: column;
  }

  .landing-metrics {
    flex-direction: column;
    gap: 16px;
  }
}
</style>
