import {
	buildGroupUrl,
	buildProjectUrl,
	normalizeDomain,
	slugify,
	transliterate,
} from "@/utils/slug";
// @m4 — Slug/URL generation unit tests
// Validates: REQ-FUN.ORG.group-crud, REQ-FUN.ORG.project-creation
// Tests: transliteration + URL generation for groups and projects
// (Latin and non-Latin names).
import { expect, it } from "vitest";

describe("slugify", () => {
	it("should slugify a Latin name", () => {
		expect(slugify("My Test Group")).toBe("my-test-group");
	});

	it("should transliterate and slugify a Cyrillic name", () => {
		expect(slugify("Моя группа")).toBe("moya-gruppa");
	});

	it("should transliterate and slugify a mixed name", () => {
		expect(slugify("Проект Omega")).toBe("proekt-omega");
	});

	it("should handle Cyrillic hard/soft signs and diphthongs", () => {
		expect(slugify("Щука Ёлка Юля Яхта Ь Ъ Ы")).toBe(
			"shchuka-elka-yulya-yakhta-y",
		);
	});

	it("should collapse whitespace and strip invalid characters", () => {
		expect(slugify("  Hello   World  !!! ")).toBe("hello-world");
	});

	it("should collapse consecutive hyphens and trim edge hyphens", () => {
		expect(slugify("--foo--bar--")).toBe("foo-bar");
	});

	it("should respect maxLength", () => {
		expect(slugify("a".repeat(100), 10)).toBe("a".repeat(10));
	});
});

describe("transliterate", () => {
	it("should lowercase and transliterate Cyrillic", () => {
		expect(transliterate("Привет Мир")).toBe("privet mir");
	});

	it("should pass Latin and digits through", () => {
		expect(transliterate("Hello 123")).toBe("hello 123");
	});
});

describe("normalizeDomain", () => {
	it("should strip trailing slashes", () => {
		expect(normalizeDomain("https://vedo-core.local/")).toBe(
			"https://vedo-core.local",
		);
		expect(normalizeDomain("vedo-core.local///")).toBe("vedo-core.local");
	});
});

describe("buildGroupUrl", () => {
	it("should build a group URL from a Latin name", () => {
		expect(buildGroupUrl("vedo-core.local", "Research Team")).toBe(
			"vedo-core.local/research-team",
		);
	});

	it("should build a group URL from a Cyrillic name (transliterated)", () => {
		expect(buildGroupUrl("vedo-core.local", "Научная группа")).toBe(
			"vedo-core.local/nauchnaya-gruppa",
		);
	});
});

describe("buildProjectUrl", () => {
	it("should build a project URL from a Latin name", () => {
		expect(
			buildProjectUrl("vedo-core.local", "research-team", "Product Ontology"),
		).toBe("vedo-core.local/research-team/product-ontology");
	});

	it("should build a project URL from a Cyrillic name (transliterated)", () => {
		expect(
			buildProjectUrl("vedo-core.local", "research-team", "Онтология продукта"),
		).toBe("vedo-core.local/research-team/ontologiya-produkta");
	});

	it("should transliterate a Cyrillic group slug in the project URL", () => {
		expect(
			buildProjectUrl("vedo-core.local", "Научная группа", "Онтология"),
		).toBe("vedo-core.local/nauchnaya-gruppa/ontologiya");
	});
});
