<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism/RecentCommits — recent commits list with pill tabs, no card wrapper -->
<!-- Matches design/ui-kit.lib.pen orgRecentCommits (Organism/RecentCommits) -->
<template>
    <div class="recent-commits" role="region" :aria-label="'Recent commits'">
        <div class="rc-tabs">
            <Tab v-model="activeTab" :tabs="tabs" label="Commit filter" />
        </div>

        <div
            class="rc-timeline"
            role="tabpanel"
            :id="'rc-tabpanel-' + activeTab"
            :aria-label="'Commits filtered by ' + activeTab"
        >
            <CommitItem
                v-for="commit in filteredCommits"
                :key="commit.sha"
                :author="commit.author"
                :handle="commit.handle"
                :timestamp="commit.timestamp"
                :action="commit.action"
                :sha="commit.sha"
                :commit-message="commit.commitMessage"
                :more-info="commit.moreInfo"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Tab from '../ui-kit/Tab.vue'
import CommitItem from './CommitItem.vue'

// @hlv:sec [INPUT_VALIDATION] — tab filter IDs are static enum, no dynamic injection
const tabs = [
  { value: 'all', label: 'All' },
  { value: 'my', label: 'My projects' },
  { value: 'starred', label: 'Starred projects' }
]

const activeTab = ref('all')

interface CommitEntry {
  author: string
  handle: string
  avatarSrc?: string
  timestamp: string
  action: string
  sha: string
  commitMessage: string
  moreInfo?: string
}

// @hlv:sec [INPUT_VALIDATION] — mock data, replaced with Apollo query in production
const commits: CommitEntry[] = [
  {
    author: 'Nikolay Marusov',
    handle: '@nikomaru',
    timestamp: '39 minutes ago',
    action: 'pushed to branch main at VEDO / VEDO Core',
    sha: 'b178461f',
    commitMessage: 'refactor: Доработка дашборда',
    moreInfo: '... and 1 more commit. Compare 9af554d7...b178461f'
  },
  {
    author: 'Anna Petrova',
    handle: '@anna',
    timestamp: '2 hours ago',
    action: 'pushed to branch develop at VEDO / VEDO Core',
    sha: '7c3d5f1',
    commitMessage: 'feat: Add SPARQL endpoint validation',
    moreInfo: 'Compare 7c3d5f1...9af554d7'
  },
  {
    author: 'Ivan Sidorov',
    handle: '@ivan',
    timestamp: '5 hours ago',
    action: 'pushed to branch fix/validation at VEDO / VEDO Core',
    sha: 'd4e6f2a',
    commitMessage: 'fix: Correct OWL axiom serialization',
    moreInfo: 'Compare d4e6f2a...b178461f'
  }
]

const filteredCommits = computed(() => {
  // @hlv ACTIVE_TAB_FILTER
  // For MVP, show all commits regardless of tab filter
  return commits
})
</script>

<style scoped>
.recent-commits {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 0;
}

.rc-tabs {
    display: flex;
}

/* Timeline */
.rc-timeline {
    display: flex;
    flex-direction: column;
    gap: 0;
}
</style>
