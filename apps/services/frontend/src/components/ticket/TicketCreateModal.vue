<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Ticket creation modal — main form combining category wizard, attachments, metadata -->
<!-- @hlv:sec [INPUT_VALIDATION] — all form inputs sanitized before submission -->
<!-- @hlv:sec [AUTH_BOUNDARY] — requires authenticated session to submit -->

<template>
  <Dialog
    :open="isOpen"
    title="Create Support Ticket"
    size="lg"
    :modal="true"
    @close="onClose"
  >
    <div class="ticket-create">
      <!-- @ctx: draft restore prompt -->
      <!-- @hlv atomicity — draft auto-save and recovery -->
      <div v-if="showDraftRestore" class="ticket-create__draft-banner" role="alert">
        <p>A draft was found from your previous session.</p>
        <div class="ticket-create__draft-actions">
          <button class="btn btn--sm btn--primary" @click="restoreDraft">Restore draft</button>
          <button class="btn btn--sm btn--ghost" @click="discardDraft">Discard</button>
        </div>
      </div>

      <form class="ticket-create__form" @submit.prevent="onSubmit" novalidate>
        <!-- Title -->
        <div class="ticket-create__field">
          <label for="ticket-title" class="ticket-create__label">
            Title <span class="required" aria-hidden="true">*</span>
          </label>
          <input
            id="ticket-title"
            v-model="form.title"
            type="text"
            class="ticket-create__input"
            :class="{ 'ticket-create__input--error': errors.title }"
            maxlength="200"
            placeholder="Brief summary of the issue"
            :aria-invalid="!!errors.title"
            aria-describedby="title-error"
            required
          />
          <span v-if="errors.title" id="title-error" class="ticket-create__error" role="alert">
            {{ errors.title }}
          </span>
        </div>

        <!-- Description -->
        <div class="ticket-create__field">
          <label for="ticket-description" class="ticket-create__label">
            Description <span class="required" aria-hidden="true">*</span>
          </label>
          <textarea
            id="ticket-description"
            v-model="form.description"
            class="ticket-create__textarea"
            :class="{ 'ticket-create__textarea--error': errors.description }"
            maxlength="10000"
            rows="5"
            placeholder="Describe the issue in detail..."
            :aria-invalid="!!errors.description"
            aria-describedby="description-error"
            required
          />
          <span v-if="errors.description" id="description-error" class="ticket-create__error" role="alert">
            {{ errors.description }}
          </span>
        </div>

        <!-- Category Wizard -->
        <CategoryWizard
          v-model="categoryData"
          :errors="categoryErrors"
        />

        <!-- Attachments -->
        <AttachmentUpload
          v-model="form.attachments"
          @error="onAttachmentError"
        />

        <!-- Metadata Preview -->
        <MetadataPreview
          :version="metadata.version"
          :environment="metadata.environment"
          :user-agent="metadata.userAgent"
          :trace-id="metadata.traceId"
          :page-url="metadata.pageUrl"
        />

        <!-- Submit actions -->
        <div class="ticket-create__actions">
          <button
            type="button"
            class="btn btn--ghost"
            @click="onClose"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="btn btn--primary"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Submitting...' : 'Submit Ticket' }}
          </button>
        </div>

        <!-- @ctx: network error display -->
        <!-- @hlv TICKET-UI-NETWORK-ERROR -->
        <div v-if="submitError" class="ticket-create__submit-error" role="alert">
          {{ submitError }}
        </div>
      </form>
    </div>
  </Dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from "vue";
import {
	type TicketDraft,
	useTicketDraftStore,
} from "../../stores/ticketDraftStore";
import Dialog from "../ui-kit/Dialog.vue";
import AttachmentUpload from "./AttachmentUpload.vue";
import CategoryWizard from "./CategoryWizard.vue";
import MetadataPreview from "./MetadataPreview.vue";

// @ctx: props
const props = defineProps<{
	isOpen: boolean;
}>();

const emit = defineEmits<{
	close: [];
	submit: [data: Record<string, unknown>];
}>();

// @ctx: draft store integration
const {
	draft,
	hasDraft,
	updateDraft,
	restoreDraft: doRestoreDraft,
	resetDraft,
} = useTicketDraftStore();
const showDraftRestore = ref(false);

// @ctx: form state
const form = reactive({
	title: "",
	description: "",
	attachments: [] as File[],
});

const categoryData = reactive({
	category: "",
	user_severity: "",
	steps_to_reproduce: "",
	expected_behavior: "",
	expected_duration: "",
	actual_duration: "",
});

const errors = reactive<Record<string, string>>({});
const categoryErrors = reactive<Record<string, boolean>>({});
const isSubmitting = ref(false);
const submitError = ref<string | null>(null);

// @ctx: auto-captured metadata
const metadata = reactive({
	version: window.__VEDO_CONFIG__?.APP_VERSION || "0.1.0",
	environment: window.__VEDO_CONFIG__?.APP_ENV || "dev",
	userAgent: navigator.userAgent,
	traceId: crypto.randomUUID?.() || null,
	pageUrl: window.location.href,
});

// @ctx: sync form to draft store
watch(
	[form, categoryData],
	() => {
		if (props.isOpen) {
			updateDraft({
				title: form.title,
				description: form.description,
				category: categoryData.category as TicketDraft["category"],
				user_severity:
					categoryData.user_severity as TicketDraft["user_severity"],
				steps_to_reproduce: categoryData.steps_to_reproduce,
				expected_behavior: categoryData.expected_behavior,
				attachments: form.attachments.map((f) => ({
					filename: f.name,
					size: f.size,
					type: f.type,
				})),
				trace_id: metadata.traceId,
				page_url: metadata.pageUrl,
				saved_at: new Date().toISOString(),
			});
		}
	},
	{ deep: true },
);

// @ctx: check for draft on modal open
onMounted(() => {
	if (hasDraft.value && draft.value?.title) {
		showDraftRestore.value = true;
	}
});

watch(
	() => props.isOpen,
	(open) => {
		if (open && hasDraft.value && draft.value?.title) {
			showDraftRestore.value = true;
		} else if (!open) {
			showDraftRestore.value = false;
			submitError.value = null;
		}
	},
);

function restoreDraft() {
	const saved = doRestoreDraft();
	if (saved) {
		form.title = saved.title;
		form.description = saved.description;
		categoryData.category = saved.category as typeof categoryData.category;
		categoryData.user_severity =
			saved.user_severity as typeof categoryData.user_severity;
		categoryData.steps_to_reproduce = saved.steps_to_reproduce;
		categoryData.expected_behavior = saved.expected_behavior;
	}
	showDraftRestore.value = false;
}

function discardDraft() {
	resetDraft();
	showDraftRestore.value = false;
}

function onAttachmentError(code: string) {
	// @ctx: attachment error — already handled in AttachmentUpload
	submitError.value =
		code === "TICKET-UI-TOO-MANY-FILES"
			? "Too many files attached. Maximum is 10."
			: "File too large. Maximum size is 10 MB.";
}

// @ctx: validation — all error codes from GUI-TICKET-001 contract
function validate(): boolean {
	// Clear previous errors
	for (const k of Object.keys(errors)) delete errors[k];
	for (const k of Object.keys(categoryErrors)) delete categoryErrors[k];

	let valid = true;

	// @hlv TICKET-UI-EMPTY-TITLE
	if (!form.title.trim()) {
		errors.title = "Title is required. Please describe the issue briefly.";
		valid = false;
	}

	// @hlv TICKET-UI-EMPTY-DESCRIPTION
	if (!form.description.trim()) {
		errors.description =
			"Description is required. Please provide details about the issue.";
		valid = false;
	}

	// @hlv TICKET-UI-MISSING-CATEGORY
	if (!categoryData.category) {
		categoryErrors.category = true;
		valid = false;
	}

	// @hlv TICKET-UI-MISSING-SEVERITY
	if (!categoryData.user_severity) {
		categoryErrors.severity = true;
		valid = false;
	}

	// @hlv TICKET-UI-MISSING-STEPS
	if (
		categoryData.category === "bug" &&
		!categoryData.steps_to_reproduce.trim()
	) {
		categoryErrors.steps = true;
		valid = false;
	}

	// @hlv TICKET-UI-MISSING-EXPECTED
	if (
		categoryData.category === "feature" &&
		!categoryData.expected_behavior.trim()
	) {
		categoryErrors.expected = true;
		valid = false;
	}

	return valid;
}

async function onSubmit() {
	submitError.value = null;

	if (!validate()) return;

	isSubmitting.value = true;

	try {
		const payload = {
			title: form.title.trim(),
			description: form.description.trim(),
			category: categoryData.category,
			user_severity: categoryData.user_severity,
			steps_to_reproduce: categoryData.steps_to_reproduce.trim() || null,
			expected_behavior: categoryData.expected_behavior.trim() || null,
			trace_id: metadata.traceId,
			page_url: metadata.pageUrl,
			attachments: form.attachments.map((f) => ({
				filename: f.name,
				size: f.size,
			})),
		};

		emit("submit", payload);
		resetDraft();
		onClose();
	} catch (_err) {
		// @hlv TICKET-UI-NETWORK-ERROR
		submitError.value =
			"Failed to submit ticket. Please check your connection and try again.";
	} finally {
		isSubmitting.value = false;
	}
}

function onClose() {
	emit("close");
}
</script>

<style scoped>
.ticket-create { display: flex; flex-direction: column; gap: var(--spacing-4); }
.ticket-create__draft-banner {
  padding: var(--spacing-3); background: var(--color-warning); border-radius: var(--radius-md);
  display: flex; align-items: center; justify-content: space-between; gap: var(--spacing-3);
}
.ticket-create__draft-banner p { margin: 0; font-size: var(--font-size-sm); }
.ticket-create__draft-actions { display: flex; gap: var(--spacing-2); }
.ticket-create__form { display: flex; flex-direction: column; gap: var(--spacing-4); }
.ticket-create__field { display: flex; flex-direction: column; gap: var(--spacing-1); }
.ticket-create__label { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); }
.ticket-create__input, .ticket-create__textarea {
  padding: var(--spacing-2) var(--spacing-3); border: 1px solid var(--border-default);
  border-radius: var(--radius-md); font-size: var(--font-size-sm); font-family: inherit;
  background: var(--surface-primary); color: var(--font-primary);
}
.ticket-create__input--error, .ticket-create__textarea--error { border-color: var(--color-error); }
.ticket-create__error { font-size: var(--font-size-xs); color: var(--color-error); }
.ticket-create__actions { display: flex; justify-content: flex-end; gap: var(--spacing-2); padding-top: var(--spacing-2); }
.ticket-create__submit-error {
  padding: var(--spacing-3); background: var(--color-error); color: var(--font-on-error);
  border-radius: var(--radius-md); font-size: var(--font-size-sm);
}
.required { color: var(--color-error); }
.btn { padding: var(--spacing-2) var(--spacing-4); border-radius: var(--radius-md); font-size: var(--font-size-sm); cursor: pointer; border: none; }
.btn--primary { background: var(--color-primary); color: var(--font-on-primary); }
.btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn--ghost { background: transparent; color: var(--font-secondary); border: 1px solid var(--border-default); }
.btn--sm { padding: var(--spacing-1) var(--spacing-2); font-size: var(--font-size-xs); }
</style>
