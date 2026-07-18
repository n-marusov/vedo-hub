import { describe, expect, it } from "vitest";

/**
 * Smoke test verifying the Vitest setup is correctly wired.
 *
 * This file exists so `npm test` succeeds on a fresh checkout before any
 * component tests are written. It validates that the Vitest configuration,
 * test environment (jsdom), and TypeScript transpilation pipeline all work.
 */
describe("vitest setup", () => {
	it("runs a basic assertion", () => {
		expect(1 + 1).toBe(2);
	});

	it("supports TypeScript syntax", () => {
		const value: string = "vedo-core";
		expect(value).toContain("vedo");
	});

	it("provides jsdom globals", () => {
		expect(document).toBeDefined();
		expect(window).toBeDefined();
	});
});
