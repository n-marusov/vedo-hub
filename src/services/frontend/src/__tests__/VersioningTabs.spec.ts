// @m4 — VersioningTabs vitest spec (GREEN phase: uses mountWithProviders)
// Tests: Tags, Graph, and Compare tabs using Apollo queries
import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("VersioningTabs", () => {
	afterEach(() => {
		resetMockResults();
	});

	it("should show Tags tab with tag data label rendered", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Tags");
	});

	it("should show Repository Graph tab rendered", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Repository Graph");
	});

	it("should show Compare Revisions tab rendered", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.text()).toContain("Compare Revisions");
	});

	it("should render all 6 versioning tab buttons", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		const tabs = wrapper.findAll(".tab");
		expect(tabs.length).toBeGreaterThanOrEqual(6);
	});

	it("should display active Commits tab as highlighted by default", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		const activeTab = wrapper.find(".tab--active");
		expect(activeTab.exists()).toBe(true);
		expect(activeTab.text()).toBe("Commits");
	});

	it("should display tabs container after loading", async () => {
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".tab--active").exists()).toBe(true);
	});

	it("should not crash when versioning API fails", async () => {
		setMockOperationResult(
			"GetCommitHistory",
			null,
			new Error("Failed to load commits"),
		);
		const VersioningPage = (await import("@/pages/VersioningPage.vue")).default;
		const wrapper = mountWithProviders(VersioningPage);
		await waitForQuery();
		await nextTick();
		// Component should render without crashing on API failure
		expect(
			wrapper.find(".tab").exists() || wrapper.find(".vng-page").exists(),
		).toBe(true);
	});
});
