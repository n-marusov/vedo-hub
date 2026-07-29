<template>
  <aside
    :class="['group-sidebar', { 'group-sidebar--collapsed': collapsed }]"
    role="complementary"
    :aria-label="collapsed ? 'Group sidebar collapsed' : 'Group sidebar'"
  >
    <template v-if="!collapsed">
      <nav class="group-nav">
        <div class="sidebar-header">
          <span class="sidebar-header-text">Group</span>
        </div>

        <button class="sidebar-item" type="button" @click="navigate(`/${groupName}`)">
          <Folders :size="16" class="sidebar-item-icon" />
          <span class="sidebar-item-label">{{ groupNameLabel }}</span>
        </button>

        <div class="sidebar-divider nav-divider"></div>

        <button
          v-for="item in navItems"
          :key="item.label"
          class="sidebar-item"
          type="button"
          @click="navigate(item.to)"
        >
          <component :is="item.icon" :size="16" class="sidebar-item-icon" />
          <span class="sidebar-item-label">{{ item.label }}</span>
          <span v-if="item.badge != null" class="sidebar-badge">{{ item.badge }}</span>
        </button>
      </nav>

      <button class="sidebar-item" type="button" @click="navigate(`/${groupName}/settings`)">
        <Cog :size="16" class="sidebar-item-icon" />
        <span class="sidebar-item-label">Settings</span>
      </button>

      <span class="sidebar-spacer"></span>

      <button class="sidebar-item" type="button">
        <Info :size="16" class="sidebar-item-icon" />
        <span class="sidebar-item-label">Help</span>
      </button>

      <div class="sidebar-divider"></div>

      <button class="sidebar-item" type="button" @click="toggleCollapse">
        <PanelLeftClose :size="16" class="sidebar-item-icon" />
        <span class="sidebar-item-label">Collapse sidebar</span>
      </button>
    </template>

    <template v-else>
      <span class="collapsed-avatar" aria-label="Group avatar">
        <Folders :size="20" />
      </span>

      <nav class="collapsed-nav">
        <button
          v-for="item in navItems"
          :key="item.label"
          class="icon-btn"
          type="button"
          :aria-label="item.label"
          @click="navigate(item.to)"
        >
          <component :is="item.icon" :size="16" />
        </button>
      </nav>

      <div class="sidebar-divider"></div>

      <button class="icon-btn" type="button" aria-label="Settings" @click="navigate(`/${groupName}/settings`)">
        <Cog :size="16" />
      </button>

      <span class="sidebar-spacer"></span>

      <button class="icon-btn" type="button" aria-label="Help">
        <Info :size="16" />
      </button>

      <div class="sidebar-divider"></div>

      <button class="icon-btn" type="button" aria-label="Expand sidebar" @click="toggleCollapse">
        <PanelLeftOpen :size="16" />
      </button>
    </template>
  </aside>
</template>

<script setup lang="ts">
import {
	Cog,
	Folders,
	GitMerge,
	History,
	Info,
	type LucideIcon,
	MessageSquare,
	PanelLeftClose,
	PanelLeftOpen,
} from "@lucide/vue";
import { computed } from "vue";
import { useRouter } from "vue-router";

const props = withDefaults(
	defineProps<{
		collapsed?: boolean;
		groupName?: string;
	}>(),
	{
		groupName: "company",
	},
);

const router = useRouter();

type NavItem = {
	label: string;
	icon: LucideIcon;
	to: string;
	badge?: string;
};

const groupNameLabel = computed(() => {
	return props.groupName.charAt(0).toUpperCase() + props.groupName.slice(1);
});

const navItems = computed<NavItem[]>(() => {
	const base = `/${props.groupName}`;
	return [
		{
			label: "Merge requests",
			icon: GitMerge,
			to: `${base}/merge_requests`,
			badge: "0",
		},
		{ label: "Commits", icon: History, to: `${base}/commits`, badge: "0" },
		{
			label: "Comments",
			icon: MessageSquare,
			to: `${base}/comments`,
			badge: "0",
		},
	];
});

function navigate(to: string): void {
	router.push(to);
}

const emit = defineEmits<{
	toggle: [];
}>();

function toggleCollapse(): void {
	emit("toggle");
}
</script>

<style scoped>
.group-sidebar {
  width: 256px;
  height: 100%;
  background: var(--card);
  border-right: 1px solid var(--border);
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  flex-shrink: 0;
  transition: width 0.2s ease-in-out, padding 0.2s ease-in-out;
}

.group-sidebar--collapsed {
  width: 56px;
  padding: 16px 8px;
}

.group-sidebar--collapsed .collapsed-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--muted);
  color: var(--muted-foreground);
}

/* ── Expanded layout ── */

.group-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sidebar-header {
  width: 100%;
  padding: 0 8px;
  display: flex;
  align-items: center;
  min-height: 24px;
  margin-bottom: 6px;
}

.sidebar-header-text {
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  font-weight: 600;
}

.sidebar-divider {
  width: 100%;
  height: 1px;
  background: var(--border);
  margin: 8px 0;
}

.nav-divider {
  margin: 12px 0 8px;
}

.sidebar-item {
  width: 100%;
  border-radius: 6px;
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--muted-foreground);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  text-align: left;
}

.sidebar-item:hover {
  background: rgba(20, 20, 20, 0.3);
}

.sidebar-item-icon {
  flex-shrink: 0;
  color: #6b7280;
}

.sidebar-item-label {
  flex: 1;
}

.sidebar-badge {
  min-width: 20px;
  height: 20px;
  border-radius: 999px;
  padding: 0 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--muted);
  border: 1px solid var(--border);
  color: var(--foreground);
  font-size: 11px;
}

.sidebar-spacer {
  flex: 1;
}

/* ── Collapsed layout ── */

.collapsed-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.icon-btn {
  width: 100%;
  height: 36px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--background);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
}

.icon-btn:hover {
  background: rgba(20, 20, 20, 0.3);
}
</style>
