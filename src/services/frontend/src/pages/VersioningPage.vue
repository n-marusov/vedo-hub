<!-- @ctx: Versioning page aligned to design/frontend.pen frames Commits/Branches/Compare/Tags/Repository Graph/Merge Requests -->
<template>
  <div class="version-page" role="main" aria-label="Versioning content">
    <section class="version-head">
      <h1 class="version-title">{{ titles[view] || 'Commit History' }}</h1>
      <span v-if="view === 'commits'" class="branch-badge">main</span>
    </section>

    <section class="version-tabs" role="tablist" aria-label="Versioning tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="['tab', { 'tab--active': tab.id === view }]"
        type="button"
        @click="go(tab.id)"
      >
        {{ tab.label }}
      </button>
    </section>

    <section v-if="view === 'commits'" class="filter-row">
      <div class="filter-box">
        <GitBranch :size="14" class="muted" />
        <span class="filter-text">main</span>
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
  </div>
</template>

<script setup lang="ts">
import BranchList from '@/components/organisms/BranchList.vue'
import CommitHistory from '@/components/organisms/CommitHistory.vue'
import DiffView from '@/components/organisms/DiffView.vue'
import RepositoryGraph from '@/components/organisms/RepositoryGraph.vue'
import TagList from '@/components/organisms/TagList.vue'
import { GitBranch, Search, User } from 'lucide-vue-next'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const view = computed(() => String(route.params.view || 'commits'))

const tabs = [
  { id: 'commits', label: 'Commits' },
  { id: 'branches', label: 'Branches' },
  { id: 'compare', label: 'Compare Revisions' },
  { id: 'tags', label: 'Tags' },
  { id: 'graph', label: 'Repository Graph' },
  { id: 'merge_requests', label: 'Merge Requests' }
]

const titles: Record<string, string> = {
  commits: 'Commit History',
  branches: 'Branches',
  compare: 'Compare Revisions',
  tags: 'Tags',
  graph: 'Repository Graph',
  merge_requests: 'Merge Requests'
}

const commits = [
  {
    author: 'Nikolay Marusov',
    message: 'feat: add Product class and relations',
    sha: '15c3035cf9e3a2a',
    date: '2026-05-24T10:30:00Z',
    branch: 'main'
  },
  {
    author: 'Anna Petrova',
    message: 'fix: align validation constraints',
    sha: '4c6c1c25d231b1d',
    date: '2026-05-23T12:10:00Z',
    branch: 'develop'
  }
]

const branches = [
  { name: 'main', status: 'active', lastCommit: '15c3035' },
  { name: 'develop', status: 'active', lastCommit: '4c6c1c2' },
  { name: 'feature/new-property', status: 'stale', lastCommit: 'bd0ab22' }
]

const changes = [
  {
    path: 'ontology/ProductOntology.ttl',
    type: 'modified' as const,
    diff: '@@ Product rdfs:label "Product" @@'
  }
]
const commitOptions = [
  { value: '15c3035', label: '15c3035 - feat: add Product class and relations' },
  { value: '4c6c1c2', label: '4c6c1c2 - fix: align validation constraints' }
]

const tags = [
  {
    name: 'v1.2.0',
    commit: '15c3035',
    description: 'May release',
    updated: '2026-05-24T10:30:00Z'
  },
  {
    name: 'v1.1.0',
    commit: '4c6c1c2',
    description: 'Validation fixes',
    updated: '2026-05-12T09:00:00Z'
  }
]

const graphNodes = [
  {
    sha: '15c3035',
    message: 'feat: add Product class and relations',
    type: 'commit' as const,
    x: 120,
    y: 100
  },
  {
    sha: '4c6c1c2',
    message: 'fix: align validation constraints',
    type: 'commit' as const,
    x: 240,
    y: 180
  }
]

function go(next: string): void {
  router.replace({ name: 'ontology-versioning', params: { id: route.params.id, view: next } })
}
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
