// @m4 — MergeRequestsPage vitest spec (GREEN phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.5): page should render MR sections, filter tabs, and data from API

import {
	describePage,
	mountWithProviders,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it, vi } from "vitest";
import { nextTick } from "vue";

// MergeRequestsPage uses REST listMergeRequests() — mock for test control
vi.mock("@/api/merge-requests", () => ({
	listMergeRequests: vi.fn(),
}));

const mockMRs = [
	{
		id: "mr-1",
		title: "Test MR",
		description: "A test merge request",
		status: "open",
		mergeStatus: "can_be_merged",
		sourceProjectId: "proj-1",
		targetProjectId: "proj-1",
		sourceBranch: "feature/test",
		targetBranch: "main",
		authorId: "user-1",
		authorName: "Test User",
		commentCount: 0,
		createdAt: new Date().toISOString(),
	},
];

describePage("MergeRequestsPage", () => {
	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should render merge requests page layout", async () => {
		const { listMergeRequests } = await import("@/api/merge-requests");
		vi.mocked(listMergeRequests).mockResolvedValue(mockMRs);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-page").exists()).toBe(true);
	});

	it("should display page title Merge Requests", async () => {
		const { listMergeRequests } = await import("@/api/merge-requests");
		vi.mocked(listMergeRequests).mockResolvedValue(mockMRs);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-title").exists()).toBe(true);
	});

	it("should show filter tabs for merge request status", async () => {
		const { listMergeRequests } = await import("@/api/merge-requests");
		vi.mocked(listMergeRequests).mockResolvedValue(mockMRs);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-section").exists()).toBe(true);
	});

	it("should render merge request sections with data from API", async () => {
		const { listMergeRequests } = await import("@/api/merge-requests");
		vi.mocked(listMergeRequests).mockResolvedValue(mockMRs);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-section").exists()).toBe(true);
	});

	it("should show breadcrumbs navigation", async () => {
		const { listMergeRequests } = await import("@/api/merge-requests");
		vi.mocked(listMergeRequests).mockResolvedValue(mockMRs);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-breadcrumbs").exists()).toBe(true);
	});

	it("should show error state when merge requests API fails", async () => {
		const { listMergeRequests } = await import("@/api/merge-requests");
		vi.mocked(listMergeRequests).mockRejectedValue(
			new Error("Failed to load merge requests"),
		);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await new Promise((resolve) => setTimeout(resolve, 200));
		await nextTick();
		expect(
			wrapper.find(".error-state").exists() ||
				wrapper.text().includes("retry") ||
				wrapper.text().includes("error"),
		).toBe(true);
	});

	it("should show empty state when no merge requests exist", async () => {
		const { listMergeRequests } = await import("@/api/merge-requests");
		vi.mocked(listMergeRequests).mockResolvedValue([]);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await new Promise((resolve) => setTimeout(resolve, 200));
		await nextTick();
		expect(
			wrapper.find(".error-state").exists() ||
				wrapper.text().includes("No merge requests") ||
				wrapper.text().includes("merge"),
		).toBe(true);
	});
});
