 <template>
  <div class="dash-page" role="main" aria-label="Dashboard content">
    <div class="dash-breadcrumbs">

      <span class="crumb">Workspace</span>
      <ChevronRight :size="12" class="crumb-sep" />
      <span class="crumb crumb--current">Home</span>
    </div>

    <section class="dash-greeting card">
      <div class="dash-avatar"><User :size="24" /></div>
      <div class="dash-greeting-text">
        <h1 class="dash-name">Nikolay Marusov</h1>
        <p class="dash-role">Knowledge Engineer</p>
      </div>
      <button class="status-btn" type="button"><Smile :size="14" />Set status</button>
    </section>

    <section class="dash-widgets" aria-label="Collaboration widgets">
      <article v-for="widget in widgets" :key="widget.title + widget.subtitle" class="widget card">
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
        <article class="card attention-card">
          <header class="attention-header">
            <h2>Items that need your attention</h2>
            <button class="filter-btn" type="button">
              <span>Everything</span>
              <ChevronDown :size="10" />
            </button>
          </header>
          <div v-for="item in attentionItems" :key="item.text" class="attention-item">
            <span class="attention-dot" :style="{ background: item.color }"></span>
            <span class="attention-text">{{ item.text }}</span>
            <span class="attention-time">{{ item.time }}</span>
          </div>
        </article>

        <article class="card activity-card">
          <header class="activity-header">
            <h2>Team Activity</h2>
            <div class="activity-toggle">
              <button class="toggle-btn toggle-btn--active" type="button">All team</button>
              <button class="toggle-btn" type="button">Mine</button>
            </div>
          </header>
          <div class="sep"></div>
          <template v-for="(group, gi) in activityGroups" :key="group.label">
            <p class="activity-group-title">{{ group.label }}</p>
            <div v-for="item in group.items" :key="item.id" class="activity-item">
              <div class="activity-icon" :style="{ background: item.bg }">
                <component :is="item.icon" :size="12" :style="{ color: item.color }" />
              </div>
              <span class="activity-text">{{ item.text }}</span>
              <span v-if="item.user" class="activity-user">{{ item.user }}</span>
              <span class="activity-time">{{ item.time }}</span>
            </div>
            <div v-if="gi < activityGroups.length - 1" class="sep"></div>
          </template>
        </article>
      </div>

      <article class="card quick-card">
        <header class="section-header">
          <h2>Recent project</h2>
          <Settings :size="16" class="quick-settings" />
        </header>
        <div class="onto-list">
          <div v-for="onto in ontologies" :key="onto.name" class="onto-item">
            <FileText :size="14" class="onto-icon" />
            <div class="onto-body">
              <span class="onto-name">{{ onto.name }}</span>
              <span class="onto-meta">{{ onto.time }} &middot; {{ onto.path }}</span>
            </div>
            <ChevronRight :size="12" class="onto-chevron" />
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
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
  UserCheck
} from 'lucide-vue-next'

const widgets = [
  {
    title: 'Merge requests',
    icon: GitMerge,
    iconColor: 'icon-primary',
    value: '2',
    subtitle: 'Waiting for your review',
    time: 'Just now'
  },
  {
    title: 'Merge requests',
    icon: UserCheck,
    iconColor: 'icon-primary',
    value: '1',
    subtitle: 'Assigned to you',
    time: 'Just now'
  },
  {
    title: 'Active Comments',
    icon: MessageSquare,
    iconColor: 'icon-warning',
    value: '3',
    subtitle: 'Awaiting your reply',
    time: '2 hours ago'
  }
]

const attentionItems = [
  { text: 'Pipeline failed in sparql-gateway', time: '15 min ago', color: '#ef4444' },
  { text: 'MR !42 has merge conflicts', time: '1 hour ago', color: '#f59e0b' },
  { text: 'Deploy failed in billing-demo', time: '3 hours ago', color: '#ef4444' }
]

const activityGroups = [
  {
    label: 'Today',
    items: [
      {
        id: 'act-t1',
        icon: GitMerge,
        color: '#6366f1',
        bg: '#6366f11a',
        text: 'vedo-core — Opened MR: fix/shacl-validation',
        user: '@alice',
        time: '3h ago'
      },
      {
        id: 'act-t2',
        icon: MessageSquare,
        color: '#f59e0b',
        bg: '#f59e0b1a',
        text: 'ontology-engine — Comment on MR !42',
        user: '@bob',
        time: '5h ago'
      }
    ]
  },
  {
    label: 'Yesterday',
    items: [
      {
        id: 'act-y1',
        icon: GitMerge,
        color: '#6366f1',
        bg: '#6366f11a',
        text: 'vedo-core — Merged MR: feat/rdf-optimization',
        user: '@nikolay',
        time: '1d ago'
      },
      {
        id: 'act-y2',
        icon: AlertCircle,
        color: '#ef4444',
        bg: '#ef44441a',
        text: 'sparql-gateway — Pipeline failure',
        user: '',
        time: '1d ago'
      }
    ]
  },
  {
    label: 'Earlier',
    items: [
      {
        id: 'act-e1',
        icon: MessageSquare,
        color: '#f59e0b',
        bg: '#f59e0b1a',
        text: 'vedo-core — Comment on commit a3f2c1',
        user: '@nikolay',
        time: '3d ago'
      }
    ]
  }
]

const ontologies = [
  { name: 'ProductOntology', path: 'ontologies/product/', time: '2h ago' },
  { name: 'OrganizationOntology', path: 'ontologies/organization/', time: 'yesterday' },
  { name: 'CustomerOntology', path: 'ontologies/customer/', time: '3d ago' },
  { name: 'Billing Demo / Master', path: 'billing-demo/master/', time: '6d ago' }
]
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
