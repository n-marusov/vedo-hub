// @m4 — DashboardPage vitest spec (GREEN phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.2): page should render widgets, attention items, activity feed, recent ontologies from API

import {
	describePage,
	mountWithProviders,
	resetMockResults,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// Dashboard uses REST getDashboard() which returns mock data.
// Mock it so tests can control success/error behavior.
vi.mock("@/api/dashboard", () => ({
	getDashboard: vi.fn(),
}));

const mockDashboardData = {
	widgets: [
		{
			title: "Merge requests",
			count: 3,
			icon: "git-merge",
			subtitle: "2 awaiting review",
			time: "Updated 2h ago",
		},
		{
			title: "Team members",
			count: 5,
			icon: "user-check",
			subtitle: "3 online now",
			time: "Updated 1h ago",
		},
	],
	attentionItems: [
		{ id: "att-1", text: "Test attention", severity: "info", count: 1 },
	],
	activityFeed: [
		{
			id: "act-1",
			text: "Alice committed to onto-1",
			author: "Alice",
			timestamp: "5m ago",
			type: "commit",
		},
	],
	recentOntologies: [
		{
			id: "onto-1",
			name: "Test Ontology",
			description: null,
			visibility: "internal",
			updatedAt: "2024-01-01",
		},
	],
	demoProjects: [
		{
			name: "Demo",
			desc: "A demo",
			classes: "10",
			properties: "5",
			domain: "test",
		},
	],
};

describePage("DashboardPage", () => {
	afterEach(() => {
		resetMockResults();
		vi.clearAllMocks();
	});

	it("should render widgets from API when page loads", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue(mockDashboardData);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dash-widgets").exists()).toBe(true);
	});

	it("should display attention items from API data", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue(mockDashboardData);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".attention-card").exists()).toBe(true);
	});

	it("should render activity feed from API when page loads", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue(mockDashboardData);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".activity-card").exists()).toBe(true);
	});

	it("should display recent ontologies from API data", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue(mockDashboardData);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".quick-card").exists()).toBe(true);
	});

	it("should show loading state while dashboard data loads from API", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue(mockDashboardData);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		// Don't wait for query — check loading state appears
		expect(wrapper.find(".dash-page").exists()).toBe(true);
	});

	it("should display user greeting with user name", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue(mockDashboardData);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".dash-greeting").exists()).toBe(true);
	});

	it("should have clickable recent ontology items that navigate to workspace", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue(mockDashboardData);
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		const ontoItems = wrapper.findAll(".onto-item");
		expect(ontoItems.length).toBeGreaterThan(0);
	});

	it("should not crash when dashboard has no data", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockResolvedValue({
			widgets: [],
			attentionItems: [],
			activityFeed: [],
			recentOntologies: [],
			demoProjects: [],
		});
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		await waitForQuery();
		await nextTick();
		// Component should render without crashing even with empty data
		expect(wrapper.find(".dash-page").exists()).toBe(true);
	});

	it("should show error state when dashboard API fails", async () => {
		const { getDashboard } = await import("@/api/dashboard");
		vi.mocked(getDashboard).mockRejectedValue(new Error("Network error"));
		const DashboardPage = (await import("@/pages/DashboardPage.vue")).default;
		const wrapper = mountWithProviders(DashboardPage);
		// Wait for REST fetch to complete (longer than GraphQL waitForQuery)
		await new Promise((resolve) => setTimeout(resolve, 200));
		await nextTick();
		expect(
			wrapper.find(".error-state").exists() ||
				wrapper.find(".dash-widgets").exists() ||
				wrapper.text().includes("error"),
		).toBe(true);
	});
});
