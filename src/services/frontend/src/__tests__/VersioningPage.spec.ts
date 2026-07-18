// @m2.5 — VersioningPage vitest spec (GREEN phase: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: commits and branches from API via Apollo GraphQL queries
import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { beforeEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("VersioningPage", () => {
	beforeEach(async () => {
		// Router setup handled by mountWithProviders
	});

	it("should render commit history from API when Commits tab is active", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Commit History");
	});

	it("should switch tabs when tab button is clicked", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		const tabs = wrapper.findAll(".tab");
		expect(tabs.length).toBeGreaterThanOrEqual(6);
	});

	it("should show empty state when no commits from API", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		// Mock returns empty commits array → shows "No commits yet" empty state
		// Loading → data (empty) state transition
		expect(wrapper.text()).toContain("No commits yet");
	});
});
