<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism Header component — top navigation bar with logo, search, user menu, theme toggle -->
<template>
  <header class="header" role="banner">
    <div class="header__left">
      <button class="header__menu-btn" :aria-label="t('nav.toggle_sidebar')" @click="toggleSidebar">
        ☰
      </button>
      <span class="header__logo">{{ t('app.title') }}</span>
    </div>

    <div class="header__center">
      <SearchInput
        :model-value="searchQuery"
        :placeholder="t('common.search')"
        label="Global search"
        @update:model-value="searchQuery = $event"
      />
    </div>

    <div class="header__right">
      <button
        class="header__theme-btn"
        :aria-label="t('theme.toggle')"
        @click="toggleTheme"
      >
        {{ themeMode === 'dark' ? '☀️' : '🌙' }}
      </button>

      <button
        class="header__locale-btn"
        :aria-label="locale === 'ru' ? 'Switch to English' : 'Переключить на русский'"
        @click="toggleLocale"
      >
        {{ locale === 'ru' ? 'RU' : 'EN' }}
      </button>

      <Avatar
        :src="user?.avatar_url"
        :display-name="user?.display_name || 'User'"
        size="sm"
      />
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '../../composables/useI18n'
import { useNavigationState } from '../../composables/useNavigationState'
import { applyThemeMode, detectSystemPreference } from '../../theme/manager'
import type { ThemeMode } from '../../theme/types'
import Avatar from '../ui-kit/Avatar.vue'
import SearchInput from '../ui-kit/SearchInput.vue'

const { toggleSidebar } = useNavigationState()
const { locale, toggleLocale, t } = useI18n()

const searchQuery = ref('')
const themeMode = ref<ThemeMode>(detectSystemPreference())

function toggleTheme() {
  themeMode.value = themeMode.value === 'light' ? 'dark' : 'light'
  applyThemeMode(themeMode.value)
}

interface User {
  avatar_url?: string
  display_name?: string
}

defineProps<{
  user?: User
}>()

defineEmits<{
  search: [query: string]
}>()
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--header-height);
  padding: 0 var(--spacing-6);
  background-color: var(--surface-primary);
  border-bottom: 1px solid var(--border-default);
  z-index: var(--z-sticky);
}

.header__left, .header__right {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
}

.header__center {
  flex: 1;
  max-width: 400px;
  margin: 0 var(--spacing-4);
}

.header__menu-btn {
  font-size: var(--font-size-xl);
  color: var(--text-secondary);
  cursor: pointer;
}

.header__logo {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--primary);
}

.header__theme-btn, .header__locale-btn {
  padding: var(--spacing-1) var(--spacing-2);
  font-size: var(--font-size-sm);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  cursor: pointer;
}

.header__theme-btn:hover, .header__locale-btn:hover {
  background-color: var(--surface-secondary);
}
</style>
