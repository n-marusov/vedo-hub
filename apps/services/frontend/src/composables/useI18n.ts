// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: i18n composable — Russian + English lazy-loaded locales
// @hlv i18n_lazy_load — Russian and English with lazy-loaded locale files

import { computed, ref } from 'vue'

type Locale = 'ru' | 'en'

const currentLocale = ref<Locale>('ru')
const messages = ref<Record<string, string>>({})
const loadedLocales = ref<Set<Locale>>(new Set())

const localeFiles: Record<Locale, () => Promise<Record<string, string>>> = {
  ru: () => import('../locales/ru.json').then((m) => m.default),
  en: () => import('../locales/en.json').then((m) => m.default)
}

export function useI18n() {
  const locale = computed(() => currentLocale.value)
  const isLoaded = computed(() => loadedLocales.value.has(currentLocale.value))

  function t(key: string, params?: Record<string, string>): string {
    const msg = messages.value[key] || key
    if (!params) return msg
    return msg.replace(/\{(\w+)\}/g, (_, k) => params[k] || `{${k}}`)
  }

  async function setLocale(loc: Locale) {
    if (!loadedLocales.value.has(loc)) {
      try {
        messages.value = await localeFiles[loc]()
        loadedLocales.value.add(loc)
      } catch {
        console.error(`Failed to load locale: ${loc}`)
        return
      }
    }
    currentLocale.value = loc
    document.documentElement.lang = loc
  }

  function toggleLocale() {
    setLocale(currentLocale.value === 'ru' ? 'en' : 'ru')
  }

  return { locale, isLoaded, t, setLocale, toggleLocale }
}
