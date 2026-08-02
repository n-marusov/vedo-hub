<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Comments page — full-page view with real Apollo-backed comment feed -->
<!-- Matches design/pages/comments.pen frame commentsPage -->
<template>
    <div class="comments-page" role="main" :aria-label="t('comments.page_label')">
        <div class="cm-title-col">
            <div class="cm-breadcrumbs">
                <span class="cm-crumb">{{ t('nav.workspace') }}</span>
                <ChevronRight :size="12" class="cm-crumb-sep" />
                <span class="cm-crumb">{{ t('nav.comments') }}</span>
            </div>
            <h1 class="cm-page-title">{{ t('nav.comments') }}</h1>
        </div>

        <div v-if="loading" class="cm-loading">{{ t('comments.loading') }}</div>
        <div v-else-if="error" class="cm-error">{{ t('comments.load_error', { error }) }}</div>
        <div v-else-if="mutationError" class="cm-error">{{ t('comments.send_error', { error: mutationError }) }}</div>
        <Comments v-else :comments="commentItems" class="cm-section" />

        <div class="cm-new-comment">
            <textarea
                v-model="newCommentText"
                class="cm-new-comment__input"
                :placeholder="t('comments.write_placeholder')"
                rows="3"
                aria-label="New comment"
                @keydown.enter.exact.prevent="submitComment"
            />
            <div v-if="commentError" class="validation-error" role="alert">
                {{ commentError }}
            </div>
            <div class="cm-new-comment__footer">
                <span class="cm-scope">
                    {{ t('comments.scope_label') }}: {{ entityScopeLabel }}
                </span>
                <button class="cm-new-comment__submit" @click="submitComment">
                    {{ t('comments.send') }}
                </button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import {
	type CommentInfo,
	createComment as apiCreateComment,
	listComments as apiListComments,
} from "@/api/comments";
import Comments from "@/components/organisms/Comments.vue";
import { useI18n } from "@/composables/useI18n";
import { ChevronRight } from "@lucide/vue";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();
const ontologyId = (route.params.ontologyId as string) || "default";
// Entity scoping: the route may carry an optional entity context (e.g.
// ?entityId=cls-42&entityType=class). When absent, comments are scoped to the
// ontology itself.
const entityId = ref<string>((route.query.entityId as string) || ontologyId);
const entityType = ref<string>(
	(route.query.entityType as string) || "ontology",
);

const entityScopeLabel = computed(() => {
	if (entityType.value === "class" && entityId.value !== ontologyId) {
		return t("comments.scope_class", { id: entityId.value });
	}
	return t("comments.scope_ontology");
});

// ── REST Data ────────────────────────────────────────────────────────────────────

const loading = ref(false);
const error = ref<string | null>(null);
const feedData = ref<CommentInfo[]>([]);

async function fetchFeed() {
	loading.value = true;
	error.value = null;
	try {
		const result = await apiListComments(ontologyId, entityId.value, 1, 50);
		feedData.value = result.comments;
	} catch (e: unknown) {
		error.value = e instanceof Error ? e.message : String(e);
	} finally {
		loading.value = false;
	}
}

onMounted(() => {
	fetchFeed();
});

// Replies (parentCommentId set) are shown with an "in reply to" prefix.
const commentItems = computed(() =>
	feedData.value.map((c) => ({
		author: c.authorName ?? c.author ?? "",
		handle: `@${c.author ?? ""}`,
		timestamp: formatRelativeTime(c.createdAt ?? ""),
		action: c.parentCommentId
			? t("comments.reply_to", { id: String(c.parentCommentId).slice(0, 8) })
			: t("comments.commented_on_entity", { id: String(c.entityId ?? "") }),
		text: c.text ?? "",
	})),
);

const mutationError = ref<string | null>(null);
const commentError = ref<string | null>(null);
const newCommentText = ref("");

async function submitComment() {
	const text = newCommentText.value.trim();
	// Empty-input validation: inline error, no request is sent.
	if (!text) {
		commentError.value = t("comments.empty_error");
		return;
	}
	commentError.value = null;
	try {
		await apiCreateComment(ontologyId, entityId.value, text);
		newCommentText.value = "";
		await fetchFeed();
	} catch (err: unknown) {
		const message =
			err instanceof Error ? err.message : t("common.unknown_error");
		mutationError.value = message;
	}
}

function formatRelativeTime(dateStr: string): string {
	const now = Date.now();
	const then = new Date(dateStr).getTime();
	const diffMs = now - then;
	const minutes = Math.floor(diffMs / 60000);
	if (minutes < 1) return t("comments.just_now");
	if (minutes < 60) return t("comments.time_ago_m", { n: String(minutes) });
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return t("comments.time_ago_h", { n: String(hours) });
	const days = Math.floor(hours / 24);
	return t("comments.time_ago_d", { n: String(days) });
}

// ── Logging ─────────────────────────────────────────────────────────────────

watch(commentItems, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "comments.feed.loaded",
			count: val.length,
			ts: new Date().toISOString(),
		}),
	);
});

watch(error, (err) => {
	if (err) {
		console.error(
			JSON.stringify({
				level: "error",
				msg: "comments.query.error",
				error: String(err),
				ts: new Date().toISOString(),
			}),
		);
	}
});
</script>

<style scoped>
.comments-page {
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.cm-title-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.cm-breadcrumbs {
    display: flex;
    align-items: center;
    gap: 8px;
}

.cm-crumb {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 500;
    color: var(--muted-foreground);
}

.cm-crumb-sep {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.cm-page-title {
    margin: 0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 24px;
    font-weight: 600;
    color: var(--foreground);
}

.cm-loading,
.cm-error {
    font-family: "IBM Plex Mono", monospace;
    font-size: 13px;
    padding: 16px;
    color: var(--muted-foreground);
}

.cm-error {
    color: var(--danger);
}

.cm-section {
    width: 100%;
}

.cm-new-comment {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 16px;
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--card);
}

.cm-new-comment__input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--background);
    color: var(--foreground);
    font-family: "IBM Plex Mono", monospace;
    font-size: 13px;
    resize: vertical;
}

.cm-new-comment__submit {
    align-self: flex-end;
    padding: 6px 16px;
    background: var(--primary);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
}

.cm-new-comment__footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
}

.cm-scope {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    color: var(--muted-foreground);
}

.validation-error {
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    color: var(--danger);
    padding: 4px 0;
}

.cm-new-comment__submit:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
</style>
