<template>
  <div class="dash-page" role="main" aria-label="Dashboard content">
    <div class="dash-breadcrumbs">
      <span class="crumb">Workspace</span>
      <ChevronRight :size="12" class="crumb-sep" />
      <span class="crumb crumb--current">Home</span>
    </div>

    <!-- @m4 Greeting — wired to useCurrentUser for real user name/role -->
    <section class="dash-greeting card">
      <div class="dash-avatar">
        <span v-if="displayInitials !== '?'" class="avatar-initials">{{ displayInitials }}</span>
        <User v-else :size="24" />
      </div>
      <div class="dash-greeting-text">
        <h1 class="dash-name">{{ displayName }}</h1>
        <p class="dash-role">Knowledge Engineer</p>
      </div>
      <button class="status-btn" type="button"><Smile :size="14" />Set status</button>
    </section>

    <!-- @m4 Widgets — wired to dashboard.widgets from GQL -->
    <section v-if="loading" class="dash-widgets" aria-label="Loading">
      <article v-for="n in 3" :key="n" class="widget card skeleton">
        <div class="skeleton-line skeleton-line--title"></div>
        <div class="skeleton-line skeleton-line--value"></div>
      </article>
    </section>
    <section v-else-if="error" class="dash-widgets" aria-label="Widgets error">
      <div class="error-state card">
        <p>Failed to load dashboard data.</p>
        <button class="retry-btn" type="button" @click="refetch()">Retry</button>
      </div>
    </section>
    <section v-else class="dash-widgets" aria-label="Collaboration widgets">
      <article v-for="widget in resolvedWidgets" :key="widget.title" class="widget card">
        <header class="widget-head">
          <p class="widget-title">{{ widget.title }}</p>
          <component :is="widget.icon" :size="20" :class="['widget-icon', widget.iconColor]" />
        </header>
        <div class="widget-body">
          <p class="widget-value">{{ widget.value }}</p>
          <p class="widget-subtitle">{{ widget.subtitle }}</p>
          <p class="widget-time">{{ widget.time }}</p>
        </div>
      </article>
    </section>

    <section class="dash-columns">
      <div class="dash-left">
        <!-- @m4 Attention items — wired to dashboard.attentionItems from GQL -->
        <article class="card attention-card">
          <header class="attention-header">
            <h2>Items that need your attention</h2>
            <button class="filter-btn" type="button" @click="toggleActivityFilter">
              <span>{{ activityFilter }}</span>
              <ChevronDown :size="10" />
            </button>
          </header>
          <div v-for="item in resolvedAttentionItems" :key="item.id" class="attention-item">
            <span class="attention-dot" :style="{ background: severityColor(item.severity) }"></span>
            <span class="attention-text">{{ item.text }}</span>
            <span class="attention-time">{{ item.time }}</span>
          </div>
        </article>

        <!-- @m4 Activity feed — wired to dashboard.activityFeed from GQL -->
        <article class="card activity-card">
          <header class="activity-header">
            <h2>Team Activity</h2>
            <div class="activity-toggle">
              <button
                :class="['toggle-btn', { 'toggle-btn--active': activityFilter === 'All team' }]"
                type="button"
                @click="activityFilter = 'All team'"
              >All team</button>
              <button
                :class="['toggle-btn', { 'toggle-btn--active': activityFilter === 'Mine' }]"
                type="button"
                @click="activityFilter = 'Mine'"
              >Mine</button>
            </div>
          </header>
          <div class="sep"></div>
          <template v-for="(group, gi) in resolvedActivityGroups" :key="group.label">
            <p class="activity-group-title">{{ group.label }}</p>
            <div v-for="item in group.items" :key="item.id" class="activity-item">
              <div class="activity-icon" :style="{ background: item.bg }">
                <component :is="item.icon" :size="12" :style="{ color: item.color }" />
              </div>
              <span class="activity-text">{{ item.text }}</span>
              <span v-if="item.user" class="activity-user">{{ item.user }}</span>
              <span class="activity-time">{{ item.time }}</span>
            </div>
            <div v-if="gi < resolvedActivityGroups.length - 1" class="sep"></div>
          </template>
        </article>
      </div>

      <!-- @m4 Recent ontologies — wired to dashboard.recentOntologies from GQL, clickable -->
      <article class="card quick-card">
        <header class="section-header">
          <h2>Recent project</h2>
          <Settings :size="16" class="quick-settings" />
        </header>
        <div class="onto-list">
          <div
            v-for="onto in resolvedRecentOntologies"
            :key="onto.id"
            class="onto-item"
            role="button"
            tabindex="0"
            @click="navigateToOntology(onto.id)"
            @keydown.enter="navigateToOntology(onto.id)"
          >
            <FileText :size="14" class="onto-icon" />
            <div class="onto-body">
              <span class="onto-name">{{ onto.name }}</span>
              <span class="onto-meta">{{ onto.description }} &middot; {{ onto.visibility }}</span>
            </div>
            <ChevronRight :size="12" class="onto-chevron" />
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { DASHBOARD_QUERY } from "@/apollo/queries";
import { useCurrentUser } from "@/composables/useCurrentUser";
import { useQuery } from "@vue/apollo-composable";
import {
	AlertCircle,
	ChevronDown,
	ChevronRight,
	FileText,
	GitMerge,
	MessageSquare,
	Settings,
	Smile,
	User,
} from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();
const { displayName, displayInitials } = useCurrentUser();

const activityFilter = ref("All team");

// @m4 — Wire dashboard to DASHBOARD_QUERY (GraphQL)
const { result, loading, error, refetch } = useQuery(DASHBOARD_QUERY);

// ── Resolvers: map GQL data to UI shapes ──

interface WidgetItem {
	title: string;
	value: string;
	subtitle: string;
	time: string;
	icon: typeof GitMerge;
	iconColor: string;
}

function mapWidgetIcon(iconName: string): typeof GitMerge {
	const icons: Record<string, typeof GitMerge> = {
		"git-merge": GitMerge,
		eye: GitMerge,
		"list-todo": GitMerge,
	};
	return icons[iconName] || GitMerge;
}

const resolvedWidgets = computed<WidgetItem[]>(() => {
	const widgets = result.value?.dashboard?.widgets;
	if (!widgets) return [];
	return widgets.map(
		(w: { title: string; count: number; icon: string; route: string }) => ({
			title: w.title,
			value: String(w.count),
			subtitle: w.route,
			time: "",
			icon: mapWidgetIcon(w.icon),
			iconColor: w.icon === "eye" ? "icon-warning" : "icon-primary",
		}),
	);
});

interface AttentionItem {
	id: string;
	text: string;
	severity: string;
	time: string;
}

const resolvedAttentionItems = computed<AttentionItem[]>(() => {
	const items = result.value?.dashboard?.attentionItems;
	if (!items) return [];
	return items.map(
		(a: { id: string; text: string; severity: string; count: number }) => ({
			id: a.id,
			text: a.text,
			severity: a.severity,
			time: "",
		}),
	);
});

function severityColor(severity: string): string {
	const colors: Record<string, string> = {
		warning: "#f59e0b",
		error: "#ef4444",
		info: "#6366f1",
	};
	return colors[severity] || "#6b7280";
}

interface ActivityGroup {
	label: string;
	items: Array<{
		id: string;
		icon: typeof GitMerge;
		color: string;
		bg: string;
		text: string;
		user: string;
		time: string;
	}>;
}

function mapActivityIcon(type: string): typeof GitMerge {
	const icons: Record<string, typeof GitMerge> = {
		merge_request: GitMerge,
		commit: GitMerge,
		comment: MessageSquare,
	};
	return icons[type] || AlertCircle;
}

function mapActivityColor(type: string): string {
	const colors: Record<string, string> = {
		merge_request: "#6366f1",
		commit: "#10b981",
		comment: "#f59e0b",
	};
	return colors[type] || "#ef4444";
}

const resolvedActivityGroups = computed<ActivityGroup[]>(() => {
	const feed = result.value?.dashboard?.activityFeed;
	if (!feed) return [];
	const items = feed.map(
		(a: {
			id: string;
			text: string;
			author: string;
			timestamp: string;
			type: string;
		}) => ({
			id: a.id,
			icon: mapActivityIcon(a.type),
			color: mapActivityColor(a.type),
			bg: `${mapActivityColor(a.type)}1a`,
			text: a.text,
			user: a.author ? `@${a.author.toLowerCase()}` : "",
			time: formatTimeAgo(a.timestamp),
		}),
	);
	return [{ label: "Recent", items }];
});

interface RecentOntology {
	id: string;
	name: string;
	description: string;
	visibility: string;
}

const resolvedRecentOntologies = computed<RecentOntology[]>(() => {
	const ontos = result.value?.dashboard?.recentOntologies;
	if (!ontos) return [];
	return ontos.map(
		(o: {
			id: string;
			name: string;
			description: string;
			visibility: string;
		}) => ({
			id: o.id,
			name: o.name,
			description: o.description || "",
			visibility: o.visibility || "",
		}),
	);
});

// @m4 — Navigate to ontology workspace via router
function navigateToOntology(ontologyId: string): void {
	console.debug(
		JSON.stringify({
			level: "debug",
			msg: "dashboard.navigate_to_ontology",
			ontologyId,
			ts: new Date().toISOString(),
		}),
	);
	router.push(`/ontology/${ontologyId}/workspace`);
}

function toggleActivityFilter(): void {
	// Toggle between All team / Mine
	activityFilter.value =
		activityFilter.value === "All team" ? "Mine" : "All team";
}

function formatTimeAgo(timestamp: string): string {
	if (!timestamp) return "";
	const now = Date.now();
	const then = new Date(timestamp).getTime();
	const minutes = Math.floor((now - then) / 60000);
	if (minutes < 60) return `${minutes}m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours}h ago`;
	return `${Math.floor(hours / 24)}d ago`;
}
</script>

<style scoped>
.dash-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
}

.dash-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.crumb {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
}

.crumb-sep { color: var(--muted-foreground); }

.dash-greeting {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.dash-avatar {
  width: 64px;
  height: 64px;
  border-radius: 999px;
  background: var(--muted);
  color: var(--muted-foreground);
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-initials {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 24px;
  font-weight: 700;
  color: var(--primary);
}

.dash-greeting-text { flex: 1; }

.dash-name {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 20px;
  font-weight: 600;
}

.dash-role {
  margin: 4px 0 0;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
}

.status-btn {
  height: 32px;
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0 10px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  background: transparent;
  cursor: pointer;
}

.dash-widgets {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.widget {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.widget-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.widget-title {
  margin: 0;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 600;
}

.icon-primary { color: var(--primary); }
.icon-warning { color: var(--warning); }

.widget-body p { margin: 0; }

.widget-value {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 32px;
  font-weight: 600;
}

.widget-subtitle {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
}

.widget-time {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
}

/* @m4 Skeleton loading states */
.skeleton {
  animation: pulse 1.5s ease-in-out infinite;
}

.skeleton-line {
  height: 16px;
  border-radius: 4px;
  background: var(--muted);
}

.skeleton-line--title { width: 60%; }
.skeleton-line--value { width: 40%; height: 32px; margin-top: 8px; }

@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

/* @m4 Error state */
.error-state {
  grid-column: 1 / -1;
  padding: 32px;
  text-align: center;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  color: var(--muted-foreground);
}

.retry-btn {
  margin-top: 12px;
  height: 32px;
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0 12px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  background: var(--card);
  color: var(--foreground);
  cursor: pointer;
}

.retry-btn:hover {
  background: var(--muted);
}

.dash-columns {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 460px;
  gap: 24px;
}

.dash-left {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.attention-card,
.activity-card,
.quick-card {
  display: flex;
  flex-direction: column;
}

.attention-card {
  padding: 18px;
  gap: 14px;
}

.activity-card {
  padding: 20px;
  gap: 4px;
}

.quick-card {
  padding: 20px;
  gap: 16px;
}

.attention-header,
.activity-header,
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.attention-header h2,
.activity-header h2 {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 15px;
  font-weight: 600;
}

.section-header h2 {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 16px;
  font-weight: 600;
}

.filter-btn {
  height: 28px;
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0 10px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  background: transparent;
}

.attention-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
}

.attention-dot {
  width: 6px;
  height: 6px;
  border-radius: 3px;
  flex-shrink: 0;
}

.attention-text {
  flex: 1;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.attention-time {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.activity-toggle {
  height: 28px;
  background: var(--muted);
  border-radius: 6px;
  padding: 0 2px;
  display: flex;
  align-items: center;
  gap: 2px;
}

.toggle-btn {
  height: 24px;
  border: none;
  border-radius: 4px;
  padding: 0 8px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  background: transparent;
}

.toggle-btn--active {
  background: var(--card);
  color: var(--foreground);
  font-weight: 500;
}

.activity-group-title {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  font-weight: 600;
  color: var(--foreground);
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
}

.activity-icon {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.activity-text {
  flex: 1;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.activity-user {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.activity-time {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.sep {
  width: 100%;
  height: 1px;
  background: var(--border);
  margin: 8px 0;
}

.onto-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.onto-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background var(--transition-fast, 0.15s ease);
}

.onto-item:hover,
.onto-item:focus {
  background: var(--muted);
}

.quick-settings {
  color: var(--muted-foreground);
}

.onto-icon {
  color: var(--primary);
  flex-shrink: 0;
}

.onto-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.onto-name {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.onto-meta {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 10px;
  color: var(--muted-foreground);
}

.onto-chevron {
  color: var(--muted-foreground);
  flex-shrink: 0;
}

@media (max-width: 1600px) {
  .dash-widgets { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}

@media (max-width: 1100px) {
  .dash-widgets { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .dash-columns { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .dash-page { padding: 16px; }
  .dash-greeting { align-items: flex-start; flex-wrap: wrap; }
  .dash-widgets { grid-template-columns: 1fr; }
}
</style>
