<!-- @ctx: Versioning page aligned to design/frontend.pen frames Commits/Branches/Compare/Tags/Repository Graph/Merge Requests -->
<template>
  <div class="version-page" role="main" aria-label="Versioning content">
    <section class="version-head">
      <h1 class="version-title">{{ titles[view] || 'Commit History' }}</h1>
      <span v-if="view === 'commits'" class="branch-badge">{{ currentBranch }}</span>
    </section>

    <section class="version-tabs" role="tablist" aria-label="Versioning tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="['tab', { 'tab--active': tab.id === view }]"
        type="button"
        role="tab"
        :aria-selected="tab.id === view"
        @click="go(tab.id)"
      >
        {{ tab.label }}
      </button>
    </section>

    <!-- Loading state -->
    <div v-if="loading" class="version-loading loading-indicator">
      <div class="skeleton" v-for="n in 3" :key="n"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="version-error" role="alert">
      <span>Failed to load versioning data</span>
      <button class="retry-btn" type="button" @click="refetchAll">Retry</button>
    </div>

    <!-- Empty state -->
    <div v-else-if="view === 'commits' && commits.length === 0" class="version-empty">
      No commits yet.
    </div>

    <!-- Data state -->
    <template v-else>
      <section v-if="view === 'commits'" class="filter-row">
        <div class="filter-box">
          <GitBranch :size="14" class="muted" />
          <span class="filter-text">{{ currentBranch }}</span>
        </div>
        <span class="fill"></span>
        <div class="filter-box">
          <User :size="14" class="muted" />
          <span class="filter-text">All authors</span>
        </div>
        <div class="filter-box filter-box--wide">
          <Search :size="14" class="muted" />
          <span class="filter-text">Search by message...</span>
        </div>
      </section>

      <section class="version-card">
        <CommitHistory v-if="view === 'commits'" :commits="commits" />
        <BranchList v-else-if="view === 'branches'" :branches="branches" />
        <DiffView v-else-if="view === 'compare'" :changes="changes" :commit-options="commitOptions" />
        <TagList v-else-if="view === 'tags'" :tags="tags" />
        <RepositoryGraph v-else-if="view === 'graph'" :nodes="graphNodes" />
        <div v-else class="mr-placeholder">Merge requests content goes here</div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import {
	COMPARE_REVISIONS_QUERY,
	GET_BRANCHES_QUERY,
	GET_COMMIT_HISTORY_QUERY,
	GET_TAGS_QUERY,
	GRAPH_NEIGHBORHOOD_QUERY,
} from "@/apollo/queries";
import BranchList from "@/components/organisms/BranchList.vue";
import CommitHistory from "@/components/organisms/CommitHistory.vue";
import DiffView from "@/components/organisms/DiffView.vue";
import RepositoryGraph from "@/components/organisms/RepositoryGraph.vue";
import TagList from "@/components/organisms/TagList.vue";
import { useQuery } from "@vue/apollo-composable";
import { GitBranch, Search, User } from "lucide-vue-next";
import { computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();

const view = computed(() => String(route.params.view || "commits"));
const ontologyId = computed(() => String(route.params.id));

const tabs = [
	{ id: "commits", label: "Commits" },
	{ id: "branches", label: "Branches" },
	{ id: "compare", label: "Compare Revisions" },
	{ id: "tags", label: "Tags" },
	{ id: "graph", label: "Repository Graph" },
	{ id: "merge_requests", label: "Merge Requests" },
];

const titles: Record<string, string> = {
	commits: "Commit History",
	branches: "Branches",
	compare: "Compare Revisions",
	tags: "Tags",
	graph: "Repository Graph",
	merge_requests: "Merge Requests",
};

// ── Apollo Queries ──────────────────────────────────────────────────────────────────

const {
	result: commitResult,
	loading: commitsLoading,
	error: commitsError,
	refetch: refetchCommits,
} = useQuery(
	GET_COMMIT_HISTORY_QUERY,
	() => ({
		ontologyId: ontologyId.value,
		page: 1,
		perPage: 50,
	}),
	{ fetchPolicy: "cache-and-network" },
);

const {
	result: branchesResult,
	loading: branchesLoading,
	error: branchesError,
	refetch: refetchBranches,
} = useQuery(
	GET_BRANCHES_QUERY,
	() => ({
		ontologyId: ontologyId.value,
	}),
	{ fetchPolicy: "cache-and-network" },
);

const {
	result: tagsResult,
	loading: tagsLoading,
	error: tagsError,
	refetch: refetchTags,
} = useQuery(
	GET_TAGS_QUERY,
	() => ({
		ontologyId: ontologyId.value,
	}),
	{ fetchPolicy: "cache-and-network" },
);

const {
	result: compareResult,
	loading: compareLoading,
	error: compareError,
	refetch: refetchCompare,
} = useQuery(
	COMPARE_REVISIONS_QUERY,
	() => ({
		ontologyId: ontologyId.value,
		fromRevision: "",
		toRevision: "",
	}),
	{
		fetchPolicy: "cache-and-network",
		enabled: computed(() => view.value === "compare"),
	},
);

const {
	result: graphResult,
	loading: graphLoading,
	error: graphError,
	refetch: refetchGraph,
} = useQuery(
	GRAPH_NEIGHBORHOOD_QUERY,
	() => ({
		ontologyId: ontologyId.value,
		classId: "",
		depth: 2,
	}),
	{
		fetchPolicy: "cache-and-network",
		enabled: computed(() => view.value === "graph"),
	},
);

// ── Computed Data ───────────────────────────────────────────────────────────────────

const commits = computed(() => {
	const items = commitResult.value?.commits?.items;
	if (!items) return [];
	return items.map((c: Record<string, unknown>) => ({
		author: c.authorName,
		message: c.message,
		sha: c.id,
		date: c.createdAt,
		branch: c.branchId,
	}));
});

const branches = computed(() => {
	const items = branchesResult.value?.branches?.items;
	if (!items) return [];
	return items.map((b: Record<string, unknown>) => ({
		name: b.name,
		status: b.isProtected ? "active" : "default",
		lastCommit: b.headCommitId,
	}));
});

const tags = computed(() => {
	const items = tagsResult.value?.tags;
	if (!items) return [];
	return items.map((t: Record<string, unknown>) => ({
		name: t.name,
		commit: t.commitId,
		description: t.message,
		updated: t.createdAt,
	}));
});

const currentBranch = computed(() => {
	if (branchesResult.value?.branches?.items?.length) {
		const active = branchesResult.value.branches.items.find(
			(b: Record<string, unknown>) => b.isProtected,
		);
		return (active?.name as string) || "main";
	}
	return "main";
});

// Compare view — computed from Apollo compareResult
const changes = computed(() => {
	const data = compareResult.value?.compareRevisions;
	if (!data?.changes?.length) {
		return [];
	}
	return data.changes.map((c: Record<string, unknown>) => ({
		path: String(c.entityLabel || c.entityId || ""),
		type: (c.changeType === "added"
			? "added"
			: c.changeType === "removed"
				? "removed"
				: "modified") as "added" | "removed" | "modified",
		diff: `@@ ${c.field || ""}: ${String(c.oldValue ?? "")} → ${String(c.newValue ?? "")} @@`,
	}));
});

const commitOptions = computed(() => {
	const data = compareResult.value?.compareRevisions;
	if (!data) return [];
	// Derive commit options from comparison metadata if available
	return [];
});

const graphNodes = computed(() => {
	const data = graphResult.value?.graphNeighborhood;
	if (!data?.nodes) return [];
	return data.nodes.map((n: Record<string, unknown>, i: number) => ({
		sha: n.id,
		message: n.label,
		type: "commit" as const,
		x: 120 + i * 80,
		y: 100 + i * 40,
	}));
});

// ── Combined Loading / Error ────────────────────────────────────────────────────────

const loading = computed(() => {
	if (view.value === "commits") return commitsLoading.value;
	if (view.value === "branches") return branchesLoading.value;
	if (view.value === "tags") return tagsLoading.value;
	if (view.value === "compare") return compareLoading.value;
	if (view.value === "graph") return graphLoading.value;
	return false;
});

const error = computed(() => {
	if (view.value === "commits") return commitsError.value;
	if (view.value === "branches") return branchesError.value;
	if (view.value === "tags") return tagsError.value;
	if (view.value === "compare") return compareError.value;
	if (view.value === "graph") return graphError.value;
	return null;
});

function refetchAll(): void {
	refetchCommits();
	refetchBranches();
	refetchTags();
	refetchCompare();
	refetchGraph();
}

// ── Navigation ───────────────────────────────────────────────────────────────────────

function go(next: string): void {
	router.replace({
		name: "ontology-versioning",
		params: { id: route.params.id, view: next },
	});
}

// ── Logging ──────────────────────────────────────────────────────────────────────────

watch(commits, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "versioning.commits.loaded",
			count: val.length,
			ts: new Date().toISOString(),
		}),
	);
});

watch(tags, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "versioning.tags.loaded",
			count: val.length,
			ts: new Date().toISOString(),
		}),
	);
});

watch(changes, (val) => {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "versioning.compare.loaded",
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
				msg: "versioning.query.error",
				error: String(err),
				ts: new Date().toISOString(),
			}),
		);
	}
});
</script>

<style scoped>
.version-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.version-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
}

.version-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

.branch-badge {
  border-radius: 999px;
  border: 1px solid var(--primary);
  color: var(--primary);
  padding: 2px 8px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
}

.version-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tab {
  height: 30px;
  border-radius: 6px;
  border: 1px solid var(--border);
  padding: 0 10px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

.tab--active {
  border-color: var(--primary);
  color: var(--primary);
  background: rgba(16, 185, 129, 0.12);
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 16px;
}

.fill { flex: 1; }

.filter-box {
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0 10px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 180px;
}

.filter-box--wide { width: 520px; }

.filter-text {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.version-card {
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card);
  overflow: hidden;
}

.mr-placeholder {
  padding: 24px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
}

.muted { color: var(--muted-foreground); }

/* @m2.5 Skeleton loading */
.version-loading {
  padding: 16px;
}
.version-loading .skeleton {
  height: 48px;
  border-radius: 6px;
  background: var(--muted);
  margin-bottom: 12px;
}
.version-loading .skeleton:last-child {
  margin-bottom: 0;
}

@media (max-width: 1000px) {
  .filter-row {
    flex-wrap: wrap;
    gap: 10px;
  }
  .fill { display: none; }
  .filter-box,
  .filter-box--wide {
    width: 100%;
    min-width: 0;
  }
}

@media (max-width: 768px) {
  .version-page { padding: 16px; }
}
</style>
