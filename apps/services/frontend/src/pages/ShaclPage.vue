<!-- @m4 — SHACL Rule Builder page -->
<!-- @hlv:artifact shacl-page implements GUI-OW-001 -->
<template>
  <div class="shacl-page" role="main" :aria-label="t('shacl.title')">
    <section class="shacl-title-row">
      <div class="shacl-title-wrap">
        <Shield :size="20" class="warning" />
        <h1 class="shacl-title">{{ t('shacl.title') }}</h1>
      </div>
    </section>

    <nav class="shacl-breadcrumbs" :aria-label="t('shacl.breadcrumbs')">
      <router-link
        :to="{ name: 'ontology-workspace', params: { id: ontologyId } }"
        class="breadcrumb-link"
      >
        {{ t('nav.workspace') }}
      </router-link>
      <ChevronRight :size="12" class="breadcrumb-sep" />
      <span class="breadcrumb-current">{{ ontologyName || ontologyId }}</span>
      <ChevronRight :size="12" class="breadcrumb-sep" />
      <span class="breadcrumb-current">{{ t('shacl.title') }}</span>
    </nav>

    <!-- Loading state -->
    <div v-if="loading" class="shacl-loading">
      <div v-for="n in 3" :key="n" class="skeleton-rule">
        <div class="skeleton-line skeleton-line--title" />
        <div class="skeleton-line skeleton-line--detail" />
        <div class="skeleton-line skeleton-line--detail" />
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="shacl-error">
      <AlertTriangle :size="24" class="error-icon" />
      <p class="error-message">{{ error }}</p>
      <PrimaryButton @click="retry">{{ t('common.retry') }}</PrimaryButton>
    </div>

    <!-- Empty state -->
    <div v-else-if="rules.length === 0" class="shacl-empty">
      <FileText :size="32" class="empty-icon" />
      <p class="empty-message">{{ t('shacl.no_rules') }}</p>
    </div>

    <!-- Data state -->
    <div v-else class="shacl-content">
      <SHACLRuleBuilder :rules="rules" @run-validation="runValidation" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { runValidation as apiRunValidation } from "@/api/validation";
import type { ValidationViolation } from "@/api/validation";
import SHACLRuleBuilder from "@/components/organisms/SHACLRuleBuilder.vue";
import PrimaryButton from "@/components/ui-kit/PrimaryButton.vue";
import { useErrorPresentation } from "@/composables/useErrorPresentation";
import { useI18n } from "@/composables/useI18n";
import { gql } from "@apollo/client";
import { AlertTriangle, ChevronRight, FileText, Shield } from "@lucide/vue";
import { useQuery } from "@vue/apollo-composable";
import { computed, ref } from "vue";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();
const ontologyId = computed(() => (route.params.id as string) || "");
const ontologyName = computed(() => (route.query.name as string) || "");

const { addError } = useErrorPresentation();

// ── SHACL rules query ────────────────────────────────────────────────────────
// Placeholder — real SHACL rule storage is post-M4
const SHACL_RULES_QUERY = gql`
  query GetShaclRules($ontologyId: ID!) {
    shaclRules(ontologyId: $ontologyId) {
      id
      name
      severity
      target
      condition
      action
    }
  }
`;

const {
	result,
	loading,
	error: queryError,
	refetch,
} = useQuery(SHACL_RULES_QUERY, { ontologyId: ontologyId.value || "default" });

const rules = computed(() => {
	if (!result.value?.shaclRules) return [];
	return result.value.shaclRules as Array<{
		id: string;
		name: string;
		severity: string;
		target: string;
		condition: string;
		action: string;
	}>;
});

const error = computed(() => {
	if (!queryError.value) return null;
	const msg = queryError.value.message || "Failed to load SHACL rules";
	addError("SHACL-RULES-LOAD-FAILED", msg);
	return msg;
});

// ── Validation handler ───────────────────────────────────────────────────────

const validationResult = ref<{
	status: string;
	violations: ValidationViolation[];
	validatedAt: string;
} | null>(null);

const validating = ref(false);

async function runValidation(): Promise<void> {
	if (validating.value) return;
	validating.value = true;
	try {
		const report = await apiRunValidation(ontologyId.value || "default");
		const { status, validatedAt, violations } = report;
		console.debug(
			JSON.stringify({
				level: "debug",
				msg: "Shacl.page.validation_completed",
				status,
				validatedAt,
				violationsCount: violations?.length || 0,
				ts: new Date().toISOString(),
			}),
		);
		if (status === "ok" && (!violations || violations.length === 0)) {
			validationResult.value = {
				status: "ok",
				violations: [],
				validatedAt: validatedAt ?? new Date().toISOString(),
			};
		} else {
			validationResult.value = {
				status: status ?? "error",
				violations: violations ?? [],
				validatedAt: validatedAt ?? new Date().toISOString(),
			};
		}
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		addError("SHACL-VALIDATION-FAILED", msg);
		console.error(
			JSON.stringify({
				level: "error",
				msg: "Shacl.page.validation_failed",
				error: msg,
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		validating.value = false;
	}
}

async function retry(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "Shacl.page.retry",
			ontologyId: ontologyId.value,
			ts: new Date().toISOString(),
		}),
	);
	await refetch();
}
</script>

<style scoped>
.shacl-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.shacl-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
}

.shacl-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.shacl-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

/* Breadcrumbs */
.shacl-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: var(--muted-foreground);
}

.breadcrumb-link {
  color: var(--primary);
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.breadcrumb-sep {
  color: var(--muted-foreground);
}

.breadcrumb-current {
  color: var(--foreground);
}

/* Loading skeleton */
.shacl-loading {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-rule {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-line {
  height: 12px;
  border-radius: 4px;
  background: rgba(128, 128, 128, 0.15);
  animation: pulse 1.5s ease-in-out infinite;
}

.skeleton-line--title {
  width: 40%;
  height: 16px;
}

.skeleton-line--detail {
  width: 70%;
}

/* Error state */
.shacl-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  text-align: center;
}

.error-icon {
  color: var(--danger);
}

.error-message {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
  max-width: 400px;
}

/* Empty state */
.shacl-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  text-align: center;
}

.empty-icon {
  color: var(--muted-foreground);
}

.empty-message {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
}

/* Content area */
.shacl-content {
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card);
  overflow: hidden;
}

@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}

.warning { color: var(--warning); }

@media (max-width: 768px) {
  .shacl-page { padding: 16px; }
}
</style>
