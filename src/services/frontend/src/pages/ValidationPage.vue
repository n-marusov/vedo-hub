<!-- @ctx: Validation page strictly mirrored from design/frontend.pen frame valRep -->
<!-- @m2.5 — Wired to RUN_VALIDATION_MUTATION via Apollo GraphQL -->
<template>
  <div class="validation-page" role="main" aria-label="Validation Report content">
    <section class="validation-title-row">
      <div class="validation-title-wrap">
        <Shield :size="20" class="warning" />
        <h1 class="validation-title">Validation Report</h1>
      </div>
    </section>

    <section class="validation-context">
      <Folder :size="14" class="muted" />
      <span class="context-label">Validating ontology:</span>
      <span class="context-badge context-badge--primary">{{ ontologyId || '—' }}</span>
      <GitBranch :size="14" class="muted" />
      <span class="context-badge">main</span>
      <Calendar :size="14" class="muted" />
      <span class="context-time">Last validation: {{ lastValidatedAt }}</span>
    </section>

    <section class="validation-actions">
      <button
        class="run-btn"
        type="button"
        :disabled="loading"
        @click="runValidation"
      >
        <Loader v-if="loading" :size="14" class="spinning" />
        <Play v-else :size="14" />
        {{ loading ? 'Running...' : 'Run validation' }}
      </button>
    </section>

    <section class="validation-card">
      <ValidationReport v-if="validationResult" :summary="summary" :results="validationResult.violations" />
    </section>
  </div>
</template>

<script setup lang="ts">
import { RUN_VALIDATION_MUTATION } from "@/apollo/queries";
import ValidationReport from "@/components/organisms/ValidationReport.vue";
import { useMutation } from "@vue/apollo-composable";
import {
	Calendar,
	Folder,
	GitBranch,
	Loader,
	Play,
	Shield,
} from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const ontologyId = computed(
	() =>
		(route.params.ontologyId as string) ||
		(route.query.ontologyId as string) ||
		"",
);

// @m2.5 — Wire validation to RUN_VALIDATION_MUTATION
const { mutate, loading } = useMutation(RUN_VALIDATION_MUTATION);

const validationResult = ref<{
	status: string;
	violations: Array<Record<string, unknown>>;
	validatedAt: string;
} | null>(null);

const summary = computed(() => {
	if (!validationResult.value)
		return { total_rules: 0, passed: 0, failed: 0, warnings: 0 };
	const violations = validationResult.value.violations || [];
	const errors = violations.filter(
		(v: Record<string, unknown>) => v.severity === "error",
	).length;
	const warnings = violations.filter(
		(v: Record<string, unknown>) => v.severity === "warning",
	).length;
	const total = violations.length;
	return {
		total_rules: total + errors + warnings + 1,
		passed: Math.max(0, total + errors + warnings + 1 - errors - warnings),
		failed: errors,
		warnings,
	};
});

const lastValidatedAt = computed(() => {
	if (!validationResult.value?.validatedAt) return "Never";
	return new Date(validationResult.value.validatedAt).toLocaleString("en-US", {
		month: "short",
		day: "numeric",
		year: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
	});
});

async function runValidation(): Promise<void> {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "validation.run_started",
			ontologyId: ontologyId.value,
			ts: new Date().toISOString(),
		}),
	);
	try {
		const res = await mutate({ ontologyId: ontologyId.value || "default" });
		validationResult.value = res?.data?.runValidation || {
			status: "ok",
			violations: [],
			validatedAt: new Date().toISOString(),
		};
	} catch (e) {
		console.error(
			JSON.stringify({
				level: "error",
				msg: "validation.run_failed",
				error: String(e),
				ts: new Date().toISOString(),
			}),
		);
	}
}
</script>

<style scoped>
.validation-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.validation-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
}

.validation-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.validation-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

.validation-context {
  border-radius: 6px;
  background: rgba(20, 20, 20, 0.3);
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.context-label,
.context-time {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.context-badge {
  border-radius: 999px;
  border: 1px solid var(--border);
  background: #000;
  padding: 2px 8px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.context-badge--primary {
  border-color: var(--primary);
  color: var(--primary);
}

.validation-actions {
  display: flex;
  justify-content: flex-end;
  padding-right: 8px;
}

.run-btn {
  height: 36px;
  border-radius: 8px;
  background: var(--primary);
  color: var(--primary-foreground);
  border: 1px solid var(--primary);
  padding: 0 14px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.run-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.validation-card {
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card);
  overflow: hidden;
}

/* @m2.5 Spinner */
.spinning {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.warning { color: var(--warning); }
.muted { color: var(--muted-foreground); }

@media (max-width: 768px) {
  .validation-page { padding: 16px; }
  .validation-actions { justify-content: flex-start; padding-right: 0; }
}
</style>
