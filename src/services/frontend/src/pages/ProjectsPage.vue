Ь
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
            <button class="pp-new-btn" type="button">
                <Plus :size="14" />New project
            </button>
        </section>

        <section class="pp-toolbar">
            <div class="pp-search-wrap">
                <Search :size="14" class="pp-search-icon" />
                <input
                    class="pp-search-input"
                    type="text"
                    placeholder="Search projects"
                    aria-label="Search projects"
                />
            </div>
            <div class="pp-sort-wrap">
                <span class="pp-sort-label">Name</span>
                <ChevronDown :size="12" class="pp-sort-chevron" />
                <span class="pp-sort-divider"></span>
                <span class="pp-sort-label">Ascending</span>
                <ChevronDown :size="12" class="pp-sort-chevron" />
            </div>
        </section>

        <section class="pp-list">
            <article v-for="p in projects" :key="p.name" class="pp-row">
                <div class="pp-row-body">
                    <div class="pp-row-body-top">
                        <Folder :size="20" class="pp-row-folder-icon" />
                        <span
                            class="pp-row-logo"
                            :style="{ background: p.logoBg }"
                            >{{ p.logoLetter }}</span
                        >
                        <span class="pp-row-name">{{ p.name }}</span>
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
    </div>
</template>

<script setup lang="ts">
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
	Star,
} from "lucide-vue-next";

interface ProjectRow {
	name: string;
	visibility: "public" | "private";
	description: string;
	tags: string[];
	stars: number;
	forks: number;
	mergeRequests: number;
	created: string;
	verified: boolean;
	logoLetter: string;
	logoBg: string;
}

const projects: ProjectRow[] = [
	{
		name: "vedo-core",
		visibility: "public",
		description: "Core VEDO ontology platform",
		tags: ["OWL", "RDF", "Python"],
		stars: 12,
		forks: 8,
		mergeRequests: 3,
		created: "Created 3 months ago",
		verified: true,
		logoLetter: "V",
		logoBg: "#6366f126",
	},
	{
		name: "ontology-service",
		visibility: "private",
		description: "Rust-based ontology storage and query service",
		tags: ["Rust", "Neo4j", "gRPC"],
		stars: 18,
		forks: 5,
		mergeRequests: 1,
		created: "Created 3 months ago",
		verified: false,
		logoLetter: "O",
		logoBg: "#05966926",
	},
	{
		name: "ProductOntology",
		visibility: "public",
		description:
			"Product domain ontology covering classifications and properties",
		tags: ["OWL", "SKOS"],
		stars: 24,
		forks: 12,
		mergeRequests: 6,
		created: "Created 2 months ago",
		verified: true,
		logoLetter: "P",
		logoBg: "#d9770626",
	},
	{
		name: "versioning-service",
		visibility: "private",
		description: "Git-like version control for ontology operations",
		tags: ["Rust", "PostgreSQL"],
		stars: 15,
		forks: 6,
		mergeRequests: 4,
		created: "Created 1 month ago",
		verified: false,
		logoLetter: "V",
		logoBg: "#dc262626",
	},
	{
		name: "OrganizationOntology",
		visibility: "public",
		description: "Organizational structure ontology for departments and roles",
		tags: ["OWL", "RDF"],
		stars: 9,
		forks: 2,
		mergeRequests: 2,
		created: "Created 3 weeks ago",
		verified: false,
		logoLetter: "O",
		logoBg: "#0891b226",
	},
];
</script>

<style scoped>
.projects-page {
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.card {
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
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
</style>
