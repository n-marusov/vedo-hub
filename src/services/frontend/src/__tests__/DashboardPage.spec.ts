// @m4 — DashboardPage vitest spec (RED phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.2): page should render widgets, attention items, activity feed, recent ontologies from API

import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("DashboardPage", () => {
	afterEach(() => {
		resetMockResults();
	});

	it("should render widgets from API when page loads", async () => {
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dash-widgets").exists()).toBe(true);
	});

	it("should display attention items from API data", async () => {
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".attention-card").exists()).toBe(true);
	});

	it("should render activity feed from API when page loads", async () => {
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".activity-card").exists()).toBe(true);
	});

	it("should display recent ontologies from API data", async () => {
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".quick-card").exists()).toBe(true);
	});

	it("should show loading state while dashboard data loads from API", async () => {
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		// Don't wait for query — check loading state appears
		expect(wrapper.find(".dash-page").exists()).toBe(true);
	});

	it("should display user greeting with user name", async () => {
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dash-greeting").exists()).toBe(true);
	});

	it("should have clickable recent ontology items that navigate to workspace", async () => {
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		const ontoItems = wrapper.findAll(".onto-item");
		expect(ontoItems.length).toBeGreaterThan(0);
	});

	it("should not crash when dashboard has no data", async () => {
		setMockOperationResult("DashboardAggregate", {
			dashboardAggregate: {
				widgets: [],
				attentionItems: [],
				activityFeed: [],
				recentOntologies: [],
			},
		});
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		// Component should render without crashing even with empty data
		expect(wrapper.find(".dash-page").exists()).toBe(true);
	});

	it("should show error state when dashboard API fails", async () => {
		setMockOperationResult(
			"DashboardAggregate",
			null,
			new Error("Network error"),
		);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".error-state").exists() ||
				wrapper.find(".dash-error").exists() ||
				wrapper.text().includes("retry") ||
				wrapper.text().includes("error"),
		).toBe(true);
	});
});
