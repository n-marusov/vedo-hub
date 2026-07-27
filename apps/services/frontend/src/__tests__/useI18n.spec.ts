// @m4 — useI18n vitest spec
// Validates: REQ-USR.UI.gui-implementation
// Tests: translation, fallback, interpolation, localStorage persistence
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// Mock locale files
vi.mock("@/locales/en.json", () => ({
	default: {
		"test.key": "Hello World",
		"test.param": "Hello {name}",
		"test.only_en": "English only",
		"toast.close": "Close",
	},
}));

vi.mock("@/locales/ru.json", () => ({
	default: {
		"test.key": "Привет мир",
		"test.param": "Привет {name}",
		"toast.close": "Закрыть",
	},
}));

describe("useI18n", () => {
	beforeEach(() => {
		localStorage.clear();
	});

	afterEach(() => {
		vi.clearAllMocks();
		localStorage.clear();
	});

	it("should return t function", async () => {
		const { useI18n } = await import("@/composables/useI18n");
		const { t } = useI18n();
		expect(typeof t).toBe("function");
	});

	it("should translate known key", async () => {
		const { useI18n } = await import("@/composables/useI18n");
		const { setLocale, t } = useI18n();
		await setLocale("ru");
		await nextTick();
		expect(t("test.key")).toBe("Привет мир");
	});

	it("should return key when missing", async () => {
		const { useI18n } = await import("@/composables/useI18n");
		const { setLocale, t } = useI18n();
		await setLocale("ru");
		await nextTick();
		expect(t("nonexistent.key")).toBe("nonexistent.key");
	});

	it("should fallback to English when key missing in Russian", async () => {
		const { useI18n } = await import("@/composables/useI18n");
		const { setLocale, t, preloadEn } = useI18n();
		await preloadEn();
		await setLocale("ru");
		await nextTick();
		// test.only_en exists in en.json but not ru.json
		const result = t("test.only_en");
		expect(result).toBe("English only");
	});

	it("should interpolate {placeholder} params", async () => {
		const { useI18n } = await import("@/composables/useI18n");
		const { setLocale, t } = useI18n();
		await setLocale("ru");
		await nextTick();
		expect(t("test.param", { name: "Иван" })).toBe("Привет Иван");
	});

	it("should persist locale selection to localStorage", async () => {
		const { useI18n } = await import("@/composables/useI18n");
		const { setLocale } = useI18n();
		await setLocale("en");
		await nextTick();
		expect(localStorage.getItem("vedo-locale")).toBe("en");
	});

	it("should toggle between ru and en", async () => {
		const { useI18n } = await import("@/composables/useI18n");
		const { setLocale, toggleLocale, locale } = useI18n();
		await setLocale("ru");
		await nextTick();
		expect(locale.value).toBe("ru");

		toggleLocale();
		await nextTick();
		expect(locale.value).toBe("en");
	});
});
