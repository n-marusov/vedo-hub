<!-- @m8 — ForkDemoDialog: fork demo project dialog -->
<!-- Validates: REQ-USR.UI.gui-implementation -->
<!-- @ctx: Modal dialog that lists 5 demo projects from VEDO Demos group -->
<template>
	<Teleport to="body">
		<div
			v-if="modelValue"
			class="fork-dialog-overlay"
			role="dialog"
			aria-label="Fork demo project"
			@click.self="close"
		>
			<div class="fork-dialog-box">
				<div class="fork-dialog-header">
					<h2 class="fork-dialog-title">Fork demo project</h2>
					<p class="fork-dialog-desc">
						Choose a demo project to fork into your workspace
					</p>
					<button
						class="fork-dialog-close"
						type="button"
						aria-label="Close dialog"
						@click="close"
					>
						<X :size="16" />
					</button>
				</div>

				<div class="fork-dialog-search">
					<Search :size="14" class="fork-search-icon" />
					<input
						v-model="searchQuery"
						class="fork-search-input"
						type="text"
						placeholder="Search demo projects..."
						aria-label="Search demo projects"
					/>
				</div>

				<div v-if="loading" class="fork-dialog-loading">
					<span>Loading demo projects...</span>
				</div>

				<div v-else-if="error" class="fork-dialog-error">
					<p class="fork-error-text">{{ error }}</p>
					<button class="fork-retry-btn" type="button" @click="loadProjects">
						Retry
					</button>
				</div>

				<div
					v-else-if="filteredProjects.length === 0"
					class="fork-dialog-empty"
				>
					<Inbox :size="32" class="fork-empty-icon" />
					<p>No demo projects available</p>
				</div>

				<div v-else class="fork-dialog-grid">
					<div
						v-for="(row, ri) in projectRows"
						:key="ri"
						class="fork-dialog-row"
					>
						<div
							v-for="project in row"
							:key="project.id"
							class="fork-demo-card"
						>
							<Folder :size="18" class="fork-card-icon" />
							<h3 class="fork-card-name">{{ project.name }}</h3>
							<p class="fork-card-desc">{{ project.description }}</p>
							<div class="fork-card-meta">
								<span class="fork-card-badge">
									<Box :size="10" />{{ project.classCount }} classes
								</span>
								<span class="fork-card-badge">
									<Pencil :size="10" />{{ project.propertyCount }} properties
								</span>
								<span class="fork-card-badge">
									<Tag :size="10" />{{ project.domain }}
								</span>
							</div>
							<div class="fork-card-sep"></div>
							<button
								:data-testid="`fork-${project.id}`"
								class="fork-card-btn"
								type="button"
								:disabled="forking === project.id"
								@click="handleFork(project)"
							>
								<GitFork v-if="forking !== project.id" :size="12" />
								<Loader v-else :size="12" class="fork-spin" />
								{{ forking === project.id ? "Forking..." : "Fork" }}
							</button>
						</div>
					</div>
				</div>

				<div class="fork-dialog-actions">
					<button
						class="fork-cancel-btn"
						type="button"
						data-testid="cancel-button"
						@click="close"
					>
						Cancel
					</button>
				</div>
			</div>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import { fetchDemoProjects, forkProject } from "@/api/fork";
import type { DemoProject } from "@/api/fork";
import {
	Box,
	Folder,
	GitFork,
	Inbox,
	Loader,
	Pencil,
	Search,
	Tag,
	X,
} from "lucide-vue-next";
import { computed, onMounted, ref } from "vue";

const props = defineProps<{
	modelValue: boolean;
}>();

const emit = defineEmits<{
	"update:modelValue": [value: boolean];
}>();

const demoProjects = ref<DemoProject[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const searchQuery = ref("");
const forking = ref<string | null>(null);

const filteredProjects = computed(() => {
	if (!searchQuery.value) return demoProjects.value;
	const query = searchQuery.value.toLowerCase();
	return demoProjects.value.filter((p) => p.name.toLowerCase().includes(query));
});

/** Split projects into rows of 2 for the card grid layout. */
const projectRows = computed(() => {
	const rows: DemoProject[][] = [];
	for (let i = 0; i < filteredProjects.value.length; i += 2) {
		rows.push(filteredProjects.value.slice(i, i + 2));
	}
	return rows;
});

async function loadProjects() {
	loading.value = true;
	error.value = null;
	try {
		demoProjects.value = await fetchDemoProjects();
	} catch (e) {
		error.value = "Failed to load demo projects. Please try again.";
		console.info("[fork]", "load failed", e);
	} finally {
		loading.value = false;
	}
}

async function handleFork(project: DemoProject) {
	forking.value = project.id;
	error.value = null;
	try {
		const result = await forkProject(project.id);
		console.info(
			JSON.stringify({
				level: "info",
				msg: "[fork]",
				sourceProjectId: project.id,
				newProjectId: result.project_id,
				ts: new Date().toISOString(),
			}),
		);
		window.location.href = `/projects/${result.project_id}/workspace`;
	} catch (e) {
		const axiosError = e as {
			response?: { status?: number; data?: { error?: string } };
		};
		const serverMsg = axiosError.response?.data?.error;
		if (axiosError.response?.status === 403) {
			error.value =
				serverMsg ||
				"Access denied. You don't have permission to fork this project.";
		} else if (axiosError.response?.status === 503) {
			error.value = serverMsg || "Service unavailable, please retry.";
		} else {
			error.value =
				serverMsg || "An unexpected error occurred. Please try again.";
		}
		console.info(
			JSON.stringify({
				level: "warn",
				msg: "[fork] failed",
				sourceProjectId: project.id,
				error: String(e),
				ts: new Date().toISOString(),
			}),
		);
	} finally {
		forking.value = null;
	}
}

function close() {
	emit("update:modelValue", false);
}

onMounted(() => {
	if (props.modelValue) {
		loadProjects();
	}
});
</script>

<style scoped>
.fork-dialog-overlay {
	position: fixed;
	inset: 0;
	z-index: 1000;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(0, 0, 0, 0.5);
}

.fork-dialog-box {
	width: 720px;
	max-height: 80vh;
	overflow-y: auto;
	background: var(--card);
	border: 1px solid var(--border);
	border-radius: 12px;
	padding: 24px;
	display: flex;
	flex-direction: column;
	gap: 20px;
}

.fork-dialog-header {
	position: relative;
}

.fork-dialog-title {
	margin: 0;
	font-family: "IBM Plex Mono", monospace;
	font-size: 20px;
	font-weight: 600;
	color: var(--foreground);
}

.fork-dialog-desc {
	margin: 4px 0 0;
	font-family: "IBM Plex Mono", monospace;
	font-size: 12px;
	color: var(--muted-foreground);
}

.fork-dialog-close {
	position: absolute;
	top: 0;
	right: 0;
	background: none;
	border: none;
	cursor: pointer;
	color: var(--muted-foreground);
	padding: 4px;
}

.fork-dialog-search {
	display: flex;
	align-items: center;
	gap: 8px;
	border: 1px solid var(--input);
	border-radius: 8px;
	padding: 4px 12px;
	background: var(--card);
}

.fork-search-icon {
	color: var(--muted-foreground);
	flex-shrink: 0;
}

.fork-search-input {
	flex: 1;
	border: none;
	background: transparent;
	font-family: "IBM Plex Mono", monospace;
	font-size: 12px;
	color: var(--foreground);
	outline: none;
	padding: 4px 0;
}

.fork-search-input::placeholder {
	color: var(--muted-foreground);
}

.fork-dialog-loading,
.fork-dialog-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 32px;
	gap: 8px;
	color: var(--muted-foreground);
	font-family: "IBM Plex Mono", monospace;
	font-size: 12px;
}

.fork-empty-icon {
	opacity: 0.5;
}

.fork-dialog-error {
	background: var(--secondary);
	border-radius: 8px;
	padding: 12px;
	display: flex;
	align-items: center;
	gap: 12px;
}

.fork-error-text {
	flex: 1;
	margin: 0;
	font-family: "IBM Plex Mono", monospace;
	font-size: 12px;
	color: var(--destructive);
}

.fork-retry-btn {
	border: 1px solid var(--border);
	background: var(--card);
	border-radius: 8px;
	padding: 4px 12px;
	cursor: pointer;
	font-family: "IBM Plex Mono", monospace;
	font-size: 12px;
	color: var(--foreground);
}

.fork-dialog-grid {
	display: flex;
	flex-direction: column;
	gap: 12px;
}

.fork-dialog-row {
	display: flex;
	gap: 12px;
}

.fork-demo-card {
	flex: 1;
	background: var(--card);
	border: 1px solid var(--border);
	border-radius: 8px;
	padding: 12px;
	display: flex;
	flex-direction: column;
	gap: 4px;
}

.fork-card-icon {
	color: var(--primary);
}

.fork-card-name {
	margin: 0;
	font-family: "IBM Plex Mono", monospace;
	font-size: 13px;
	font-weight: 600;
	color: var(--foreground);
}

.fork-card-desc {
	margin: 0;
	font-family: "IBM Plex Mono", monospace;
	font-size: 10px;
	color: var(--muted-foreground);
}

.fork-card-meta {
	display: flex;
	gap: 6px;
	align-items: center;
	padding-bottom: 8px;
}

.fork-card-badge {
	display: inline-flex;
	align-items: center;
	gap: 4px;
	background: var(--secondary);
	border-radius: 4px;
	padding: 2px 6px;
	font-family: "IBM Plex Mono", monospace;
	font-size: 9px;
	color: var(--muted-foreground);
}

.fork-card-sep {
	height: 1px;
	background: var(--border);
	width: 100%;
}

.fork-card-btn {
	width: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 4px;
	background: var(--primary);
	color: var(--primary-foreground);
	border: none;
	border-radius: 6px;
	padding: 4px 8px;
	cursor: pointer;
	font-family: "IBM Plex Mono", monospace;
	font-size: 12px;
	font-weight: 500;
}

.fork-card-btn:disabled {
	opacity: 0.6;
	cursor: not-allowed;
}

.fork-spin {
	animation: spin 1s linear infinite;
}

@keyframes spin {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(360deg);
	}
}

.fork-dialog-actions {
	display: flex;
	justify-content: flex-end;
	gap: 8px;
}

.fork-cancel-btn {
	border: 1px solid var(--border);
	background: transparent;
	border-radius: 6px;
	padding: 8px 16px;
	cursor: pointer;
	font-family: "IBM Plex Mono", monospace;
	font-size: 14px;
	font-weight: 500;
	color: var(--foreground);
}
</style>
