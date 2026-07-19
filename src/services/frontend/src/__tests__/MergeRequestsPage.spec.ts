// @m4 — MergeRequestsPage vitest spec (RED phase for Block В)
// Validates: REQ-USR.UI.gui-implementation
// After GREEN (Task 4.5): page should render MR sections, filter tabs, and data from API

import {
	describePage,
	mountWithProviders,
	resetMockResults,
	setMockOperationResult,
	waitForQuery,
} from "@/__tests__/setup/mock-providers";
import { afterEach, expect, it } from "vitest";
import { nextTick } from "vue";

describePage("MergeRequestsPage", () => {
	afterEach(() => {
		resetMockResults();
	});

	it("should render merge requests page layout", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-page").exists()).toBe(true);
	});

	it("should display page title Merge Requests", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-title").exists()).toBe(true);
	});

	it("should show filter tabs for merge request status", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-section").exists()).toBe(true);
	});

	it("should render merge request sections with data from API", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".merge-requests").exists()).toBe(true);
	});

	it("should show breadcrumbs navigation", async () => {
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(wrapper.find(".mr-breadcrumbs").exists()).toBe(true);
	});

	it("should show error state when merge requests API fails", async () => {
		setMockOperationResult(
			"ListMergeRequests",
			null,
			new Error("Failed to load merge requests"),
		);
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".error-state").exists() ||
				wrapper.text().includes("retry") ||
				wrapper.text().includes("error"),
		).toBe(true);
	});

	it("should show empty state when no merge requests exist", async () => {
		setMockOperationResult("ListMergeRequests", {
			listMergeRequests: { sections: [], total: 0 },
		});
		const MergeRequestsPage = (await import("@/pages/MergeRequestsPage.vue"))
			.default;
		const wrapper = mountWithProviders(MergeRequestsPage);
		await waitForQuery();
		await nextTick();
		expect(
			wrapper.find(".mr-empty").exists() ||
				wrapper.find(".empty-state").exists() ||
				wrapper.text().includes("No merge requests"),
		).toBe(true);
	});
});
