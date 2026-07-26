<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Comments page — full-page view with real Apollo-backed comment feed -->
<!-- Matches design/pages/comments.pen frame commentsPage -->
<template>
    <div class="comments-page" role="main" aria-label="Comments page">
        <div class="cm-title-col">
            <div class="cm-breadcrumbs">
                <span class="cm-crumb">Workspace</span>
                <ChevronRight :size="12" class="cm-crumb-sep" />
                <span class="cm-crumb">Comments</span>
            </div>
            <h1 class="cm-page-title">Comments</h1>
        </div>

        <div v-if="loading" class="cm-loading">Loading comments...</div>
        <div v-else-if="error" class="cm-error">Failed to load comments: {{ error }}</div>
        <div v-else-if="mutationError" class="cm-error">Failed to send comment: {{ mutationError }}</div>
        <Comments v-else :comments="commentItems" class="cm-section" />

        <div class="cm-new-comment">
            <textarea
                v-model="newCommentText"
                class="cm-new-comment__input"
                placeholder="Write a comment..."
                rows="3"
            />
            <button class="cm-new-comment__submit" :disabled="!newCommentText.trim()" @click="submitComment">
                Send
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import {
	createComment as apiCreateComment,
	listComments as apiListComments,
	type CommentInfo,
} from "@/api/comments";
import Comments from "@/components/organisms/Comments.vue";
import { ChevronRight } from "lucide-vue-next";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const ontologyId = (route.params.ontologyId as string) || "default";

// ── REST Data ────────────────────────────────────────────────────────────────────

const loading = ref(false);
const error = ref<string | null>(null);
const feedData = ref<CommentInfo[]>([]);

async function fetchFeed() {
	loading.value = true;
	error.value = null;
	try {
		const result = await apiListComments(ontologyId, ontologyId, 1, 50);
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

const commentItems = computed(() =>
	feedData.value.map(
		(c: {
			authorName?: string;
			author?: string;
			text?: string;
			entityId?: string;
			entityType?: string;
			createdAt?: string;
			updatedAt?: string;
			parentCommentId?: string;
		}) => ({
			author: c.authorName ?? c.author ?? "",
			handle: `@${c.author ?? ""}`,
			timestamp: formatRelativeTime(c.createdAt ?? ""),
			action: `commented on entity ${c.entityId}`,
			text: c.text ?? "",
		}),
	),
);

const mutationError = ref<string | null>(null);
const newCommentText = ref("");

async function submitComment() {
	if (!newCommentText.value.trim()) return;
	try {
		await apiCreateComment(ontologyId, ontologyId, newCommentText.value.trim());
		newCommentText.value = "";
		await fetchFeed();
	} catch (err: unknown) {
		const message = err instanceof Error ? err.message : "Unknown error";
		mutationError.value = message;
	}
}

function formatRelativeTime(dateStr: string): string {
	const now = Date.now();
	const then = new Date(dateStr).getTime();
	const diffMs = now - then;
	const minutes = Math.floor(diffMs / 60000);
	if (minutes < 1) return "just now";
	if (minutes < 60) return `${minutes}m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours}h ago`;
	const days = Math.floor(hours / 24);
	return `${days}d ago`;
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

.cm-new-comment__submit:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
</style>
