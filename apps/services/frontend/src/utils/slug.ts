// @ctx: Slug generation — Cyrillic→Latin transliteration and URL building
// for groups and projects. Shared by CreateGroupPage / CreateProjectPage
// so unit tests cover the same code paths the forms use.

const CYRILLIC_TO_LATIN: Record<string, string> = {
	а: "a",
	б: "b",
	в: "v",
	г: "g",
	д: "d",
	е: "e",
	ё: "e",
	ж: "zh",
	з: "z",
	и: "i",
	й: "y",
	к: "k",
	л: "l",
	м: "m",
	н: "n",
	о: "o",
	п: "p",
	р: "r",
	с: "s",
	т: "t",
	у: "u",
	ф: "f",
	х: "kh",
	ц: "ts",
	ч: "ch",
	ш: "sh",
	щ: "shch",
	ъ: "",
	ы: "y",
	ь: "",
	э: "e",
	ю: "yu",
	я: "ya",
};

/**
 * Transliterates Cyrillic characters to Latin (lowercased). Non-Cyrillic
 * characters are passed through unchanged so mixed text keeps its Latin part.
 */
export function transliterate(input: string): string {
	return input
		.toLowerCase()
		.split("")
		.map((ch) => CYRILLIC_TO_LATIN[ch] ?? ch)
		.join("");
}

/**
 * Converts an arbitrary name into a URL-safe slug:
 * transliterate → whitespace to hyphens → strip invalid chars → collapse
 * consecutive hyphens → trim edge hyphens → cap length.
 */
export function slugify(input: string, maxLength = 64): string {
	return transliterate(input)
		.replace(/\s+/g, "-")
		.replace(/[^a-z0-9-]/g, "")
		.replace(/-{2,}/g, "-")
		.replace(/^-+|-+$/g, "")
		.substring(0, maxLength);
}

/**
 * Normalizes a runtime-provided domain (e.g. "https://vedo-core.local/"
 * or "vedo-core.local") to a trailing-slash-free base.
 */
export function normalizeDomain(domain: string): string {
	return domain.trim().replace(/\/+$/, "");
}

/**
 * Full group URL: {domain}/{groupSlug}.
 */
export function buildGroupUrl(domain: string, nameOrSlug: string): string {
	return `${normalizeDomain(domain)}/${slugify(nameOrSlug)}`;
}

/**
 * Full project URL: {domain}/{groupSlug}/{projectSlug}.
 */
export function buildProjectUrl(
	domain: string,
	groupSlug: string,
	projectNameOrSlug: string,
): string {
	return `${normalizeDomain(domain)}/${slugify(groupSlug)}/${slugify(projectNameOrSlug)}`;
}
