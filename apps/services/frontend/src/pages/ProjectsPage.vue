<template>
    <div class="projects-page" role="main" aria-label="Projects page">
        <section class="pp-top">
            <div class="pp-title-col">
                <div class="pp-breadcrumbs">
                    <span class="pp-breadcrumb-text">Workspace</span>
                    <ChevronRight :size="12" class="pp-breadcrumb-sep" />
                    <span class="pp-breadcrumb-text">Projects</span>
                </div>
                <h1 class="pp-title">Projects</h1>
            </div>
            <div class="pp-action-row">
                <button class="pp-fork-btn" type="button" @click="showForkDialog = true">
                    <GitFork :size="14" />Fork demo project
                </button>
                <button class="pp-new-btn" type="button" @click="showCreateProjectDialog = true">
                    <Plus :size="14" />New project
                </button>
            </div>
        </section>

        <section class="pp-toolbar">
            <div class="pp-search-wrap">
                <Search :size="14" class="pp-search-icon" />
                <input
                    class="pp-search-input"
                    type="text"
                    v-model="searchQuery"
                    placeholder="Search projects"
                    aria-label="Search projects"
                />
            </div>
            <div class="pp-sort-wrap">
                <button class="pp-sort-btn" type="button" @click="toggleSortField">
                    <span class="pp-sort-label">{{ sortField }}</span>
                    <ChevronDown :size="12" class="pp-sort-chevron" />
                </button>
                <span class="pp-sort-divider"></span>
                <button class="pp-sort-btn" type="button" @click="toggleSortDir">
                    <span class="pp-sort-label">{{ sortDir === 'ASC' ? 'Ascending' : 'Descending' }}</span>
                    <ChevronDown :size="12" class="pp-sort-chevron" />
                </button>
            </div>
        </section>

        <!-- Loading state -->
        <div v-if="loading" class="pp-list">
            <article v-for="n in 3" :key="n" class="pp-row pp-skeleton-row">
                <div class="pp-row-body">
                    <div class="pp-row-body-top">
                        <div class="skeleton skeleton--circle"></div>
                        <div class="skeleton skeleton--text skeleton--name"></div>
                    </div>
                </div>
            </article>
        </div>

        <!-- Error state -->
        <div v-else-if="error" class="pp-error" role="alert">
            <span>Failed to load projects</span>
            <button class="retry-btn" type="button" @click="fetchProjects">Retry</button>
        </div>

        <!-- Empty state -->
        <div v-else-if="projects.length === 0" class="pp-empty">
            <span>No projects found.</span>
        </div>

        <!-- Data state -->
        <section v-else class="pp-list">
            <article v-for="p in projects" :key="p.name" class="pp-row" @click="router.push({ name: 'ontology-workspace', params: { id: p.name } })">
                <div class="pp-row-body">
                    <div class="pp-row-body-top">
                        <Folder :size="20" class="pp-row-folder-icon" />
                        <span
                            class="pp-row-logo"
                            :style="{ background: p.logoBg }"
                            >{{ p.logoLetter }}</span
                        >
                        <span class="pp-row-name project-name">{{ p.name }}</span>
                        <Globe
                            v-if="p.visibility === 'public'"
                            :size="12"
                            class="pp-row-vis-icon"
                        />
                        <Lock
                            v-if="p.visibility === 'private'"
                            :size="12"
                            class="pp-row-vis-icon"
                        />
                        <BadgeCheck
                            v-if="p.verified"
                            :size="18"
                            class="pp-row-verified"
                        />
                    </div>
                    <span class="pp-row-desc" style="padding-left: 68px">{{
                        p.description
                    }}</span>
                    <div class="pp-row-tags" style="padding-left: 68px">
                        <span class="pp-tags-label">Topics:</span>
                        <span v-for="t in p.tags" :key="t" class="pp-tag">{{
                            t
                        }}</span>
                    </div>
                </div>
                <div class="pp-row-actions">
                    <div class="pp-row-counters">
                        <div class="pp-counter">
                            <Star :size="14" /><span>{{ p.stars }}</span>
                        </div>
                        <div class="pp-counter">
                            <GitFork :size="14" /><span>{{ p.forks }}</span>
                        </div>
                        <div class="pp-counter">
                            <GitMerge :size="14" /><span>{{
                                p.mergeRequests
                            }}</span>
                        </div>
                    </div>
                    <span class="pp-row-created">{{ p.created }}</span>
                </div>
                <div class="pp-row-menu-wrap">
                    <MoreVertical :size="16" class="pp-row-menu" />
                </div>
            </article>
        </section>
        <ForkDemoDialog v-model="showForkDialog" />
        <CreateProjectDialog
            :open="showCreateProjectDialog"
            @close="showCreateProjectDialog = false"
            @created="onProjectCreated"
        />
    </div>
</template>

<script setup lang="ts">
import { listProjects } from '@/api/org'
import CreateProjectDialog from '@/components/projects/CreateProjectDialog.vue'
import ForkDemoDialog from '@/components/projects/ForkDemoDialog.vue'
import {
  BadgeCheck,
  ChevronDown,
  ChevronRight,
  Folder,
  GitFork,
  GitMerge,
  Globe,
  Lock,
  MoreVertical,
  Plus,
  Search,
  Star
} from 'lucide-vue-next'
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const showForkDialog = ref(false)
const showCreateProjectDialog = ref(false)

const searchQuery = ref('')
const sortField = ref('Name')
const sortDir = ref('ASC')

const sortByMap: Record<string, string> = {
  Name: 'name',
  'Updated At': 'updatedAt'
}

function toggleSortField(): void {
  sortField.value = sortField.value === 'Name' ? 'Updated At' : 'Name'
  console.debug(
    JSON.stringify({
      level: 'debug',
      msg: 'Projects.sort',
      field: sortField.value,
      dir: sortDir.value,
      ts: new Date().toISOString()
    })
  )
}

function toggleSortDir(): void {
  sortDir.value = sortDir.value === 'ASC' ? 'DESC' : 'ASC'
  console.debug(
    JSON.stringify({
      level: 'debug',
      msg: 'Projects.sort',
      field: sortField.value,
      dir: sortDir.value,
      ts: new Date().toISOString()
    })
  )
}

interface ProjectRow {
  name: string
  visibility: 'public' | 'private'
  description: string
  tags: string[]
  stars: number
  forks: number
  mergeRequests: number
  created: string
  verified: boolean
  logoLetter: string
  logoBg: string
}

const loading = ref(false)
const error = ref<string | null>(null)
async function fetchProjects() {
  loading.value = true
  error.value = null
  try {
    const result = await listProjects({
      q: searchQuery.value || undefined,
      sortBy: sortByMap[sortField.value] || 'name',
      sortDir: sortDir.value,
      page: 1,
      perPage: 50
    })
    projectsData.value = result.items as unknown as Record<string, unknown>[]
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

const projectsData = ref<Record<string, unknown>[]>([])

function onProjectCreated(name: string): void {
  showCreateProjectDialog.value = false
  console.debug(
    JSON.stringify({
      level: 'debug',
      msg: 'Projects.projectCreated',
      projectName: name,
      ts: new Date().toISOString()
    })
  )
  fetchProjects()
}

onMounted(() => {
  fetchProjects()
})

const projects = computed<ProjectRow[]>(() => {
  const items = projectsData.value
  if (!items || items.length === 0) {
    // Fallback to empty when no data from API
    return []
  }
  return items.map((p: Record<string, unknown>) => ({
    name: String(p.name || ''),
    visibility: (p.visibility as 'public' | 'private') || 'public',
    description: String(p.description || ''),
    tags: (p.tags as string[]) || [],
    stars: Number(p.stars || 0),
    forks: Number(p.forks || 0),
    mergeRequests: Number(p.mergeRequests || 0),
    created: String(p.created || ''),
    verified: Boolean(p.verified),
    logoLetter: String(p.name ? (p.name as string)[0] : '?').toUpperCase(),
    logoBg: '#6366f126'
  }))
})

// ── Logging ─────────────────────────────────────────────────────────────────

watch(projects, (val) => {
  console.debug(
    JSON.stringify({
      level: 'debug',
      msg: 'projects.list.loaded',
      count: val.length,
      ts: new Date().toISOString()
    })
  )
})

watch(error, (err) => {
  if (err) {
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'projects.query.error',
        error: String(err),
        ts: new Date().toISOString()
      })
    )
  }
})
</script>

<style scoped>
.projects-page {
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.pp-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.pp-title-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.pp-breadcrumbs {
    display: flex;
    align-items: center;
    gap: 8px;
}

.pp-breadcrumb-text {
    color: var(--muted-foreground);
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
}

.pp-breadcrumb-sep {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.pp-title {
    margin: 0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 24px;
    font-weight: 600;
}

.pp-new-btn {
    height: 36px;
    border-radius: 6px;
    padding: 0 14px;
    background: var(--primary);
    color: var(--primary-foreground);
    font-family: "IBM Plex Mono", monospace;
    font-size: 14px;
    font-weight: 500;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
}

.pp-action-row {
    display: flex;
    align-items: center;
    gap: 8px;
}

.pp-fork-btn {
    height: 36px;
    border-radius: 6px;
    padding: 0 14px;
    background: transparent;
    color: var(--foreground);
    font-family: "IBM Plex Mono", monospace;
    font-size: 14px;
    font-weight: 500;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
    border: 1px solid var(--border);
    cursor: pointer;
}

.pp-fork-btn:hover {
    background: var(--muted);
}

.pp-new-btn:hover {
    background: var(--primary-hover);
}

.pp-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
}

.pp-search-wrap {
    flex: 1;
    height: 36px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--card);
    padding: 0 10px;
    display: flex;
    align-items: center;
    gap: 8px;
}

.pp-search-icon {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.pp-search-input {
    border: none;
    outline: none;
    background: transparent;
    width: 100%;
    font-family: "IBM Plex Mono", monospace;
    font-size: 13px;
    color: var(--foreground);
    opacity: 0.8;
}

.pp-search-input::placeholder {
    color: var(--muted-foreground);
}

.pp-sort-wrap {
    width: 372px;
    height: 36px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--card);
    padding: 0 10px;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
}

.pp-sort-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--muted-foreground);
    font-family: 'IBM Plex Mono', monospace;
    font-size: 12px;
    padding: 4px 8px;
    border-radius: 4px;
}

.pp-sort-btn:hover {
    background: var(--muted);
}

.pp-sort-label {
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    color: var(--muted-foreground);
}

.pp-sort-chevron {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.pp-sort-divider {
    width: 1px;
    height: 18px;
    background: var(--border);
    margin: 0 4px;
    flex-shrink: 0;
}

.pp-list {
    display: flex;
    flex-direction: column;
}

.pp-row {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    display: flex;
    flex-direction: row;
    align-items: stretch;
    gap: 12px;
}

.pp-row:last-child {
    border-bottom: none;
}

.pp-row-body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    justify-content: center;
}

.pp-row-body-top {
    display: flex;
    align-items: center;
    gap: 6px;
}

.pp-row-folder-icon {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.pp-row-logo {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: "IBM Plex Mono", monospace;
    font-size: 16px;
    font-weight: 600;
    color: var(--primary-foreground);
    flex-shrink: 0;
}

.pp-row-name {
    font-family: "IBM Plex Mono", monospace;
    font-size: 14px;
    font-weight: 600;
    color: var(--foreground);
}

.pp-row-vis-icon {
    color: var(--muted-foreground);
    flex-shrink: 0;
}

.pp-row-desc {
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    color: var(--muted-foreground);
}

.pp-row-tags {
    display: flex;
    align-items: center;
    gap: 6px;
}

.pp-tags-label {
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    color: var(--muted-foreground);
}

.pp-tag {
    font-family: "IBM Plex Mono", monospace;
    font-size: 10px;
    color: var(--muted-foreground);
    background: var(--muted);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 4px 6px;
}

.pp-row-verified {
    color: #10b981;
    flex-shrink: 0;
}

.pp-row-actions {
    display: flex;
    flex-direction: column;
    gap: 4px;
    justify-content: center;
    flex-shrink: 0;
}

.pp-row-counters {
    display: flex;
    flex-direction: row;
    gap: 4px;
}

.pp-row-created {
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    color: var(--muted-foreground);
    text-align: right;
}

.pp-counter {
    display: flex;
    align-items: center;
    gap: 4px;
    width: 70px;
    justify-content: center;
    color: #6b7280;
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    font-weight: 500;
    flex-shrink: 0;
}

.pp-counter svg {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
}

.pp-row-menu-wrap {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 32px;
}

.pp-row-menu {
    color: #6b7280;
    flex-shrink: 0;
}

/* Skeleton loading */
.pp-skeleton-row {
    opacity: 0.6;
}

.skeleton {
    background: var(--muted);
    border-radius: 4px;
}

.skeleton--circle {
    width: 36px;
    height: 36px;
    border-radius: 50%;
}

.skeleton--text {
    height: 14px;
    flex: 0 0 200px;
}

.skeleton--name {
    width: 200px;
}

/* Error state */
.pp-error,
.pp-empty {
    padding: 32px;
    text-align: center;
    color: var(--muted-foreground);
    font-family: "IBM Plex Mono", monospace;
    font-size: 14px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--card);
}

.retry-btn {
    margin-top: 12px;
    height: 32px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--card);
    color: var(--foreground);
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    padding: 0 12px;
    cursor: pointer;
}
</style>
