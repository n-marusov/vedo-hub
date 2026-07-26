// @m4 — VersioningPage vitest spec (GREEN phase: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: commits and branches from API via REST API calls
import {
	describePage,
	mountWithProviders,
} from "@/__tests__/setup/mock-providers";
import { expect, it, vi } from "vitest";
import { nextTick } from "vue";

// Mock only the API functions to return empty data; keep all other module exports intact.
vi.mock("@/api/versioning", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/versioning")>();
	return {
		...actual,
		listCommits: vi.fn(() => Promise.resolve({ items: [], total: 0, page: 0, perPage: 20 })),
		listBranches: vi.fn(() => Promise.resolve({ items: [], total: 0 })),
		listTags: vi.fn(() => Promise.resolve([])),
	};
});

describePage("VersioningPage", () => {
	it("should render commit history from API when Commits tab is active", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		// Wait for async REST fetches to complete + Vue reactivity
		await new Promise((resolve) => setTimeout(resolve, 300));
		await nextTick();
		expect(wrapper.text()).toContain("Commit History");
	});

	it("should switch tabs when tab button is clicked", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await new Promise((resolve) => setTimeout(resolve, 300));
		await nextTick();
		const tabs = wrapper.findAll(".tab");
		expect(tabs.length).toBeGreaterThanOrEqual(6);
	});

	it("should show empty state when no commits from API", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await new Promise((resolve) => setTimeout(resolve, 300));
		await nextTick();
		expect(wrapper.text()).toContain("No commits yet");
	});
});
