// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: i18n composable — Russian + English lazy-loaded locales
// @hlv i18n_lazy_load — Russian and English with lazy-loaded locale files

import { computed, ref } from "vue";

export type Locale = "ru" | "en";

const STORAGE_KEY = "vedo-locale";
const DEFAULT_LOCALE: Locale = "ru";

function readPersistedLocale(): Locale {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored === "ru" || stored === "en") return stored;
	} catch {
		// localStorage may be unavailable (SSR, privacy mode)
	}
	return DEFAULT_LOCALE;
}

const currentLocale = ref<Locale>(readPersistedLocale());
const messages = ref<Record<string, string>>({});
const loadedLocales = ref<Set<Locale>>(new Set());

// Pre-load English at init so fallback is always available
let enMessages: Record<string, string> = {};

const localeFiles: Record<Locale, () => Promise<Record<string, string>>> = {
	ru: () => import("../locales/ru.json").then((m) => m.default),
	en: () => import("../locales/en.json").then((m) => m.default),
};

export function useI18n() {
	const locale = computed(() => currentLocale.value);
	const isLoaded = computed(() => loadedLocales.value.has(currentLocale.value));

	function t(key: string, params?: Record<string, string>): string {
		// Try current locale first, then fallback to English, then raw key
		const msg = messages.value[key] || enMessages[key] || key;
		if (!params) return msg;
		return msg.replace(/\{(\w+)\}/g, (_, k) => params[k] || `{${k}}`);
	}

	async function setLocale(loc: Locale) {
		if (!loadedLocales.value.has(loc)) {
			try {
				messages.value = await localeFiles[loc]();
				loadedLocales.value.add(loc);
			} catch {
				console.error(
					JSON.stringify({
						event: "i18n.load_error",
						locale: loc,
					}),
				);
				return;
			}
		}

		currentLocale.value = loc;
		document.documentElement.lang = loc;

		try {
			localStorage.setItem(STORAGE_KEY, loc);
		} catch {
			// localStorage may be unavailable
		}

		console.info(
			JSON.stringify({
				event: "i18n.locale_changed",
				locale: loc,
			}),
		);
	}

	function toggleLocale() {
		const next: Locale = currentLocale.value === "ru" ? "en" : "ru";
		setLocale(next);
	}

	// Preload English at init for fallback support
	async function preloadEn() {
		try {
			enMessages = await localeFiles.en();
		} catch {
			console.error(
				JSON.stringify({
					event: "i18n.en_preload_error",
				}),
			);
		}
	}

	return { locale, isLoaded, t, setLocale, toggleLocale, preloadEn };
}
